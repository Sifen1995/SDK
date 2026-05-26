package dto

import "time"

type EventResponseDTO struct {
	EventID           string             `json:"event_id" example:"evt_001"`
	UserID            string             `json:"user_id" example:"user_abc_123"`
	Intent            string             `json:"intent,omitempty" example:"coffee_interest"`
	Confidence        float64            `json:"confidence,omitempty" example:"0.85"`
	RewardTriggered   bool               `json:"reward_triggered" example:"true"`
	Reward            *RewardResponseDTO `json:"reward,omitempty"`
	QueuedForAnalysis bool               `json:"queued_for_analysis" example:"false"`
}

type RewardResponseDTO struct {
	RewardID   string  `json:"reward_id" example:"550e8400-e29b-41d4-a716-446655440000"`
	RewardType string  `json:"reward_type" example:"cashback"`
	Amount     float64 `json:"amount" example:"20"`
	Currency   string  `json:"currency" example:"ETB"`
	Message    string  `json:"message" example:"You earned 20 ETB cashback!"`
}

type EventRequestDTO struct {
	ExternalUserID string                 `json:"external_user_id" binding:"required" example:"user_abc_123"`
	EventID        string                 `json:"event_id" binding:"required" example:"evt_001"`
	EventType      string                 `json:"event_type" binding:"required,oneof=search product_view category_view add_to_cart remove_from_cart signup_complete checkout_started checkout_complete store_visit ad_impression ad_click app_open" example:"product_view" enums:"search,product_view,category_view,add_to_cart,remove_from_cart,signup_complete,checkout_started,checkout_complete,store_visit,ad_impression,ad_click,app_open"`
	Metadata       map[string]interface{} `json:"metadata" binding:"required"`
	Timestamp      time.Time              `json:"timestamp" binding:"required" example:"2026-05-26T12:00:00Z"`
}

type BatchEventItemDTO struct {
	EventID   string                 `json:"event_id" binding:"required" example:"evt_001"`
	EventType string                 `json:"event_type" binding:"required,oneof=search product_view category_view add_to_cart remove_from_cart signup_complete checkout_started checkout_complete store_visit ad_impression ad_click app_open" example:"product_view" enums:"search,product_view,category_view,add_to_cart,remove_from_cart,signup_complete,checkout_started,checkout_complete,store_visit,ad_impression,ad_click,app_open"`
	Metadata  map[string]interface{} `json:"metadata" binding:"required"`
	Timestamp time.Time              `json:"timestamp" binding:"required" example:"2026-05-26T12:00:00Z"`
}

type BatchEventRequestDTO struct {
	ExternalUserID string               `json:"external_user_id" binding:"required" example:"user_abc_123"`
	Events         []BatchEventItemDTO  `json:"events" binding:"required,min=1,dive"`
}

type BatchEventResponseDTO struct {
	UserID         string             `json:"user_id" example:"user_abc_123"`
	EventsReceived int                `json:"events_received" example:"5"`
	Intent         string             `json:"intent,omitempty" example:"coffee_interest"`
	Confidence     float64            `json:"confidence,omitempty" example:"0.85"`
	RewardTriggered bool              `json:"reward_triggered" example:"true"`
	Reward         *RewardResponseDTO `json:"reward,omitempty"`
}