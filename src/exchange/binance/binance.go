package binance

import (
	"github.com/gorilla/websocket"
)

const (
	FuturesAPIWebSocketBaseURL    = "wss://ws-fapi.binance.com/ws-fapi/v1"
	FuturesStreamWebSocketBaseURL = "wss://fstream.binance.com/stream"
)

type Client struct {
	closed                         bool
	futuresAPIWebSocketConn        *websocket.Conn
	futuresAPIWebSocketResponses   map[string]chan *FuturesAPIWebSocketResponse
	futuresStreamWebSocketConn     *websocket.Conn
	futuresStreamWebSocketHandlers map[string]func(*FuturesStreamWebSocketStream)
}

func NewClient() *Client {
	return &Client{
		futuresAPIWebSocketResponses:   make(map[string]chan *FuturesAPIWebSocketResponse, 100),
		futuresStreamWebSocketHandlers: make(map[string]func(*FuturesStreamWebSocketStream), 100),
	}
}

func (c *Client) Clean() {
	c.closed = true
	if c.futuresAPIWebSocketConn != nil {
		c.futuresAPIWebSocketConn.Close()
	}
	if c.futuresStreamWebSocketConn != nil {
		c.futuresStreamWebSocketConn.Close()
	}
}
