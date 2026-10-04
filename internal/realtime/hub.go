package realtime

import (
	"fmt"
	"sync"
)

type Hub struct {
	mu sync.RWMutex

	clients map[string]map[*Client]struct{}
}

func NewHub() *Hub {
	return &Hub{
		clients: make(
			map[string]map[*Client]struct{},
		),
	}
}

func clientKey(
	appID string,
	userID string,
) string {
	return fmt.Sprintf(
		"%s:%s",
		appID,
		userID,
	)
}

func (h *Hub) Register(
	client *Client,
) {
	key := clientKey(
		client.AppID,
		client.UserID,
	)

	h.mu.Lock()
	defer h.mu.Unlock()

	if h.clients[key] == nil {
		h.clients[key] =
			make(
				map[*Client]struct{},
			)
	}

	h.clients[key][client] =
		struct{}{}
}

func (h *Hub) Unregister(
	client *Client,
) {
	key := clientKey(
		client.AppID,
		client.UserID,
	)

	h.mu.Lock()
	defer h.mu.Unlock()

	userClients, ok :=
		h.clients[key]

	if !ok {
		return
	}

	delete(
		userClients,
		client,
	)

	if len(userClients) == 0 {
		delete(
			h.clients,
			key,
		)
	}
}

func (h *Hub) SendToUser(
	appID string,
	userID string,
	data any,
) {
	key := clientKey(
		appID,
		userID,
	)

	h.mu.RLock()

	clientsMap :=
		h.clients[key]

	clients := make(
		[]*Client,
		0,
		len(clientsMap),
	)

	for client := range clientsMap {

		clients = append(
			clients,
			client,
		)
	}

	h.mu.RUnlock()

	for _, client := range clients {

		if client.Enqueue(data) {
			continue
		}

		h.Unregister(client)

		_ = client.Close()
	}
}
