package realtime

import (
	"net/http"

	"notification-service/internal/identity"

	"github.com/gorilla/websocket"
)

type Handler struct {
	hub *Hub
}

func NewHandler(
	hub *Hub,
) *Handler {
	return &Handler{
		hub: hub,
	}
}

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,

	CheckOrigin: func(
		r *http.Request,
	) bool {
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

	h.hub.Register(
		client,
	)

	go client.WritePump(
		h.hub,
	)

	client.ReadPump(
		h.hub,
	)
}
