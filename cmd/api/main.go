package main

import (
	"fmt"
	"log"
	"net/http"
	"skykin-platform/configs" // Import your new central routes setup
	"skykin-platform/internal/common/database"
	"skykin-platform/internal/common/route" // Import your new central routes setup
	"skykin-platform/internal/common/websocket"

	"github.com/gin-gonic/gin"
)

func main() {
	// Initialize clean, bare engine instance
	r := gin.New()

	// Load configuration properties
	cfg, err := configs.LoadConfig()
	if err != nil {
		log.Fatalf("failed to load system configuration: %v", err)
	}

	// Establish connection pool to the Postgres container
	db, err := database.ConnectDB(cfg)
	if err != nil {
		log.Fatalf("failed to open database target connection: %v", err)
	}

	// Initialize the structural communication hub
	hub := websocket.NewHub()

	// Global Health check line
	r.GET("/ping", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok", "message": "Ready to build!"})
	})

	// Hand off the engine instance to isolate structural setups cleanly
	route.InitRouter(r, db, cfg, hub)

	// Fire up the HTTP engine instance
	serverAddress := fmt.Sprintf("0.0.0.0:%s", cfg.Port)
	log.Printf("Listening on %s — Boot sequence successful.", serverAddress)
	if err := r.Run(serverAddress); err != nil {
		log.Fatalf("critical failure running engine server: %v", err)
	}
}
