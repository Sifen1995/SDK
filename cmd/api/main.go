package main

import (
	"context"
	_ "embed"
	"fmt"
	"log"
	"log/slog"
	"net/http"
	"time"

	"skykin-platform/configs"
	_ "skykin-platform/docs"
	advertiserApp "skykin-platform/internal/ad_portal/application"
	advertiserInfra "skykin-platform/internal/ad_portal/infrastructure"
	"skykin-platform/internal/platform/bootstrap"
	"skykin-platform/internal/platform/database"
	"skykin-platform/internal/platform/messaging"
	"skykin-platform/internal/platform/route"
	"skykin-platform/internal/platform/websocket"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

//go:embed swagger_index.html
var swaggerIndexHTML []byte

// @title           Skykin Platform API
// @version         1.0
// @description     Skykin platform API — developer portal (SDK keys), ad campaign portal (advertisers/operators), SDK consent registration, event ingestion, intent prediction, campaign ad delivery via WebSocket, and reward notifications.

// @host            localhost:8081
// @BasePath        /api/v1

// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @description Paste the JWT from POST /ad-portal/login (token field only, no "Bearer " prefix needed)

// @securityDefinitions.apikey APIKeyAuth
// @in header
// @name X-API-Key
// @description SDK publishable key (pk_live_...). Required for /api/v1 SDK routes.

// @securityDefinitions.apikey SDKSecretAuth
// @in header
// @name X-SDK-Secret
// @description SDK secret key (sk_secret_...). Swagger UI uses this ONLY to auto-compute X-Signature (HMAC-SHA256 of the request body). The secret is stripped before the request is sent.

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

	adRepo := advertiserInfra.NewRepository(db)
	_ = adRepo.SeedRoles(context.Background())
	if err := advertiserApp.NewAuthService(adRepo, cfg).EnsureOperatorAdmin(
		context.Background(), cfg.AdminEmail, cfg.AdminPassword, "Operator Admin", "",
	); err != nil {
		log.Printf("operator admin seed skipped: %v", err)
	} else {
		log.Printf("operator admin ready: %s", cfg.AdminEmail)
	}

	hub := websocket.NewHub()
	bus := messaging.NewBus()

	var rdb *redis.Client
	if addr := cfg.RedisAddr; addr != "" {
		client := redis.NewClient(&redis.Options{Addr: addr})
		if err := client.Ping(context.Background()).Err(); err != nil {
			slog.Warn("redis unavailable for permissions cache", "error", err)
		} else {
			rdb = client
		}
	}
	checker, permHandler := bootstrap.NewPermissionSystem(db, rdb, bus, slog.Default())

	r.GET("/ping", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok", "message": "Ready to build!"})
	})

	// Custom Swagger index: Authorize with pk + sk; UI computes X-Signature via HMAC.
	swaggerHandler := ginSwagger.WrapHandler(swaggerFiles.Handler)
	r.GET("/swagger", func(c *gin.Context) { c.Redirect(http.StatusFound, "/swagger/index.html") })
	r.GET("/swagger/*any", func(c *gin.Context) {
		path := c.Param("any")
		if path == "/" || path == "/index.html" || path == "" {
			c.Data(http.StatusOK, "text/html; charset=utf-8", swaggerIndexHTML)
			return
		}
		swaggerHandler(c)
	})

	classJobs := route.InitRouter(r, db, cfg, hub, bus, checker, permHandler)
	bootstrap.StartTargetingJob(db, bus, slog.Default(), 5*time.Minute)
	bootstrap.StartIntentConsistencyJobs(classJobs, slog.Default())

	// Fire up the HTTP engine instance
	serverAddress := fmt.Sprintf("0.0.0.0:%s", cfg.Port)
	log.Printf("Listening on %s — Boot sequence successful.", serverAddress)
	if err := r.Run(serverAddress); err != nil {
		log.Fatalf("critical failure running engine server: %v", err)
	}
}
