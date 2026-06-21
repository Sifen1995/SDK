package main

import (
	"context"
	"fmt"
	"log"
	"log/slog"
	"net/http"
	"time"

	"skykin-platform/configs"
	_ "skykin-platform/docs"
	adportalApp "skykin-platform/internal/ad_portal/application"
	adportalInfra "skykin-platform/internal/ad_portal/infrastructure"
	"skykin-platform/internal/platform/bootstrap"
	"skykin-platform/internal/platform/database"
	"skykin-platform/internal/platform/messaging"
	"skykin-platform/internal/platform/route"
	"skykin-platform/internal/platform/websocket"

	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

// @title           Skykin Platform API
// @version         1.0
// @description     Skykin platform API — developer portal (SDK keys), ad campaign portal (advertisers/operators), SDK event ingestion, intent prediction, campaign ad delivery via WebSocket, and reward notifications.

// @host            localhost:8081
// @BasePath        /api/v1

// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @description Enter your JWT token as: Bearer <token> (developer portal or ad portal)

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

	adRepo := adportalInfra.NewRepository(db)
	_ = adRepo.SeedRoles(context.Background())
	if err := adportalApp.NewPortalAuthService(adRepo, cfg).EnsureOperatorAdmin(
		context.Background(), cfg.AdminEmail, cfg.AdminPassword, "Operator Admin", "",
	); err != nil {
		log.Printf("operator admin seed skipped: %v", err)
	} else {
		log.Printf("operator admin ready: %s", cfg.AdminEmail)
	}

	hub := websocket.NewHub()
	bus := messaging.NewBus()

	r.GET("/ping", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok", "message": "Ready to build!"})
	})

	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	route.InitRouter(r, db, cfg, hub, bus)
	bootstrap.StartTargetingJob(db, bus, slog.Default(), 5*time.Minute)

	// Fire up the HTTP engine instance
	serverAddress := fmt.Sprintf("0.0.0.0:%s", cfg.Port)
	log.Printf("Listening on %s — Boot sequence successful.", serverAddress)
	if err := r.Run(serverAddress); err != nil {
		log.Fatalf("critical failure running engine server: %v", err)
	}
}
