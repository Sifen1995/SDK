package websocket

import (
	"encoding/json"
	"fmt"
	"sync"

	"github.com/gorilla/websocket"
)

// Notifier defines the interface for sending real-time messages
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

// ConnectUser adds a user to the active registry
func (h *Hub) ConnectUser(userID string, conn *websocket.Conn) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.connections[userID] = conn
	fmt.Printf("User %s connected to WebSocket\n", userID)
}

// DisconnectUser removes a user and closes the connection
func (h *Hub) DisconnectUser(userID string) {
	h.mu.Lock()
	defer h.mu.Unlock()
	if conn, exists := h.connections[userID]; exists {
		conn.Close()
		delete(h.connections, userID)
		fmt.Printf("User %s disconnected\n", userID)
	}
}

// NotifyUser pushes data to a specific user if they are online
func (h *Hub) NotifyUser(userID string, payload interface{}) error {
	h.mu.RLock()
	conn, exists := h.connections[userID]
	h.mu.RUnlock()

	if !exists {
		// User is offline, we don't need to do anything
		return nil
	}

	// Convert payload to JSON and send
	message, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	return conn.WriteMessage(websocket.TextMessage, message)
}
