package application

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"time"

	advertiserdomain "skykin-platform/internal/advertisers/domain"
	audienceapp "skykin-platform/internal/audience/application"
	billingapp "skykin-platform/internal/billing/application"
	billingdomain "skykin-platform/internal/billing/domain"
	"skykin-platform/internal/campaigns/infrastructure"
	"skykin-platform/internal/campaigns/model"

	"gorm.io/datatypes"
	"gorm.io/gorm"
)

// CampaignService orchestrates campaign CRUD with subscription and audience checks.
type CampaignService struct {
	repo         *infrastructure.Repository
	subscriptions *billingapp.SubscriptionEnforcer
	audience     *audienceapp.PurchaseService
	channels     billingdomain.ChannelRepository
}

func NewCampaignService(
	repo *infrastructure.Repository,
	subscriptions *billingapp.SubscriptionEnforcer,
	audience *audienceapp.PurchaseService,
	channels billingdomain.ChannelRepository,
) *CampaignService {
	return &CampaignService{
		repo:          repo,
		subscriptions: subscriptions,
		audience:      audience,
		channels:      channels,
	}
}

// Create validates subscription entitlements, campaign fields, and optionally records a segment purchase.
func (s *CampaignService) Create(ctx context.Context, advertiserID, role string, cmd CreateCampaignCommand) (*model.Campaign, error) {
	if !advertiserdomain.CanWrite(role) {
		return nil, errors.New("forbidden")
	}
	if strings.TrimSpace(advertiserID) == "" {
		return nil, errors.New("account has no advertiser company; operator admins cannot create campaigns directly")
	}

	wantsSegment := cmd.SegmentID != nil && strings.TrimSpace(*cmd.SegmentID) != ""

	// 1. Subscription gate — plan limits, channel tier, Audiencemart access.
	subCtx, err := s.subscriptions.AssertCanCreateCampaign(ctx, advertiserID, cmd.ChannelID, cmd.DailyBudgetCap, wantsSegment)
	if err != nil {
		return nil, err
	}

	// 2. Load channel for creative validation rules.
	channel, err := s.channels.GetByID(ctx, cmd.ChannelID)
	if err != nil {
		return nil, errors.New("channel not found")
	}

	// 3. Audience segment quote (when segment_id provided).
	var purchaseQuote *audienceapp.PurchaseQuote
	if wantsSegment {
		if err := s.audience.ValidateTargetIntent(ctx, *cmd.SegmentID, cmd.TargetIntent); err != nil {
			return nil, err
		}
		purchaseQuote, err = s.audience.PreparePurchase(ctx, advertiserID, *cmd.SegmentID, subCtx.Plan)
		if err != nil {
			return nil, err
		}
	}

	canvas := datatypes.JSON([]byte("{}"))
	if cmd.CanvasJSON != nil {
		raw, _ := json.Marshal(cmd.CanvasJSON)
		canvas = datatypes.JSON(raw)
	}

	segmentID := cmd.SegmentID
	if segmentID != nil && strings.TrimSpace(*segmentID) == "" {
		segmentID = nil
	}

	c := &model.Campaign{
		AdvertiserID:       advertiserID,
		Name:               cmd.Name,
		TargetIntent:       cmd.TargetIntent,
		ChannelID:          cmd.ChannelID,
		SegmentID:          segmentID,
		Title:              cmd.Title,
		BodyText:           cmd.BodyText,
		ImageURL:           cmd.ImageURL,
		DestinationURL:     cmd.DestinationURL,
		CanvasJSON:         canvas,
		BillingModel:       cmd.BillingModel,
		DailyBudgetCap:     cmd.DailyBudgetCap,
		TotalBudgetCap:     cmd.TotalBudgetCap,
		FrequencyCapPerDay: cmd.FrequencyCapPerDay,
		ScheduledStartAt:   cmd.ScheduledStartAt,
		ScheduledEndAt:     cmd.ScheduledEndAt,
		IsActive:           false,
	}

	// 4. Creative validation by channel code.
	vr := ValidateCampaign(c, channel.Code)
	c.ValidationStatus = vr.Status
	c.ValidationNotes = vr.Notes

	// 5. Persist campaign + segment purchase atomically.
	err = s.repo.Transaction(ctx, func(tx *gorm.DB) error {
		if err := s.repo.CreateTx(ctx, tx, c); err != nil {
			return err
		}
		return s.audience.ConfirmPurchaseTx(ctx, tx, purchaseQuote, c.ID)
	})
	if err != nil {
		return nil, err
	}
	return c, nil
}

func (s *CampaignService) List(ctx context.Context, advertiserID, role string) ([]model.Campaign, error) {
	if !advertiserdomain.CanRead(role) {
		return nil, errors.New("forbidden")
	}
	if role == advertiserdomain.RoleOperatorAdmin {
		return s.repo.ListAll(ctx)
	}
	return s.repo.ListByAdvertiser(ctx, advertiserID)
}

func (s *CampaignService) Get(ctx context.Context, advertiserID, role, id string) (*model.Campaign, error) {
	if !advertiserdomain.CanRead(role) {
		return nil, errors.New("forbidden")
	}
	c, err := s.repo.Get(ctx, id)
	if err != nil {
		return nil, err
	}
	if role != advertiserdomain.RoleOperatorAdmin && c.AdvertiserID != advertiserID {
		return nil, errors.New("forbidden")
	}
	return c, nil
}

func (s *CampaignService) Activate(ctx context.Context, advertiserID, role, id string) (*model.Campaign, error) {
	if !advertiserdomain.CanWrite(role) {
		return nil, errors.New("forbidden")
	}
	c, err := s.Get(ctx, advertiserID, role, id)
	if err != nil {
		return nil, err
	}
	if c.ValidationStatus != "passed" {
		return nil, errors.New("campaign cannot be activated until validation_status is passed")
	}
	c.IsActive = true
	c.UpdatedAt = time.Now().UTC()
	if err := s.repo.Update(ctx, c); err != nil {
		return nil, err
	}
	return c, nil
}

func (s *CampaignService) Preview(ctx context.Context, advertiserID, role, id string) (map[string]any, error) {
	c, err := s.Get(ctx, advertiserID, role, id)
	if err != nil {
		return nil, err
	}
	ch, err := s.channels.GetByID(ctx, c.ChannelID)
	if err != nil {
		return nil, errors.New("channel not found for preview")
	}
	return PreviewCampaign(c, ch.Code), nil
}
