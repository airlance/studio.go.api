package ws

import (
	"context"
	"sync"

	"github.com/gorilla/websocket"
	"github.com/resoul/studio.go.api/internal/domain"
	"github.com/sirupsen/logrus"
)

type Client struct {
	conn *websocket.Conn
	send chan interface{}
}

func (c *Client) Send(event interface{}) error {
	c.send <- event
	return nil
}

func (c *Client) Close() error {
	return c.conn.Close()
}

type Hub struct {
	clients    map[string][]domain.PresenceClient // userID -> slice of connections
	register   chan registration
	unregister chan registration
	broadcast  chan interface{}
	mu         sync.RWMutex
}

type registration struct {
	userID string
	client domain.PresenceClient
}

func NewHub() *Hub {
	return &Hub{
		clients:    make(map[string][]domain.PresenceClient),
		register:   make(chan registration),
		unregister: make(chan registration),
		broadcast:  make(chan interface{}),
	}
}

func (h *Hub) Run(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		case r := <-h.register:
			h.mu.Lock()
			isFirst := len(h.clients[r.userID]) == 0
			h.clients[r.userID] = append(h.clients[r.userID], r.client)
			h.mu.Unlock()

			if isFirst {
				h.Broadcast(ctx, domain.PresenceEvent{
					Type:   domain.PresenceEventJoin,
					UserID: r.userID,
				})
			}

			// Send current online users to the new client
			r.client.Send(domain.WebsocketEvent{
				Type: domain.WebsocketEventPresence,
				Payload: domain.PresenceSyncEvent{
					Type:    domain.PresenceEventSync,
					UserIDs: h.GetOnlineUsers(),
				},
			})

		case r := <-h.unregister:
			h.mu.Lock()
			if clients, ok := h.clients[r.userID]; ok {
				for i, c := range clients {
					if c == r.client {
						h.clients[r.userID] = append(clients[:i], clients[i+1:]...)
						break
					}
				}
				if len(h.clients[r.userID]) == 0 {
					delete(h.clients, r.userID)
					h.mu.Unlock()
					h.Broadcast(ctx, domain.PresenceEvent{
						Type:   domain.PresenceEventLeave,
						UserID: r.userID,
					})
				} else {
					h.mu.Unlock()
				}
			} else {
				h.mu.Unlock()
			}
		}
	}
}

func (h *Hub) Register(ctx context.Context, userID string, client domain.PresenceClient) {
	h.register <- registration{userID: userID, client: client}
}

func (h *Hub) Unregister(ctx context.Context, userID string, client domain.PresenceClient) {
	h.unregister <- registration{userID: userID, client: client}
}

func (h *Hub) GetOnlineUsers() []string {
	h.mu.RLock()
	defer h.mu.RUnlock()
	users := make([]string, 0, len(h.clients))
	for userID := range h.clients {
		users = append(users, userID)
	}
	return users
}

func (h *Hub) Broadcast(ctx context.Context, event interface{}) {
	h.mu.RLock()
	defer h.mu.RUnlock()

	var finalEvent interface{} = event
	// Wrap if not already wrapped
	if _, ok := event.(domain.WebsocketEvent); !ok {
		eventType := domain.WebsocketEventPresence
		// Check if it's a message
		if _, ok := event.(*domain.Message); ok {
			eventType = domain.WebsocketEventChatMessage
		} else if _, ok := event.(domain.Message); ok {
			eventType = domain.WebsocketEventChatMessage
		}

		finalEvent = domain.WebsocketEvent{
			Type:    eventType,
			Payload: event,
		}
	}

	logrus.WithFields(logrus.Fields{
		"type": string(finalEvent.(domain.WebsocketEvent).Type),
	}).Info("Hub broadcasting event to clients")

	for _, clients := range h.clients {
		for _, client := range clients {
			if err := client.Send(finalEvent); err != nil {
				logrus.WithError(err).Warn("Failed to send event to client")
			}
		}
	}
}
