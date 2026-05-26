package main

import (
	"fmt"
	"log"
	"net/http"
	"skykin-platform/configs"
	_ "skykin-platform/docs"
	"skykin-platform/internal/common/database"
	"skykin-platform/internal/common/route"
	"skykin-platform/internal/common/websocket"

	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

// @title           Skykin Platform API
// @version         1.0
// @description     Skykin SDK backend — handles developer authentication, application management, event ingestion, intent prediction, and real-time reward notifications.

// @host            localhost:8081
// @BasePath        /api/v1

// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @description Enter your JWT token as: Bearer <token>

// @securityDefinitions.apikey APIKeyAuth
// @in header
// @name X-API-Key
// @description SDK publishable key (pk_live_...)

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

	// Run GORM auto-migrations
	if err := database.Migrate(db); err != nil {
		log.Fatalf("failed to run database migrations: %v", err)
	}
	log.Println("database migrations completed")

	// Initialize the structural communication hub
	hub := websocket.NewHub()

	r.GET("/ping", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok", "message": "Ready to build!"})
	})

	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	route.InitRouter(r, db, cfg, hub)

	// Fire up the HTTP engine instance
	serverAddress := fmt.Sprintf("0.0.0.0:%s", cfg.Port)
	log.Printf("Listening on %s — Boot sequence successful.", serverAddress)
	if err := r.Run(serverAddress); err != nil {
		log.Fatalf("critical failure running engine server: %v", err)
	}
}
