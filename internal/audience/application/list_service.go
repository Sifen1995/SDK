package application

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	audiencedomain "skykin-platform/internal/audience/domain"
	audiencevalidation "skykin-platform/internal/audience/validation"
	billingdomain "skykin-platform/internal/billing/domain"

	"gorm.io/gorm"
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
	IsActive         bool     `json:"is_active"`
}

// ListSegmentsResult is returned by segment list endpoints.
type ListSegmentsResult struct {
	PlanName            string       `json:"plan_name,omitempty"`
	AudiencemartEnabled bool         `json:"audiencemart_enabled,omitempty"`
	Segments            []SegmentDTO `json:"segments"`
	Count               int          `json:"count"`
}

// CreateSegmentCmd is the command for operator-created catalog segments.
type CreateSegmentCmd struct {
	Name             string
	Description      string
	TopIntentSignals []string
	ApproximateSize  int
	EstimatedCPM     float64
	AvailableFrom    *time.Time
	AvailableUntil   *time.Time
	IsActive         bool
}

// ListService manages Audiencemart catalog reads and operator segment mutations.
type ListService struct {
	segments audiencedomain.SegmentRepository
	subs     billingdomain.SubscriptionRepository
}

func NewListService(segments audiencedomain.SegmentRepository, subs billingdomain.SubscriptionRepository) *ListService {
	return &ListService{segments: segments, subs: subs}
}

// ListForAdvertiser returns active segments the advertiser can purchase on their current plan.
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
	result.Count = len(result.Segments)
	return result, nil
}

// ListAll returns every catalog segment for operator admin (active and inactive).
func (s *ListService) ListAll(ctx context.Context) (*ListSegmentsResult, error) {
	rows, err := s.segments.ListAll(ctx)
	if err != nil {
		return nil, err
	}
	segments := make([]SegmentDTO, len(rows))
	for i := range rows {
		purchasable := isSegmentAvailable(&rows[i], time.Now().UTC())
		segments[i] = toSegmentDTO(&rows[i], purchasable)
	}
	return &ListSegmentsResult{Segments: segments, Count: len(segments)}, nil
}

// GetForAdvertiser returns a single active segment when the advertiser plan allows Audiencemart.
func (s *ListService) GetForAdvertiser(ctx context.Context, advertiserID, segmentID string) (*SegmentDTO, error) {
	if err := audiencevalidation.ValidateSegmentID(segmentID); err != nil {
		return nil, err
	}
	sub, err := s.subs.GetActiveByAdvertiser(ctx, advertiserID)
	if err != nil {
		return nil, errors.New("no active subscription; subscribe to a plan first")
	}
	if !sub.Plan.AudiencemartEnabled {
		return nil, errors.New("audiencemart not enabled on current plan")
	}
	seg, err := s.segments.GetByID(ctx, segmentID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("segment not found")
		}
		return nil, err
	}
	if !isSegmentAvailable(seg, time.Now().UTC()) {
		return nil, errors.New("segment not found")
	}
	dto := toSegmentDTO(seg, true)
	return &dto, nil
}

// GetForAdmin returns any catalog segment by id for operator review.
func (s *ListService) GetForAdmin(ctx context.Context, segmentID string) (*audiencedomain.AudienceSegment, error) {
	if err := audiencevalidation.ValidateSegmentID(segmentID); err != nil {
		return nil, err
	}
	seg, err := s.segments.GetByID(ctx, segmentID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("segment not found")
		}
		return nil, err
	}
	return seg, nil
}

// CreateSegment adds a new Audiencemart catalog segment.
func (s *ListService) CreateSegment(ctx context.Context, cmd CreateSegmentCmd) (*audiencedomain.AudienceSegment, error) {
	if err := audiencevalidation.ValidateCreateSegment(audiencevalidation.CreateSegmentInput{
		Name:             cmd.Name,
		TopIntentSignals: cmd.TopIntentSignals,
		ApproximateSize:  cmd.ApproximateSize,
		EstimatedCPM:     cmd.EstimatedCPM,
	}); err != nil {
		return nil, err
	}
	if _, err := s.segments.GetByName(ctx, strings.TrimSpace(cmd.Name)); err == nil {
		return nil, errors.New("segment with this name already exists")
	} else if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}

	availableFrom := time.Now().UTC()
	if cmd.AvailableFrom != nil {
		availableFrom = cmd.AvailableFrom.UTC()
	}

	seg := &audiencedomain.AudienceSegment{
		Name:             strings.TrimSpace(cmd.Name),
		Description:      strings.TrimSpace(cmd.Description),
		TopIntentSignals: cmd.TopIntentSignals,
		ApproximateSize:  cmd.ApproximateSize,
		EstimatedCPM:     cmd.EstimatedCPM,
		AvailableFrom:    availableFrom,
		AvailableUntil:   cmd.AvailableUntil,
		IsActive:         cmd.IsActive,
	}
	if err := s.segments.Create(ctx, seg); err != nil {
		return nil, err
	}
	return seg, nil
}

// SuspendSegment deactivates an active catalog segment so it is hidden from advertisers.
func (s *ListService) SuspendSegment(ctx context.Context, segmentID string) (*audiencedomain.AudienceSegment, error) {
	if err := audiencevalidation.ValidateSegmentID(segmentID); err != nil {
		return nil, err
	}
	seg, err := s.segments.GetByID(ctx, segmentID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("segment not found")
		}
		return nil, err
	}
	if !seg.IsActive {
		return nil, errors.New("segment is already suspended")
	}
	seg.IsActive = false
	if err := s.segments.Update(ctx, seg); err != nil {
		return nil, fmt.Errorf("suspend segment: %w", err)
	}
	return seg, nil
}

func isSegmentAvailable(seg *audiencedomain.AudienceSegment, now time.Time) bool {
	if !seg.IsActive || seg.AvailableFrom.After(now) {
		return false
	}
	if seg.AvailableUntil != nil && seg.AvailableUntil.Before(now) {
		return false
	}
	return true
}

func toSegmentDTO(seg *audiencedomain.AudienceSegment, purchasable bool) SegmentDTO {
	price := seg.EstimatedCPM * float64(impressionBundle) / 1000.0
	return SegmentDTO{
		ID:               seg.ID,
		Name:             seg.Name,
		Description:      seg.Description,
		TopIntentSignals: seg.TopIntentSignals,
		ApproximateSize:  seg.ApproximateSize,
		EstimatedCPM:     seg.EstimatedCPM,
		EstimatedPrice:   price,
		Purchasable:      purchasable,
		IsActive:         seg.IsActive,
	}
}
