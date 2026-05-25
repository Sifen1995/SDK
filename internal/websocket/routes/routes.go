package routes

import (
	"net/http"
	"skykin-platform/internal/common/websocket"

	"github.com/gin-gonic/gin"
	gorilla "github.com/gorilla/websocket"
)

var upgrader = gorilla.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true // Allow all origins for testing
	},
}

func RegisterRoutes(r *gin.RouterGroup, hub *websocket.Hub) {
	r.GET("/ws/rewards/:user_id", func(c *gin.Context) {
		userID := c.Param("user_id")
		conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
		if err != nil {
			return
		}
		hub.ConnectUser(userID, conn)

		// Keep connection alive until client disconnects
		defer hub.DisconnectUser(userID)
		for {
			if _, _, err := conn.ReadMessage(); err != nil {
				break
			}
		}
	})
}
