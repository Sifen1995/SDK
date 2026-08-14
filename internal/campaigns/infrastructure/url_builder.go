package infrastructure

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"net/url"
)

type PlayLinkBuilder struct {
	secretKey []byte
}

func NewPlayLinkBuilder(secretKey string) *PlayLinkBuilder {
	return &PlayLinkBuilder{secretKey: []byte(secretKey)}
}

// BuildConsentedInstallURL constructs a Play Store URL containing signed referrer tracking
func (b *PlayLinkBuilder) BuildConsentedInstallURL(packageID, campaignID string) string {
	// Sign campaign_id with HMAC
	h := hmac.New(sha256.New, b.secretKey)
	h.Write([]byte(campaignID))
	token := hex.EncodeToString(h.Sum(nil))

	// Construct internal query parameters
	referrerParams := url.Values{}
	referrerParams.Set("campaign_id", campaignID)
	referrerParams.Set("token", token)
	rawReferrer := referrerParams.Encode()

	// Format final Google Play Store URL. The package id is escaped too — an id
	// carrying `&` or `#` would otherwise break out of the query parameter.
	return fmt.Sprintf(
		"https://play.google.com/store/apps/details?id=%s&referrer=%s",
		url.QueryEscape(packageID),
		url.QueryEscape(rawReferrer),
	)
}
