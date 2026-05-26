package routes

import (
	"skykin-platform/configs"
	"skykin-platform/internal/common/websocket"
	"skykin-platform/internal/events/controller"
	eventsRepo "skykin-platform/internal/events/repository"
	"skykin-platform/internal/events/service"
	"skykin-platform/internal/intents/mlclient"
	intentsRepo "skykin-platform/internal/intents/repository"
	rewardsRepo "skykin-platform/internal/rewards/repository"
	usersRepo "skykin-platform/internal/users/repository"
	"strings"
 
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func RegisterRoutes(r *gin.RouterGroup, db *gorm.DB, config *configs.Config, hub *websocket.Hub) {
	eRepo := eventsRepo.NewEventRepository(db, config)
	uRepo := usersRepo.NewUserRepository(db)
	iRepo := intentsRepo.NewIntentRepository(db, config)
	rRepo := rewardsRepo.NewRewardRepository(db)

	mlURL := strings.TrimSpace(config.MLServiceURL)
	if mlURL == "" {
		mlURL = "http://localhost:8000"
	}
	mlClient := mlclient.NewMLClient(strings.TrimSuffix(mlURL, "/"))

	svc := service.NewEventService(eRepo, uRepo, mlClient, rRepo, iRepo, hub)
	ctrl := controller.NewEventController(svc)

	r.POST("/events", ctrl.PostEvent)
	r.POST("/events/batch", ctrl.PostBatchEvents)
}
