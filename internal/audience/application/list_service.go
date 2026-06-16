package application

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	billingdomain "skykin-platform/internal/billing/domain"
	audiencedomain "skykin-platform/internal/audience/domain"
	"skykin-platform/internal/audience/model"
)

// SegmentDTO is the portal-facing view of a purchasable audience cohort.
type SegmentDTO struct {
	ID               string   `json:"id"`
	Name             string   `json:"name"`
	Description      string   `json:"description"`
	TopIntentSignals []string `json:"top_intent_signals"`
	ApproximateSize  int      `json:"approximate_size"`
	EstimatedCPM     float64  `json:"estimated_cpm"`
	EstimatedPrice   float64  `json:"estimated_price_etb"` // MVP: estimated_cpm × impression bundle / 1000
	Purchasable      bool     `json:"purchasable"`
}

// ListSegmentsResult is returned by the browse endpoint.
type ListSegmentsResult struct {
	PlanName           string       `json:"plan_name"`
	AudiencemartEnabled bool        `json:"audiencemart_enabled"`
	Segments           []SegmentDTO `json:"segments"`
}

// ListService exposes Audiencemart catalog browsing filtered by subscription plan.
type ListService struct {
	segments audiencedomain.SegmentRepository
	subs     billingdomain.SubscriptionRepository
}

func NewListService(segments audiencedomain.SegmentRepository, subs billingdomain.SubscriptionRepository) *ListService {
	return &ListService{segments: segments, subs: subs}
}

// ListForAdvertiser returns segments the advertiser can purchase on their current plan.
func (s *ListService) ListForAdvertiser(ctx context.Context, advertiserID string) (*ListSegmentsResult, error) {
	sub, err := s.subs.GetActiveByAdvertiser(ctx, advertiserID)
	if err != nil {
		return nil, errors.New("no active subscription; subscribe to a plan first")
	}
	result := &ListSegmentsResult{
		PlanName:            sub.Plan.Name,
		AudiencemartEnabled: sub.Plan.AudiencemartEnabled,
		Segments:            []SegmentDTO{},
	}
	if !sub.Plan.AudiencemartEnabled {
		return result, nil
	}
	rows, err := s.segments.ListAvailableNow(ctx, time.Now().UTC())
	if err != nil {
		return nil, err
	}
	for i := range rows {
		result.Segments = append(result.Segments, toSegmentDTO(&rows[i], true))
	}
	return result, nil
}

func toSegmentDTO(seg *model.AudienceSegment, purchasable bool) SegmentDTO {
	var signals []string
	_ = json.Unmarshal(seg.TopIntentSignals, &signals)
	price := seg.EstimatedCPM * float64(impressionBundle) / 1000.0
	return SegmentDTO{
		ID:               seg.ID,
		Name:             seg.Name,
		Description:      seg.Description,
		TopIntentSignals: signals,
		ApproximateSize:  seg.ApproximateSize,
		EstimatedCPM:     seg.EstimatedCPM,
		EstimatedPrice:   price,
		Purchasable:      purchasable,
	}
}
