package routes

import (
	intentApp "skykin-platform/internal/intents/application"
	intentHTTP "skykin-platform/internal/intents/interfaces/http"

	"github.com/gin-gonic/gin"
)

// Module wires intent prediction dependencies for the SDK API.
type Module struct {
	Handler *intentHTTP.Handler
}

// Wire constructs the intents module.
func Wire(predict *intentApp.PredictIntentUseCase) *Module {
	return &Module{Handler: intentHTTP.NewHandler(predict)}
}

// Register mounts SDK intent routes.
func (m *Module) Register(g *gin.RouterGroup) {
	intentHTTP.RegisterRoutes(g, m.Handler)
}
