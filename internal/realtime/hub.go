package realtime

import (
	"context"
	"notification-service/internal/notification"
	"sync"
)

type Hub struct {
	mu sync.RWMutex

	clients map[string]map[*Client]struct{}
}

func NewHub() *Hub {
	return &Hub{
		clients: make(map[string]map[*Client]struct{}),
	}
}

func (h *Hub) Register(client *Client) {
	h.mu.Lock()
	defer h.mu.Unlock()

	if _, exists := h.clients[client.UserID]; !exists {
		h.clients[client.UserID] = make(map[*Client]struct{})
	}

	h.clients[client.UserID][client] = struct{}{}
}
func (h *Hub) Unregister(client *Client) {
	h.mu.Lock()
	defer h.mu.Unlock()

	userClients, exists := h.clients[client.UserID]
	if !exists {
		return
	}

	delete(userClients, client)

	if len(userClients) == 0 {
		delete(h.clients, client.UserID)
	}
}
func (h *Hub) SendToUser(
	userID string,
	data any,
) {
	h.mu.RLock()

	userClients := make([]*Client, 0)

	for client := range h.clients[userID] {
		userClients = append(userClients, client)
	}

	h.mu.RUnlock()

	for _, client := range userClients {
		if err := client.WriteJSON(data); err != nil {
			h.Unregister(client)
			_ = client.Close()
		}
	}
}
func (h *Hub) Notify(
	ctx context.Context,
	n notification.Notification,
) error {
	event := map[string]any{
		"type": "notification.created",

		"data": map[string]any{
			"id": n.ID.Hex(),

			"type": n.Type,

			"title": n.Title,
			"body":  n.Body,

			"actor": n.Actor,

			"resource": n.Resource,

			"action": n.Action,

			"created_at": n.CreatedAt,
		},
	}

	h.SendToUser(
		n.RecipientID,
		event,
	)

	return nil
}
