package http

import (
	"net/http"
	"strings"

	deliveryapp "skykin-platform/internal/delivery/application"

	"github.com/gin-gonic/gin"
)

type TwilioWebhookHandler struct {
	ingest *deliveryapp.TwilioStatusIngestService
}

func NewTwilioWebhookHandler(ingest *deliveryapp.TwilioStatusIngestService) *TwilioWebhookHandler {
	return &TwilioWebhookHandler{ingest: ingest}
}

// Status godoc
// @Summary      Twilio SMS status callback
// @Description  Public webhook for Twilio delivery receipts. Verifies X-Twilio-Signature, then updates sms_send_attempts status (sent/delivered/failed). Does not write billing_events. Mounted only when SMS_PROVIDER=twilio. Content-Type: application/x-www-form-urlencoded.
// @Tags         SDK - SMS+
// @Accept       application/x-www-form-urlencoded
// @Produce      plain
// @Param        send_key        query  string  true   "Idempotent send key (campaign_id:pseudonymous_id)"
// @Param        MessageSid      formData  string  false  "Twilio message SID"
// @Param        MessageStatus   formData  string  true   "Twilio status (queued|sent|delivered|undelivered|failed)"
// @Param        X-Twilio-Signature  header  string  true  "Twilio request signature"
// @Success      204  "Status recorded"
// @Failure      400  {string}  string
// @Failure      401  {string}  string
// @Failure      503  {string}  string
// @Router       /telemetry/sms/twilio-status [post]
func (h *TwilioWebhookHandler) Status(c *gin.Context) {
	if h == nil || h.ingest == nil {
		c.Status(http.StatusServiceUnavailable)
		return
	}
	if err := c.Request.ParseForm(); err != nil {
		c.Status(http.StatusBadRequest)
		return
	}
	if !h.ingest.VerifySignature(c.Request.URL.String(), c.Request.PostForm, c.GetHeader("X-Twilio-Signature")) {
		c.Status(http.StatusUnauthorized)
		return
	}
	sendKey := strings.TrimSpace(c.Query("send_key"))
	messageSID := strings.TrimSpace(c.PostForm("MessageSid"))
	messageStatus := strings.TrimSpace(c.PostForm("MessageStatus"))
	if sendKey == "" || messageStatus == "" {
		c.Status(http.StatusBadRequest)
		return
	}
	if err := h.ingest.ProcessStatus(c.Request.Context(), sendKey, messageSID, messageStatus); err != nil {
		c.Status(http.StatusBadRequest)
		return
	}
	c.Status(http.StatusNoContent)
}
