package application

import (
	"context"
	"log/slog"
	"time"

	campaigndomain "skykin-platform/internal/campaigns/domain"
	"skykin-platform/internal/campaigns/model"
	deliverydomain "skykin-platform/internal/delivery/domain"
	intentdomain "skykin-platform/internal/intents/domain"
	"skykin-platform/internal/platform/messaging"

	"github.com/google/uuid"
)

const (
	eventCampaignMatched   = "CampaignMatched"
	defaultFrequencyCap    = 3
	defaultMinConfidence   = 0.70
	intentLookbackDuration = 24 * time.Hour
)

type TargetingJob struct {
	campaignRepo campaigndomain.CampaignRepository
	intentRepo   intentdomain.IntentRepository
	deliveryRepo deliverydomain.DeliveryRepository
	bus          *messaging.Bus
	logger       *slog.Logger
}

func NewTargetingJob(
	campaignRepo campaigndomain.CampaignRepository,
	intentRepo intentdomain.IntentRepository,
	deliveryRepo deliverydomain.DeliveryRepository,
	bus *messaging.Bus,
	logger *slog.Logger,
) *TargetingJob {
	return &TargetingJob{
		campaignRepo: campaignRepo,
		intentRepo:   intentRepo,
		deliveryRepo: deliveryRepo,
		bus:          bus,
		logger:       logger,
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
	userIDs, err := j.intentRepo.FindUsersWithIntent(ctx, campaign.TargetIntent, defaultMinConfidence, since)
	if err != nil {
		j.logger.Error("find users with intent failed", "campaign_id", campaign.ID, "error", err)
		return
	}
	campaignID, err := uuid.Parse(campaign.ID)
	if err != nil {
		return
	}
	cap := defaultFrequencyCap
	for _, userID := range userIDs {
		j.tryMatchUser(ctx, userID, campaignID, campaign, cap)
	}
}

func (j *TargetingJob) tryMatchUser(ctx context.Context, userID, campaignID uuid.UUID, campaign *model.Campaign, cap int) {
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
			"channel":     channelForFormat(campaign.CreativeFormat),
			"intent":      campaign.TargetIntent,
		},
	})
}

func channelForFormat(format string) string {
	switch format {
	case "SMS_PLUS":
		return "sms_plus"
	case "PUSH_PLUS":
		return "push"
	default:
		return "banner"
	}
}
