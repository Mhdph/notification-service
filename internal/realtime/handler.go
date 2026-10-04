package realtime

import (
	"net/http"
	"notification-service/internal/identity"

	"github.com/gorilla/websocket"
)

type Handler struct {
	hub *Hub
}

func NewHandler(hub *Hub) *Handler {
	return &Handler{
		hub: hub,
	}
}

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

func (h *Handler) ServeWS(
	w http.ResponseWriter,
	r *http.Request,
) {
	currentIdentity, ok :=
		identity.FromContext(
			r.Context(),
		)

	if !ok {
		http.Error(
			w,
			"unauthorized",
			http.StatusUnauthorized,
		)
		return
	}

	conn, err :=
		upgrader.Upgrade(
			w,
			r,
			nil,
		)

	if err != nil {
		return
	}

	client :=
		NewClient(
			currentIdentity.AppID,
			currentIdentity.UserID,
			conn,
		)

	h.hub.Register(client)

	defer func() {
		h.hub.Unregister(client)
		_ = client.Close()
	}()

	for {
		if _, _, err :=
			conn.ReadMessage(); err != nil {
			break
		}
	}
}
