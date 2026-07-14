package http

import (
	"context"
	"net/http"
	"strings"

	"skykin-platform/internal/consent/application"
	platformHTTP "skykin-platform/internal/platform/http"

	"github.com/gin-gonic/gin"
)

type consentCreator interface {
	Execute(ctx context.Context, cmd application.CreateConsentCommand) (*application.CreateConsentResult, error)
}

// Handler exposes consent HTTP endpoints for the Flutter SDK.
type Handler struct {
	createConsent consentCreator
}

func NewHandler(createConsent *application.CreateConsentUseCase) *Handler {
	return &Handler{createConsent: createConsent}
}

// CreateConsent godoc
// @Summary      Register SDK user consent
// @Description  Flutter sends only consent_level and sdk_version. In Swagger Authorize with X-API-Key (pk_live_...) and X-SDK-Secret (sk_secret_...); the UI auto-computes X-Signature. Backend generates pseudonymous_id and event-drives user + mapping + consent.
// @Tags         SDK - Consent
// @Accept       json
// @Produce      json
// @Security     APIKeyAuth && SDKSecretAuth
// @Param        body  body  CreateConsentRequest  true  "Consent registration"
// @Success      201   {object}  CreateConsentResponse
// @Failure      400   {object}  platformHTTP.APIError
// @Failure      401   {object}  platformHTTP.APIError
// @Failure      500   {object}  platformHTTP.APIError
// @Router       /consent [post]
func (h *Handler) CreateConsent(c *gin.Context) {
	var req CreateConsentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		platformHTTP.Error(c, http.StatusBadRequest, "invalid consent payload", err.Error())
		return
	}

	result, err := h.createConsent.Execute(c.Request.Context(), application.CreateConsentCommand{
		ConsentLevel: req.ConsentLevel,
		SDKVersion:   req.SDKVersion,
	})
	if err != nil {
		msg := err.Error()
		status := http.StatusInternalServerError
		if strings.Contains(msg, "required") || strings.Contains(msg, "invalid consent level") {
			status = http.StatusBadRequest
		}
		platformHTTP.Error(c, status, "consent registration failed", msg)
		return
	}

	c.JSON(http.StatusCreated, CreateConsentResponse{
		Status:         "success",
		ConsentLevel:   result.ConsentLevel,
		PseudonymousID: result.PseudonymousID,
	})
}
