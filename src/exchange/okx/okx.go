package okx

import (
	"encoding/json"

	"github.com/gorilla/websocket"
)

const (
	PublicWebSocketBaseURL  = "wss://ws.okx.com:8443/ws/v5/public"
	PrivateWebSocketBaseURL = "wss://ws.okx.com:8443/ws/v5/private"
)

type Client struct {
	closed                  bool
	publicWebSocketConn     *websocket.Conn
	publicWebSocketHandlers map[string]func(*WebSocketStream)
}

type WebSocketRequest struct {
	ID   string          `json:"id"`
	OP   string          `json:"op"`
	Args []*WebSocketArg `json:"args"`
}

type WebSocketStream struct {
	Arg  *WebSocketArg   `json:"arg"`
	Data json.RawMessage `json:"data"`
}

type WebSocketArg struct {
	Channel string `json:"channel"`
	InstID  string `json:"instId"`
}

func (a *WebSocketArg) Key() string {
	if a == nil {
		return ""
	}
	return a.Channel + "_" + a.InstID
}

func NewClient() *Client {
	return &Client{
		publicWebSocketHandlers: make(map[string]func(*WebSocketStream), 100),
	}
}

func (c *Client) Clean() {
	c.closed = true
	if c.publicWebSocketConn != nil {
		c.publicWebSocketConn.Close()
	}
}
