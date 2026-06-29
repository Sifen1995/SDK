package application

import (
	"context"
	"errors"
	"fmt"
	"time"

	campaigndomain "skykin-platform/internal/campaigns/domain"
	campaignvalidation "skykin-platform/internal/campaigns/validation"
	billingdomain "skykin-platform/internal/billing/domain"
	"skykin-platform/internal/campaigns/infrastructure"
	"skykin-platform/internal/campaigns/model"
)

// CampaignModerationService handles operator review, approval, and activation of campaigns.
type CampaignModerationService struct {
	repo     *infrastructure.Repository
	channels billingdomain.ChannelRepository
}

func NewCampaignModerationService(repo *infrastructure.Repository, channels billingdomain.ChannelRepository) *CampaignModerationService {
	return &CampaignModerationService{repo: repo, channels: channels}
}

// ListPending returns campaigns waiting for operator validation.
func (s *CampaignModerationService) ListPending(ctx context.Context) ([]model.Campaign, error) {
	return s.repo.ListPendingModeration(ctx)
}

// Validate approves or rejects a pending campaign after creative checks.
func (s *CampaignModerationService) Validate(ctx context.Context, campaignID, operatorUserID, action, notes string) (*model.Campaign, error) {
	c, err := s.repo.Get(ctx, campaignID)
	if err != nil {
		return nil, errors.New("campaign not found")
	}
	if c.ModerationStatus != campaigndomain.ModerationPending {
		return nil, fmt.Errorf("campaign moderation_status is %q, not pending", c.ModerationStatus)
	}
	if c.IsActive {
		return nil, errors.New("campaign is already active")
	}

	now := time.Now().UTC()
	switch action {
	case "approve":
		ch, err := s.channels.GetByID(ctx, c.ChannelID)
		if err != nil {
			return nil, errors.New("channel not found")
		}
		vr := campaignvalidation.Campaign(c, ch.Code)
		if vr.Status != "passed" {
			return nil, fmt.Errorf("cannot approve: %s", vr.Notes)
		}
		c.ValidationStatus = vr.Status
		c.ValidationNotes = vr.Notes
		c.ModerationStatus = campaigndomain.ModerationApproved
	case "reject":
		c.ModerationStatus = campaigndomain.ModerationRejected
		c.ValidationStatus = "failed"
		if notes != "" {
			c.ValidationNotes = notes
			c.ModerationNotes = notes
		}
	default:
		return nil, errors.New("action must be approve or reject")
	}

	if notes != "" && action == "approve" {
		c.ModerationNotes = notes
	}
	c.ModeratedAt = &now
	c.ModeratedBy = &operatorUserID
	c.UpdatedAt = now

	if err := s.repo.Update(ctx, c); err != nil {
		return nil, err
	}
	return c, nil
}

// Activate sets is_active=true after operator approval (go-live).
func (s *CampaignModerationService) Activate(ctx context.Context, campaignID, operatorUserID string) (*model.Campaign, error) {
	c, err := s.repo.Get(ctx, campaignID)
	if err != nil {
		return nil, errors.New("campaign not found")
	}
	if c.ModerationStatus != campaigndomain.ModerationApproved {
		return nil, errors.New("campaign must be approved before activation")
	}
	if c.ValidationStatus != "passed" {
		return nil, errors.New("campaign validation_status must be passed")
	}
	if c.IsActive {
		return c, nil
	}

	now := time.Now().UTC()
	c.IsActive = true
	c.UpdatedAt = now
	if c.ModeratedBy == nil {
		c.ModeratedBy = &operatorUserID
	}
	if c.ModeratedAt == nil {
		c.ModeratedAt = &now
	}

	if err := s.repo.Update(ctx, c); err != nil {
		return nil, err
	}
	return c, nil
}
