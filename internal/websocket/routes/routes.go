package routes

import (
	"log"
	"net/http"
	"time"

	"skykin-platform/internal/common/websocket"

	"github.com/gin-gonic/gin"
	gorilla "github.com/gorilla/websocket"
)

const (
	pingInterval = 30 * time.Second
	pongTimeout  = 40 * time.Second
)

var upgrader = gorilla.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

func RegisterRoutes(r *gin.RouterGroup, hub *websocket.Hub) {
	r.GET("/ws/rewards/:user_id", func(c *gin.Context) {
		userID := c.Param("user_id")
		conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
		if err != nil {
			log.Printf("[WS] upgrade failed for user %s: %v", userID, err)
			return
		}

		conn.SetReadDeadline(time.Now().Add(pongTimeout))
		conn.SetPongHandler(func(string) error {
			conn.SetReadDeadline(time.Now().Add(pongTimeout))
			return nil
		})

		hub.ConnectUser(userID, conn)
		defer hub.DisconnectUser(userID)

		// Ping ticker keeps the connection alive through proxies and load balancers
		ticker := time.NewTicker(pingInterval)
		defer ticker.Stop()

		go func() {
			for range ticker.C {
				if err := conn.WriteControl(gorilla.PingMessage, nil, time.Now().Add(5*time.Second)); err != nil {
					return
				}
			}
		}()

		for {
			if _, _, err := conn.ReadMessage(); err != nil {
				break
			}
		}
	})
}
