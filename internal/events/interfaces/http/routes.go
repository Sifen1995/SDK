package http

import (
	"log/slog"
	"strings"

	"skykin-platform/configs"
	"skykin-platform/internal/events/application"
	eventsInfra "skykin-platform/internal/events/infrastructure"
	"skykin-platform/internal/platform/messaging"
	platformredis "skykin-platform/internal/platform/redis"
	usersInfra "skykin-platform/internal/users/infrastructure"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// Module wires dependencies for the events bounded context.
type Module struct {
	Handler *Handler
	Bus     *messaging.Bus
}

// NewModule constructs the events module with dependency injection.
func NewModule(db *gorm.DB, cfg *configs.Config, bus *messaging.Bus) *Module {
	if bus == nil {
		bus = messaging.NewBus()
	}

	repo := eventsInfra.NewPostgresRepository(db)
	userRepo := usersInfra.NewUserRepository(db)
	redisClient := eventsInfra.NewRedisClientFromAddr(cfg.RedisAddr)
	dedup := eventsInfra.NewRedisDedupStore(redisClient)
	publisher := eventsInfra.NewBusEventPublisher(bus)
	var redisQueueClient *platformredis.RedisClient
	if addr := strings.TrimSpace(cfg.RedisAddr); addr != "" {
		if c, err := platformredis.NewRedisClient(addr); err == nil {
			redisQueueClient = c
		}
	}

	ingest := application.NewIngestEventsUseCase(repo, userRepo, dedup, redisQueueClient, publisher, slog.Default())
	handler := NewHandler(ingest)

	return &Module{Handler: handler, Bus: bus}
}

// RegisterRoutes mounts events HTTP routes on the SDK API group.
func RegisterRoutes(r *gin.RouterGroup, module *Module) {
	r.POST("/events", module.Handler.PostEvents)
}
