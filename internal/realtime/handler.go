package realtime

import (
	"net/http"

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
	userID := r.URL.Query().Get("user_id")

	if userID == "" {
		http.Error(
			w,
			"user_id is required",
			http.StatusBadRequest,
		)
		return
	}

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		return
	}

	client := NewClient(userID, conn)

	h.hub.Register(client)

	defer func() {
		h.hub.Unregister(client)
		_ = client.Close()
	}()

	for {
		if _, _, err := conn.ReadMessage(); err != nil {
			break
		}
	}
}
