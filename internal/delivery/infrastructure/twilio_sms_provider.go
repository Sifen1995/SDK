package infrastructure

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"

	deliveryapp "skykin-platform/internal/delivery/application"
)

type TwilioSMSProvider struct {
	accountSID          string
	authToken           string
	fromNumber          string
	messagingServiceSID string
	client              *http.Client
}

func NewTwilioSMSProvider(
	accountSID, authToken, fromNumber, messagingServiceSID string,
	client *http.Client,
) *TwilioSMSProvider {
	if client == nil {
		client = http.DefaultClient
	}
	return &TwilioSMSProvider{
		accountSID:          strings.TrimSpace(accountSID),
		authToken:           strings.TrimSpace(authToken),
		fromNumber:          strings.TrimSpace(fromNumber),
		messagingServiceSID: strings.TrimSpace(messagingServiceSID),
		client:              client,
	}
}

func (p *TwilioSMSProvider) ProviderName() string { return "twilio" }

func (p *TwilioSMSProvider) Send(ctx context.Context, msg deliveryapp.SMSMessage) (*deliveryapp.SMSSendResult, error) {
	if p.accountSID == "" || p.authToken == "" {
		return nil, fmt.Errorf("twilio credentials are not configured")
	}
	if p.messagingServiceSID == "" && p.fromNumber == "" {
		return nil, fmt.Errorf("twilio sender is not configured")
	}
	form := url.Values{}
	form.Set("To", msg.To)
	form.Set("Body", msg.Body)
	if msg.StatusCallbackURL != "" {
		form.Set("StatusCallback", msg.StatusCallbackURL)
	}
	if p.messagingServiceSID != "" {
		form.Set("MessagingServiceSid", p.messagingServiceSID)
	} else {
		form.Set("From", p.fromNumber)
	}
	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodPost,
		fmt.Sprintf("https://api.twilio.com/2010-04-01/Accounts/%s/Messages.json", p.accountSID),
		strings.NewReader(form.Encode()),
	)
	if err != nil {
		return nil, err
	}
	req.SetBasicAuth(p.accountSID, p.authToken)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	resp, err := p.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("twilio send failed with status %d", resp.StatusCode)
	}
	var payload struct {
		SID string `json:"sid"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		return nil, err
	}
	return &deliveryapp.SMSSendResult{ProviderMessageID: payload.SID}, nil
}
