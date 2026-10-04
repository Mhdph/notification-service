package realtime

import (
	"sync"

	"github.com/gorilla/websocket"
)

type Client struct {
	AppID  string
	UserID string

	conn *websocket.Conn
	mu   sync.Mutex
}

func NewClient(
	appID string,
	userID string,
	conn *websocket.Conn,
) *Client {
	return &Client{
		AppID:  appID,
		UserID: userID,
		conn:   conn,
	}
}

func (c *Client) WriteJSON(data any) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	return c.conn.WriteJSON(data)
}

func (c *Client) Close() error {
	return c.conn.Close()
}
