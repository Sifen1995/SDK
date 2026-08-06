package application

import (
	"context"
	"errors"
	"fmt"
	"time"

	adminEvents "skykin-platform/internal/admin/events"
	billingdomain "skykin-platform/internal/billing/domain"
	campaigndomain "skykin-platform/internal/campaigns/domain"
	campaignvalidation "skykin-platform/internal/campaigns/validation"
	"skykin-platform/internal/platform/messaging"
)

// LinkedGeofenceActivator turns on store zones attached to a campaign after admin approve.
type LinkedGeofenceActivator interface {
	ActivateForCampaign(ctx context.Context, campaignID string) error
}

// ModerationService handles operator review, approval, and activation of campaigns.
type ModerationService struct {
	repo      campaigndomain.CampaignRepository
	channels  billingdomain.ChannelRepository
	bus       *messaging.Bus
	geofences LinkedGeofenceActivator
}

func NewModerationService(
	repo campaigndomain.CampaignRepository,
	channels billingdomain.ChannelRepository,
	bus *messaging.Bus,
	geofences ...LinkedGeofenceActivator,
) *ModerationService {
	svc := &ModerationService{repo: repo, channels: channels, bus: bus}
	if len(geofences) > 0 {
		svc.geofences = geofences[0]
	}
	return svc
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
		// One admin approve goes live: campaign + linked geofence stores.
		c.IsActive = true
	case "reject":
		c.ModerationStatus = campaigndomain.ModerationRejected
		c.ValidationStatus = "failed"
		c.IsActive = false
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

	if action == "approve" && s.geofences != nil {
		if err := s.geofences.ActivateForCampaign(ctx, campaignID); err != nil {
			return nil, fmt.Errorf("activate linked geofence zones: %w", err)
		}
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
			s.bus.Publish(messaging.Event{
				Name: adminEvents.TopicCampaignActivated,
				Ctx:  ctx,
				Payload: adminEvents.CampaignActivatedEvent{
					CampaignID:     campaignID,
					OperatorUserID: operatorUserID,
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

	newlyActivated := false
	if !c.IsActive {
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
		newlyActivated = true
	}

	// Activate any still-inactive linked stores (e.g. linked after campaign approve).
	if s.geofences != nil {
		if err := s.geofences.ActivateForCampaign(ctx, campaignID); err != nil {
			return nil, fmt.Errorf("activate linked geofence zones: %w", err)
		}
	}

	if newlyActivated && s.bus != nil {
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
