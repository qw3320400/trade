package binance

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"
	"trade/src/common"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"github.com/shopspring/decimal"
)

type FuturesStreamWebSocketRequest struct {
	ID     string   `json:"id"`
	Method string   `json:"method"`
	Params []string `json:"params"`
}

type FuturesStreamWebSocketStream struct {
	Stream string          `json:"stream"`
	Data   json.RawMessage `json:"data"`
}

func (c *Client) InitFuturesStreamWebSocketConnection(ctx context.Context) error {
	go c.ReadFuturesStreamWebSocketMessages(ctx)
	return c.ConnectFuturesStreamWebSocket(ctx)
}

func (c *Client) ReadFuturesStreamWebSocketMessages(ctx context.Context) {
	defer common.HandlePanic()
	for {
		if c.futuresStreamWebSocketConn == nil {
			time.Sleep(100 * time.Millisecond)
			continue
		}
		_, message, err := c.futuresStreamWebSocketConn.ReadMessage()
		if err != nil {
			if c.closed {
				return
			}
			common.Logger.Sugar().Warnf("ReadFuturesStreamWebSocketMessages ReadMessage error: %v", err)
			c.ReconnectFuturesStreamWebSocket(ctx)
			continue
		}
		var stream FuturesStreamWebSocketStream
		err = json.Unmarshal(message, &stream)
		if err != nil {
			common.Logger.Sugar().Warnf("ReadFuturesStreamWebSocketMessages Unmarshal error: %v %s", err, string(message))
			continue
		}
		if hand, ok := c.futuresStreamWebSocketHandlers[stream.Stream]; ok {
			hand(&stream)
		} else {
			common.Logger.Sugar().Warnf("ReadFuturesStreamWebSocketMessages No handler for message: %s", string(message))
		}
	}
}

func (c *Client) ConnectFuturesStreamWebSocket(ctx context.Context) error {
	var err error
	c.futuresStreamWebSocketConn, _, err = websocket.DefaultDialer.DialContext(ctx, FuturesStreamWebSocketBaseURL, nil)
	if err != nil {
		return err
	}
	return nil
}

func (c *Client) ReconnectFuturesStreamWebSocket(ctx context.Context) {
	for i := 0; i < 5; i++ {
		time.Sleep(time.Second)
		err := c.ConnectFuturesStreamWebSocket(ctx)
		if err != nil {
			common.Logger.Sugar().Warnf("ReconnectFuturesStreamWebSocket error %v", err)
			continue
		}
		common.Logger.Sugar().Infof("ReconnectFuturesStreamWebSocket success after %d attempts", i+1)
	}
}

func (c *Client) Subscribe(subscribe *FuturesStreamWebSocketRequest, handler func(*FuturesStreamWebSocketStream)) error {
	if c.futuresStreamWebSocketConn == nil {
		return fmt.Errorf("SubscribeFuturesStreamMarketPrice futuresStreamWebSocketConn is nil")
	}
	if subscribe == nil || handler == nil {
		return fmt.Errorf("SubscribeFuturesStreamMarketPrice stream/handler is empty")
	}
	for _, stream := range subscribe.Params {
		c.futuresStreamWebSocketHandlers[stream] = handler
	}
	id := uuid.New().String()
	subscribe.ID = id
	err := c.futuresStreamWebSocketConn.WriteJSON(subscribe)
	if err != nil {
		return fmt.Errorf("SubscribeFuturesStreamMarketPrice WriteJSON error: %v", err)
	}
	return nil
}

type FuturesStreamWebSocketMarketPrice struct {
	EventType            string          `json:"e"`
	EventTime            int64           `json:"E"`
	Symbol               string          `json:"s"`
	MarkPrice            decimal.Decimal `json:"p"`
	IndexPrice           decimal.Decimal `json:"i"`
	EstimatedSettlePrice decimal.Decimal `json:"P"`
	FundingRate          decimal.Decimal `json:"r"`
	NextFundingTime      int64           `json:"T"`
}

func NewFuturesStreamWebSocketMarketPrice(symbol string) *FuturesStreamWebSocketMarketPrice {
	return &FuturesStreamWebSocketMarketPrice{
		Symbol: symbol,
	}
}

func (f *FuturesStreamWebSocketMarketPrice) Subscribe() *FuturesStreamWebSocketRequest {
	stream := strings.ToLower(f.Symbol) + "@markPrice"
	return &FuturesStreamWebSocketRequest{
		Method: "SUBSCRIBE",
		Params: []string{stream},
	}
}

func (f *FuturesStreamWebSocketMarketPrice) Stream(stream *FuturesStreamWebSocketStream) (*FuturesStreamWebSocketMarketPrice, error) {
	err := json.Unmarshal(stream.Data, f)
	if err != nil {
		return nil, err
	}
	return f, nil
}
