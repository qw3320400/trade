package binance

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
	"trade/src/common"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

type FuturesAPIWebSocketRequest struct {
	ID     string          `json:"id"`
	Method string          `json:"method"`
	Params json.RawMessage `json:"params"`
}

type FuturesAPIWebSocketResponse struct {
	ID     string          `json:"id"`
	Status int64           `json:"status"`
	Result json.RawMessage `json:"result"`
}

func (c *Client) InitFuturesAPIWebSocketConnection(ctx context.Context) error {
	go c.ReadFuturesAPIWebSocketMessages(ctx)
	return c.ConnectFuturesAPIWebSocket(ctx)
}

func (c *Client) ReadFuturesAPIWebSocketMessages(ctx context.Context) {
	defer common.HandlePanic()
	for {
		if c.futuresAPIWebSocketConn == nil {
			time.Sleep(100 * time.Millisecond)
			continue
		}
		_, message, err := c.futuresAPIWebSocketConn.ReadMessage()
		if err != nil {
			if c.closed {
				return
			}
			common.Logger.Sugar().Warnf("ReadFuturesAPIWebSocketMessages ReadMessage error: %v", err)
			c.ReconnectFuturesAPIWebSocket(ctx)
			continue
		}
		var response FuturesAPIWebSocketResponse
		err = json.Unmarshal(message, &response)
		if err != nil {
			common.Logger.Sugar().Warnf("ReadFuturesAPIWebSocketMessages Unmarshal error: %v %s", err, string(message))
			continue
		}
		if ch, ok := c.futuresAPIWebSocketResponses[response.ID]; ok {
			ch <- &response
		} else {
			common.Logger.Sugar().Warnf("ReadFuturesAPIWebSocketMessages No handler for message: %s", string(message))
		}
	}
}

func (c *Client) ConnectFuturesAPIWebSocket(ctx context.Context) error {
	var err error
	c.futuresAPIWebSocketConn, _, err = websocket.DefaultDialer.DialContext(ctx, FuturesAPIWebSocketBaseURL, nil)
	if err != nil {
		return err
	}
	return nil
}

func (c *Client) ReconnectFuturesAPIWebSocket(ctx context.Context) {
	for i := 0; i < 5; i++ {
		time.Sleep(time.Second)
		err := c.ConnectFuturesAPIWebSocket(ctx)
		if err != nil {
			common.Logger.Sugar().Warnf("ReconnectFuturesAPIWebSocket error %v", err)
			continue
		} else {
			common.Logger.Sugar().Infof("ReconnectFuturesAPIWebSocket success after %d attempts", i+1)
			break
		}
	}
}

func (c *Client) Call(request *FuturesAPIWebSocketRequest) (*FuturesAPIWebSocketResponse, error) {
	if c.futuresAPIWebSocketConn == nil {
		return nil, fmt.Errorf("Call futuresAPIWebSocketConn is nil")
	}
	if request == nil {
		return nil, fmt.Errorf("Call request is nil")
	}
	id := uuid.New().String()
	request.ID = id
	c.futuresAPIWebSocketResponses[id] = make(chan *FuturesAPIWebSocketResponse, 1)
	defer delete(c.futuresAPIWebSocketResponses, id)
	err := c.futuresAPIWebSocketConn.WriteJSON(request)
	if err != nil {
		return nil, err
	}
	select {
	case response := <-c.futuresAPIWebSocketResponses[id]:
		if response.Status != http.StatusOK {
			return nil, fmt.Errorf("Call error: %+v", response)
		}
		return response, nil
	case <-time.After(time.Second * 10):
		return nil, fmt.Errorf("Call timeout waiting for response")
	}
}

type FuturesAPIWebSocketPriceTicker struct {
	Symbol string `json:"symbol"`
	Price  string `json:"price"`
	Time   int64  `json:"time"`
}

func NewFuturesAPIWebSocketPriceTicker(symbol string) *FuturesAPIWebSocketPriceTicker {
	return &FuturesAPIWebSocketPriceTicker{
		Symbol: symbol,
	}
}

func (f *FuturesAPIWebSocketPriceTicker) Request() *FuturesAPIWebSocketRequest {
	params := map[string]string{
		"symbol": f.Symbol,
	}
	paramsBytes, _ := json.Marshal(params)
	return &FuturesAPIWebSocketRequest{
		Method: "ticker.price",
		Params: paramsBytes,
	}
}

func (f *FuturesAPIWebSocketPriceTicker) Response(resp *FuturesAPIWebSocketResponse) (*FuturesAPIWebSocketPriceTicker, error) {
	err := json.Unmarshal(resp.Result, f)
	if err != nil {
		return nil, err
	}
	return f, nil
}
