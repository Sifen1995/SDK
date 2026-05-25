package websocket

// Manager is a placeholder for managing websocket lifecycles.
type Manager struct {
	Hub *Hub
}

func NewManager(h *Hub) *Manager {
	return &Manager{Hub: h}
}
