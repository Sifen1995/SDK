package websocket

import (
	"encoding/json"
	"log"
	"sync"

	"github.com/gorilla/websocket"
)

type Notifier interface {
	ConnectUser(userID string, conn *websocket.Conn)
	DisconnectUser(userID string)
	NotifyUser(userID string, payload interface{}) error
}

type Hub struct {
	connections map[string]*websocket.Conn
	mu          sync.RWMutex
}

func NewHub() *Hub {
	return &Hub{
		connections: make(map[string]*websocket.Conn),
	}
}

func (h *Hub) ConnectUser(userID string, conn *websocket.Conn) {
	h.mu.Lock()
	defer h.mu.Unlock()

	if old, exists := h.connections[userID]; exists {
		old.Close()
	}
	h.connections[userID] = conn
	log.Printf("[WS] user %s connected (%d active)", userID, len(h.connections))
}

func (h *Hub) DisconnectUser(userID string) {
	h.mu.Lock()
	defer h.mu.Unlock()
	if conn, exists := h.connections[userID]; exists {
		conn.Close()
		delete(h.connections, userID)
		log.Printf("[WS] user %s disconnected (%d active)", userID, len(h.connections))
	}
}

func (h *Hub) NotifyUser(userID string, payload interface{}) error {
	h.mu.RLock()
	conn, exists := h.connections[userID]
	h.mu.RUnlock()

	if !exists {
		log.Printf("[WS] user %s is offline, notification skipped", userID)
		return nil
	}

	message, err := json.Marshal(payload)
	if err != nil {
		log.Printf("[WS] failed to marshal notification for user %s: %v", userID, err)
		return err
	}

	if err := conn.WriteMessage(websocket.TextMessage, message); err != nil {
		log.Printf("[WS] failed to send notification to user %s: %v", userID, err)
		h.DisconnectUser(userID)
		return err
	}

	log.Printf("[WS] notification sent to user %s (%d bytes)", userID, len(message))
	return nil
}
