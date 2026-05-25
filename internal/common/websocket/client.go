package websocket

import (
	"github.com/gorilla/websocket"
)

// Client represents a single websocket client connection.
type Client struct {
	Hub  *Hub
	Conn *websocket.Conn
	Send chan []byte
}
