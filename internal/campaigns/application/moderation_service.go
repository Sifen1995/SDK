package application

import (
	"context"
	"errors"
	"fmt"
	"time"

	adminEvents "skykin-platform/internal/admin/events"
	campaigndomain "skykin-platform/internal/campaigns/domain"
	campaignvalidation "skykin-platform/internal/campaigns/validation"
	billingdomain "skykin-platform/internal/billing/domain"
	"skykin-platform/internal/platform/messaging"
)

// ModerationService handles operator review, approval, and activation of campaigns.
type ModerationService struct {
	repo     campaigndomain.CampaignRepository
	channels billingdomain.ChannelRepository
	bus      *messaging.Bus
}

func NewModerationService(
	repo campaigndomain.CampaignRepository,
	channels billingdomain.ChannelRepository,
	bus *messaging.Bus,
) *ModerationService {
	return &ModerationService{repo: repo, channels: channels, bus: bus}
}

func (s *ModerationService) ListPending(ctx context.Context) ([]campaigndomain.Campaign, error) {
	return s.repo.ListPendingModeration(ctx)
}

func (s *ModerationService) Validate(ctx context.Context, campaignID, operatorUserID, action, notes string) (*campaigndomain.Campaign, error) {
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
	var channelCode string
	switch action {
	case "approve":
		ch, err := s.channels.GetByID(ctx, c.ChannelID)
		if err != nil {
			return nil, errors.New("channel not found")
		}
		channelCode = ch.Code
		vr := campaignvalidation.Campaign(c, channelCode)
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

	if s.bus != nil {
		switch action {
		case "approve":
			s.bus.Publish(messaging.Event{
				Name: adminEvents.TopicCampaignModerationPassed,
				Ctx:  ctx,
				Payload: adminEvents.CampaignModerationPassedEvent{
					CampaignID:     campaignID,
					OperatorUserID: operatorUserID,
					ChannelCode:    channelCode,
				},
			})
		case "reject":
			s.bus.Publish(messaging.Event{
				Name: adminEvents.TopicCampaignModerationRejected,
				Ctx:  ctx,
				Payload: adminEvents.CampaignModerationRejectedEvent{
					CampaignID:     campaignID,
					OperatorUserID: operatorUserID,
					Notes:          notes,
				},
			})
		}
	}

	return c, nil
}

func (s *ModerationService) Activate(ctx context.Context, campaignID, operatorUserID string) (*campaigndomain.Campaign, error) {
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

	if s.bus != nil {
		s.bus.Publish(messaging.Event{
			Name: adminEvents.TopicCampaignActivated,
			Ctx:  ctx,
			Payload: adminEvents.CampaignActivatedEvent{
				CampaignID:     campaignID,
				OperatorUserID: operatorUserID,
			},
		})
	}

	return c, nil
}
