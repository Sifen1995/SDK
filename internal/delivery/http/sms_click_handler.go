package http

import (
	"net/http"
	"strings"

	deliveryapp "skykin-platform/internal/delivery/application"

	"github.com/gin-gonic/gin"
)

type SMSClickHandler struct {
	clicks *deliveryapp.SMSClickService
}

func NewSMSClickHandler(clicks *deliveryapp.SMSClickService) *SMSClickHandler {
	return &SMSClickHandler{clicks: clicks}
}

// Track godoc
// @Summary      Track SMS+ CTA click
// @Description  Public redirect endpoint embedded in SMS+ message bodies. Validates the signed tracking token, enqueues a click event to stream:billing_events (billing + delivery_log consumers write behind), and redirects to the campaign destination_url. No X-API-Key required.
// @Tags         SDK - SMS+
// @Produce      html
// @Param        token  query  string  true  "Signed SMS click tracking token from sms_send_attempts"
// @Success      302  "Redirect to campaign destination_url"
// @Success      202  "Accepted when destination_url is empty"
// @Failure      400  {string}  string  "missing token"
// @Failure      401  {string}  string  "invalid sms click token"
// @Failure      503  {string}  string  "sms click tracking unavailable"
// @Router       /telemetry/sms/click [get]
func (h *SMSClickHandler) Track(c *gin.Context) {
	if h == nil || h.clicks == nil {
		c.String(http.StatusServiceUnavailable, "sms click tracking unavailable")
		return
	}
	token := strings.TrimSpace(c.Query("token"))
	if token == "" {
		c.String(http.StatusBadRequest, "missing token")
		return
	}
	destinationURL, err := h.clicks.ProcessClick(c.Request.Context(), token)
	if err != nil {
		c.String(http.StatusUnauthorized, "invalid sms click token")
		return
	}
	if destinationURL == "" {
		c.Status(http.StatusAccepted)
		return
	}
	c.Redirect(http.StatusFound, destinationURL)
}
