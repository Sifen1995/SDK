package websocket

type Manager struct {
	Hub *Hub
}

func NewManager(h *Hub) *Manager {
	return &Manager{Hub: h}
}

