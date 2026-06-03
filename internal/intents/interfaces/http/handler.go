package http

import (
	"net/http"

	"skykin-platform/internal/intents/application"
	platformHTTP "skykin-platform/internal/platform/http"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	predict *application.PredictIntentUseCase
}

func NewHandler(predict *application.PredictIntentUseCase) *Handler {
	return &Handler{predict: predict}
}

// PredictIntent godoc
// @Summary      Predict user intent (sync)
// @Description  Loads stored/cached events for a user, calls ML, persists intent and reward when threshold is met.
// @Tags         SDK - Intents
// @Accept       json
// @Produce      json
// @Security     APIKeyAuth
// @Param        body  body      PredictIntentRequest  true  "User id to evaluate"
// @Success      200   {object}  application.PredictIntentResult
// @Failure      400   {object}  platformHTTP.APIError
// @Failure      500   {object}  platformHTTP.APIError
// @Router       /intents/predict [post]
func (h *Handler) PredictIntent(c *gin.Context) {
	var req PredictIntentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		platformHTTP.Error(c, http.StatusBadRequest, "invalid predict payload", err.Error())
		return
	}

	result, err := h.predict.Execute(c.Request.Context(), req.UserID)
	if err != nil {
		platformHTTP.Error(c, http.StatusInternalServerError, "prediction failed", err.Error())
		return
	}

	switch result.Status {
	case "predicted":
		c.JSON(http.StatusOK, result)
	case "insufficient_history", "skipped", "ml_unavailable":
		c.JSON(http.StatusAccepted, result)
	default:
		c.JSON(http.StatusOK, result)
	}
}
