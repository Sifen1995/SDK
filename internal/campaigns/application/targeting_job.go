package application

import (
	"context"
	"log/slog"
	"strings"
	"time"

	audienceapp "skykin-platform/internal/audience/application"
	audiencedomain "skykin-platform/internal/audience/domain"
	campaigndomain "skykin-platform/internal/campaigns/domain"
	billingdomain "skykin-platform/internal/billing/domain"
	"skykin-platform/internal/campaigns/model"
	deliverydomain "skykin-platform/internal/delivery/domain"
	intentdomain "skykin-platform/internal/intents/domain"
	"skykin-platform/internal/platform/messaging"

	"github.com/google/uuid"
)

const (
	eventCampaignMatched   = "CampaignMatched"
	defaultMinConfidence   = 0.70
	intentLookbackDuration = 24 * time.Hour
)

type TargetingJob struct {
	campaignRepo   campaigndomain.CampaignRepository
	intentRepo     intentdomain.IntentRepository
	membershipRepo audiencedomain.MembershipRepository
	deliveryRepo   deliverydomain.DeliveryRepository
	channels       billingdomain.ChannelRepository
	segmentMatch   *audienceapp.TargetingResolver
	bus            *messaging.Bus
	logger         *slog.Logger
}

func NewTargetingJob(
	campaignRepo campaigndomain.CampaignRepository,
	intentRepo intentdomain.IntentRepository,
	membershipRepo audiencedomain.MembershipRepository,
	deliveryRepo deliverydomain.DeliveryRepository,
	channels billingdomain.ChannelRepository,
	segmentMatch *audienceapp.TargetingResolver,
	bus *messaging.Bus,
	logger *slog.Logger,
) *TargetingJob {
	return &TargetingJob{
		campaignRepo:   campaignRepo,
		intentRepo:     intentRepo,
		membershipRepo: membershipRepo,
		deliveryRepo:   deliveryRepo,
		channels:       channels,
		segmentMatch:   segmentMatch,
		bus:            bus,
		logger:         logger,
	}
}

func (j *TargetingJob) Run(ctx context.Context) {
	campaigns, err := j.campaignRepo.ListActive(ctx)
	if err != nil {
		j.logger.Error("targeting job list campaigns failed", "error", err)
		return
	}
	since := time.Now().Add(-intentLookbackDuration)
	for i := range campaigns {
		j.matchCampaign(ctx, &campaigns[i], since)
	}
}

func (j *TargetingJob) matchCampaign(ctx context.Context, campaign *model.Campaign, since time.Time) {
	if !j.isWithinSchedule(campaign, time.Now().UTC()) {
		return
	}

	var userIDs []uuid.UUID
	var err error

	if campaign.SegmentID != nil && strings.TrimSpace(*campaign.SegmentID) != "" {
		segID, parseErr := uuid.Parse(*campaign.SegmentID)
		if parseErr != nil {
			j.logger.Error("invalid segment_id", "campaign_id", campaign.ID, "segment_id", *campaign.SegmentID)
			return
		}
		userIDs, err = j.membershipRepo.FindUsersInSegment(ctx, segID)
		if err != nil {
			j.logger.Error("segment membership query failed", "segment_id", campaign.SegmentID, "error", err)
			return
		}
	} else {
		userIDs, err = j.intentRepo.FindUsersWithIntent(ctx, campaign.TargetIntent, defaultMinConfidence, since)
		if err != nil {
			j.logger.Error("resolve targeting users failed", "campaign_id", campaign.ID, "error", err)
			return
		}
	}
	campaignID, err := uuid.Parse(campaign.ID)
	if err != nil {
		return
	}
	cap := campaign.FrequencyCapPerDay
	if cap <= 0 {
		cap = 3
	}
	channelCode := j.resolveChannelCode(ctx, campaign.ChannelID)
	for _, userID := range userIDs {
		j.tryMatchUser(ctx, userID, campaignID, campaign, cap, channelCode)
	}
}

// isWithinSchedule skips campaigns outside their scheduled window (nil bounds = no limit).
func (j *TargetingJob) isWithinSchedule(c *model.Campaign, now time.Time) bool {
	if c.ScheduledStartAt != nil && now.Before(*c.ScheduledStartAt) {
		return false
	}
	if c.ScheduledEndAt != nil && now.After(*c.ScheduledEndAt) {
		return false
	}
	return true
}

func (j *TargetingJob) resolveChannelCode(ctx context.Context, channelID string) string {
	ch, err := j.channels.GetByID(ctx, channelID)
	if err != nil {
		return "banner"
	}
	return channelCodeForBus(ch.Code)
}

func (j *TargetingJob) tryMatchUser(ctx context.Context, userID, campaignID uuid.UUID, campaign *model.Campaign, cap int, channel string) {
	delivered, err := j.deliveryRepo.WasDelivered(ctx, userID, campaignID)
	if err != nil || delivered {
		return
	}
	count, err := j.deliveryRepo.CountToday(ctx, userID, campaignID)
	if err != nil || count >= cap {
		return
	}
	j.bus.Publish(messaging.Event{
		Name: eventCampaignMatched,
		Ctx:  ctx,
		Payload: map[string]any{
			"user_id":     userID,
			"campaign_id": campaignID,
			"creative_id": campaign.ID,
			"channel":     channel,
			"intent":      campaign.TargetIntent,
		},
	})
}

func channelCodeForBus(code string) string {
	switch strings.ToUpper(code) {
	case "SMS_PLUS":
		return "sms_plus"
	case "PUSH":
		return "push"
	case "NATIVE_FEED":
		return "native_feed"
	default:
		return "banner"
	}
}
