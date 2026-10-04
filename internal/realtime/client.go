package realtime

import (
	"time"

	"github.com/gorilla/websocket"
)

const (
	writeWait = 10 * time.Second

	pongWait = 60 * time.Second

	pingPeriod = 50 * time.Second

	maxMessageSize = 4096

	sendBufferSize = 64
)

type Client struct {
	AppID string

	UserID string

	conn *websocket.Conn

	send chan any
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

		send: make(
			chan any,
			sendBufferSize,
		),
	}
}

func (c *Client) Enqueue(
	data any,
) bool {
	select {
	case c.send <- data:
		return true

	default:
		return false
	}
}

func (c *Client) Close() error {
	return c.conn.Close()
}
func (c *Client) ReadPump(
	hub *Hub,
) {
	defer func() {
		hub.Unregister(c)
		_ = c.Close()
	}()

	c.conn.SetReadLimit(
		maxMessageSize,
	)

	_ = c.conn.SetReadDeadline(
		time.Now().Add(
			pongWait,
		),
	)

	c.conn.SetPongHandler(
		func(string) error {
			return c.conn.SetReadDeadline(
				time.Now().Add(
					pongWait,
				),
			)
		},
	)

	for {
		if _, _, err :=
			c.conn.ReadMessage(); err != nil {
			return
		}
	}
}
func (c *Client) WritePump(
	hub *Hub,
) {
	ticker :=
		time.NewTicker(
			pingPeriod,
		)

	defer func() {
		ticker.Stop()

		hub.Unregister(c)

		_ = c.Close()
	}()

	for {
		select {

		case data, ok :=
			<-c.send:

			_ = c.conn.SetWriteDeadline(
				time.Now().Add(
					writeWait,
				),
			)

			if !ok {
				_ = c.conn.WriteMessage(
					websocket.CloseMessage,
					[]byte{},
				)

				return
			}

			if err :=
				c.conn.WriteJSON(
					data,
				); err != nil {

				return
			}

		case <-ticker.C:

			_ = c.conn.SetWriteDeadline(
				time.Now().Add(
					writeWait,
				),
			)

			if err :=
				c.conn.WriteMessage(
					websocket.PingMessage,
					nil,
				); err != nil {

				return
			}
		}
	}
}
