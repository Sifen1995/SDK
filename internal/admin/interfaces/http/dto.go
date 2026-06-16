package http

// ValidateCampaignRequest is the operator approve/reject payload.
type ValidateCampaignRequest struct {
	Action string `json:"action" binding:"required,oneof=approve reject"`
	Notes  string `json:"notes" binding:"max=2000"`
}

// CampaignListResponse wraps campaigns for admin list endpoints.
type CampaignListResponse struct {
	Campaigns []interface{} `json:"campaigns"`
	Count     int           `json:"count"`
}
