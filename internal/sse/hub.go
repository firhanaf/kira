package sse

import (
	"encoding/json"
	"sync"

	"github.com/google/uuid"
)

type Event struct {
	Type    string `json:"type"`
	Payload any    `json:"payload,omitempty"`
}

type Hub struct {
	mu      sync.RWMutex
	clients map[uuid.UUID]map[chan Event]struct{}
}

func New() *Hub {
	return &Hub{
		clients: make(map[uuid.UUID]map[chan Event]struct{}),
	}
}

func (h *Hub) Subscribe(userID uuid.UUID) chan Event {
	ch := make(chan Event, 16)
	h.mu.Lock()
	if h.clients[userID] == nil {
		h.clients[userID] = make(map[chan Event]struct{})
	}
	h.clients[userID][ch] = struct{}{}
	h.mu.Unlock()
	return ch
}

func (h *Hub) Unsubscribe(userID uuid.UUID, ch chan Event) {
	h.mu.Lock()
	delete(h.clients[userID], ch)
	if len(h.clients[userID]) == 0 {
		delete(h.clients, userID)
	}
	h.mu.Unlock()
	close(ch)
}

func (h *Hub) Publish(userID uuid.UUID, event Event) {
	h.mu.RLock()
	defer h.mu.RUnlock()
	for ch := range h.clients[userID] {
		select {
		case ch <- event:
		default:
		}
	}
}

func (h *Hub) PublishJSON(userID uuid.UUID, eventType string, payload any) {
	h.Publish(userID, Event{Type: eventType, Payload: payload})
}

func MarshalEvent(e Event) ([]byte, error) {
	return json.Marshal(e)
}
