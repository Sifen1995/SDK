package application

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"strings"
	"time"

	deliverydomain "skykin-platform/internal/delivery/domain"

	"gorm.io/gorm"
)

type SMSCampaign struct {
	ID             string
	Title          string
	BodyText       string
	DestinationURL string
}

type SMSCampaignReader interface {
	GetSMSCampaign(ctx context.Context, campaignID string) (*SMSCampaign, error)
}

type SMSDispatchService struct {
	campaigns  SMSCampaignReader
	recipients deliverydomain.DemoSMSRecipientRepository
	attempts   deliverydomain.SMSSendAttemptRepository
	logs       deliverydomain.DeliveryLogRepository
	provider   SMSProvider
	baseURL    string
	secretKey  []byte
}

func NewSMSDispatchService(
	campaigns SMSCampaignReader,
	recipients deliverydomain.DemoSMSRecipientRepository,
	attempts deliverydomain.SMSSendAttemptRepository,
	logs deliverydomain.DeliveryLogRepository,
	provider SMSProvider,
	baseURL string,
	secret string,
) *SMSDispatchService {
	return &SMSDispatchService{
		campaigns:  campaigns,
		recipients: recipients,
		attempts:   attempts,
		logs:       logs,
		provider:   provider,
		baseURL:    strings.TrimRight(strings.TrimSpace(baseURL), "/"),
		secretKey:  []byte(strings.TrimSpace(secret)),
	}
}

func (s *SMSDispatchService) DispatchCampaignMatch(
	ctx context.Context,
	campaignID, pseudonymousID string,
) error {
	campaignID = strings.TrimSpace(campaignID)
	pseudonymousID = strings.TrimSpace(pseudonymousID)
	if campaignID == "" || pseudonymousID == "" {
		return fmt.Errorf("campaign_id and pseudonymous_id are required")
	}
	if s == nil || s.campaigns == nil || s.recipients == nil || s.attempts == nil || s.logs == nil || s.provider == nil {
		return fmt.Errorf("sms dispatch service is not configured")
	}

	recipient, err := s.recipients.FindActiveByPseudonymousID(ctx, pseudonymousID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil
		}
		return err
	}
	campaign, err := s.campaigns.GetSMSCampaign(ctx, campaignID)
	if err != nil {
		return err
	}

	sendKey := fmt.Sprintf("%s:%s", campaignID, pseudonymousID)
	exists, err := s.attempts.ExistsBySendKey(ctx, sendKey)
	if err != nil {
		return err
	}
	if exists {
		return nil
	}

	trackingToken := s.buildTrackingToken(sendKey)
	body := s.renderBody(campaign, trackingToken)
	attempt := &deliverydomain.SMSSendAttempt{
		SendKey:        sendKey,
		CampaignID:     campaignID,
		PseudonymousID: pseudonymousID,
		UserID:         recipient.UserID,
		PhoneE164:      recipient.PhoneE164,
		Provider:       s.provider.ProviderName(),
		Status:         deliverydomain.SMSSendStatusQueued,
		MessageBody:    body,
		TrackingToken:  trackingToken,
		DestinationURL: strings.TrimSpace(campaign.DestinationURL),
	}
	if err := s.attempts.Create(ctx, attempt); err != nil {
		if isDuplicateError(err) {
			return nil
		}
		return err
	}

	result, err := s.provider.Send(ctx, SMSMessage{
		To:                recipient.PhoneE164,
		Body:              body,
		StatusCallbackURL: s.statusCallbackURL(sendKey),
	})
	if err != nil {
		_ = s.attempts.UpdateStatus(ctx, attempt.ID, deliverydomain.SMSSendStatusFailed, "", err.Error())
		return err
	}
	providerMessageID := ""
	if result != nil {
		providerMessageID = strings.TrimSpace(result.ProviderMessageID)
	}
	if err := s.attempts.UpdateStatus(ctx, attempt.ID, deliverydomain.SMSSendStatusSent, providerMessageID, ""); err != nil {
		return err
	}
	return s.logs.CreateBatch(ctx, []deliverydomain.DeliveryLog{{
		CampaignID:     campaignID,
		PseudonymousID: pseudonymousID,
		SessionID:      sendKey,
		DeliveryStatus: deliverydomain.StatusDispatched,
		LoggedAt:       time.Now().UTC(),
	}})
}

func (s *SMSDispatchService) renderBody(campaign *SMSCampaign, trackingToken string) string {
	if campaign == nil {
		return ""
	}
	parts := make([]string, 0, 3)
	if title := strings.TrimSpace(campaign.Title); title != "" {
		parts = append(parts, title)
	}
	if body := strings.TrimSpace(campaign.BodyText); body != "" {
		parts = append(parts, body)
	}
	if s.baseURL != "" && trackingToken != "" {
		parts = append(parts, fmt.Sprintf("%s/api/v1/telemetry/sms/click?token=%s", s.baseURL, trackingToken))
	}
	return strings.Join(parts, " - ")
}

func (s *SMSDispatchService) statusCallbackURL(sendKey string) string {
	if s.baseURL == "" {
		return ""
	}
	return fmt.Sprintf("%s/api/v1/telemetry/sms/twilio-status?send_key=%s", s.baseURL, sendKey)
}

func (s *SMSDispatchService) buildTrackingToken(sendKey string) string {
	h := hmac.New(sha256.New, s.secretKey)
	h.Write([]byte(sendKey))
	return hex.EncodeToString(h.Sum(nil))
}

func isDuplicateError(err error) bool {
	if err == nil {
		return false
	}
	msg := strings.ToLower(err.Error())
	return strings.Contains(msg, "duplicate") || strings.Contains(msg, "23505")
}
