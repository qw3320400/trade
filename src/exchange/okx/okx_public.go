package okx

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

func (c *Client) InitPublicWebSocketConnection(ctx context.Context) error {
	go c.ReadPublicWebSocketMessages(ctx)
	return c.ConnectPublicWebSocket(ctx)
}

func (c *Client) ReadPublicWebSocketMessages(ctx context.Context) {
	defer common.HandlePanic()
	for {
		if c.publicWebSocketConn == nil {
			time.Sleep(100 * time.Millisecond)
			continue
		}
		_, message, err := c.publicWebSocketConn.ReadMessage()
		if err != nil {
			if c.closed {
				return
			}
			common.Logger.Sugar().Warnf("ReadPublicWebSocketMessages ReadMessage error: %v", err)
			c.ReconnectPublicWebSocket(ctx)
			continue
		}
		var stream WebSocketStream
		err = json.Unmarshal(message, &stream)
		if err != nil {
			common.Logger.Sugar().Warnf("ReadPublicWebSocketMessages Unmarshal error: %v %s", err, string(message))
			continue
		}
		if hand, ok := c.publicWebSocketHandlers[stream.Arg.Key()]; ok {
			hand(&stream)
		} else {
			common.Logger.Sugar().Warnf("ReadPublicWebSocketMessages No handler for message: %s", string(message))
		}
	}
}

func (c *Client) ConnectPublicWebSocket(ctx context.Context) error {
	var err error
	c.publicWebSocketConn, _, err = websocket.DefaultDialer.DialContext(ctx, PublicWebSocketBaseURL, nil)
	if err != nil {
		return err
	}
	return nil
}

func (c *Client) ReconnectPublicWebSocket(ctx context.Context) {
	for i := 0; i < 5; i++ {
		time.Sleep(time.Second)
		err := c.ConnectPublicWebSocket(ctx)
		if err != nil {
			common.Logger.Sugar().Warnf("ReconnectPublicWebSocket error %v", err)
			continue
		} else {
			common.Logger.Sugar().Infof("ReconnectPublicWebSocket success after %d attempts", i+1)
			break
		}
	}
}

func (c *Client) Subscribe(request *WebSocketRequest, handler func(*WebSocketStream)) error {
	if c.publicWebSocketConn == nil {
		return fmt.Errorf("SubscribePublicWebSocket publicWebSocketConn is nil")
	}
	if request == nil || handler == nil {
		return fmt.Errorf("SubscribePublicWebSocket request/handler is empty")
	}
	for _, arg := range request.Args {
		c.publicWebSocketHandlers[arg.Key()] = handler
	}
	id := strings.Replace(uuid.New().String(), "-", "", -1)
	request.ID = id
	err := c.publicWebSocketConn.WriteJSON(request)
	if err != nil {
		return fmt.Errorf("SubscribePublicWebSocket WriteJSON error: %v", err)
	}
	return nil
}

type PublicWebSocketMarkPrices []*PublicWebSocketMarkPrice

type PublicWebSocketMarkPrice struct {
	InstType  string          `json:"instType"`
	InstID    string          `json:"instId"`
	MarkPrice decimal.Decimal `json:"markPx"`
	Timestamp string          `json:"ts"`
}

func NewPublicWebSocketMarkPrices(instID string) *PublicWebSocketMarkPrices {
	return &PublicWebSocketMarkPrices{
		{
			InstID: instID,
		},
	}
}

func (p *PublicWebSocketMarkPrices) Subscribe() *WebSocketRequest {
	request := &WebSocketRequest{
		OP:   "subscribe",
		Args: []*WebSocketArg{},
	}
	for _, item := range *p {
		arg := &WebSocketArg{
			Channel: "mark-price",
			InstID:  item.InstID,
		}
		request.Args = append(request.Args, arg)
	}
	return request
}

func (p *PublicWebSocketMarkPrices) Stream(response *WebSocketStream) (*PublicWebSocketMarkPrices, error) {
	err := json.Unmarshal(response.Data, p)
	if err != nil {
		return nil, err
	}
	return p, nil
}
