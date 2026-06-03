package http

// PredictIntentRequest triggers intent prediction for a user.
type PredictIntentRequest struct {
	UserID string `json:"user_id" binding:"required" example:"user_test_batch_001"`
}
