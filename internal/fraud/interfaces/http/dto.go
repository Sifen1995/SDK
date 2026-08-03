package http

type ThreatReportRequest struct {
	ThreatType      string `json:"threat_type" binding:"required" example:"url_phishing" enums:"url_phishing,financial_scam,brand_impersonation"`
	Severity        string `json:"severity" binding:"required" example:"high" enums:"low,medium,high,critical"`
	SenderHash      string `json:"sender_hash,omitempty" example:"06f89eae59f69e7bc024476c2c77a1e3f02af36ab6692370b8e394f329afeb11"`
	URLDomain       string `json:"url_domain,omitempty" example:"https://telebirr-verify.example/login"`
	DetectionSource string `json:"detection_source" binding:"required" example:"ml" enums:"blocklist,pattern,ml"`
	SDKVersion      string `json:"sdk_version" binding:"required" example:"1.0.0"`
}

type ThreatReportAcceptedResponse struct {
	Status     string `json:"status" example:"accepted"`
	ReportID   string `json:"report_id" example:"8ae89c1c-bfd8-40d2-9b78-50f5c46c413f"`
	ReportedAt string `json:"reported_at" example:"2026-08-03T09:20:00Z"`
}

type SyncResponse struct {
	Status         string             `json:"status" example:"success"`
	Mode           string             `json:"mode" example:"delta" enums:"full,delta"`
	NextCursor     string             `json:"next_cursor" example:"2026-08-03T08:30:00.123456Z"`
	BlockedDomains []BlockedDomainDTO `json:"blocked_domains"`
	BlockedSenders []BlockedSenderDTO `json:"blocked_senders"`
	ScamPatterns   []ScamPatternDTO   `json:"scam_patterns"`
}

type BlockedDomainDTO struct {
	Domain     string  `json:"domain" example:"telebirr-verify.example"`
	ThreatType string  `json:"threat_type" example:"url_phishing"`
	Severity   string  `json:"severity" example:"critical"`
	Source     string  `json:"source" example:"manual_review"`
	Status     string  `json:"status" example:"active" enums:"active,revoked"`
	CreatedAt  string  `json:"created_at" example:"2026-08-01T10:00:00Z"`
	ExpiresAt  *string `json:"expires_at,omitempty" example:"2026-09-01T10:00:00Z"`
	UpdatedAt  string  `json:"updated_at" example:"2026-08-03T08:20:00Z"`
}

type BlockedSenderDTO struct {
	SenderHash string `json:"sender_hash" example:"a1b2c3d4..."`
	ThreatType string `json:"threat_type" example:"financial_scam"`
	Severity   string `json:"severity" example:"high"`
	Source     string `json:"source" example:"community_report"`
	Status     string `json:"status" example:"active" enums:"active,revoked"`
	CreatedAt  string `json:"created_at" example:"2026-08-01T10:00:00Z"`
	UpdatedAt  string `json:"updated_at" example:"2026-08-03T08:20:00Z"`
}

type ScamPatternDTO struct {
	ID             string  `json:"id" example:"urgent-telebirr-link"`
	PatternType    string  `json:"pattern_type" example:"regex"`
	PatternValue   string  `json:"pattern_value" example:"(?i)urgent.*telebirr.*https?://"`
	ThreatCategory string  `json:"threat_category" example:"brand_impersonation"`
	Confidence     float64 `json:"confidence" example:"0.95"`
	Language       string  `json:"language" example:"en"`
	IsActive       bool    `json:"is_active" example:"true"`
	UpdatedAt      string  `json:"updated_at" example:"2026-08-03T08:20:00Z"`
}
