package telebirr

// TelebirrClient is a placeholder client for Telebirr integrations.
type TelebirrClient struct {
	MerchantID string
	APIKey     string
}

func NewTelebirrClient(merchantID, apiKey string) *TelebirrClient {
	return &TelebirrClient{
		MerchantID: merchantID,
		APIKey:     apiKey,
	}
}
