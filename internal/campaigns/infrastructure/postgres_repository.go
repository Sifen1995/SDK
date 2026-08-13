package infrastructure

import (
	"context"
	"fmt"
	"net/url"
	"os"
	"strings"
	"time"

	billingdomain "skykin-platform/internal/billing/domain"
	campaigndomain "skykin-platform/internal/campaigns/domain"
	"skykin-platform/internal/campaigns/infrastructure/persistence"

	"gorm.io/gorm"
)

type eligibleCampaignScan struct {
	persistence.CampaignRow
	PlanID         string  `gorm:"column:plan_id"`
	PlanName       string  `gorm:"column:plan_name"`
	PlanMonthlyFee float64 `gorm:"column:plan_monthly_fee_etb"`
	ChannelCode    string  `gorm:"column:channel_code"`
}

type Repository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) *Repository {
	return &Repository{db: db}
}

var (
	_ campaigndomain.CampaignRepository = (*Repository)(nil)
	_ billingdomain.CampaignQuotaReader = (*Repository)(nil)
)

func (r *Repository) ListActive(ctx context.Context) ([]campaigndomain.Campaign, error) {
	var rows []persistence.CampaignRow
	err := r.db.WithContext(ctx).
		Where("is_active = ? AND validation_status = ? AND moderation_status = ?",
			true, "passed", campaigndomain.ModerationApproved).
		Order("created_at desc").
		Find(&rows).Error
	if err != nil {
		return nil, err
	}
	return toDomainCampaigns(rows)
}

// ListPendingModeration returns campaigns awaiting operator review.
func (r *Repository) ListPendingModeration(ctx context.Context) ([]campaigndomain.Campaign, error) {
	var rows []persistence.CampaignRow
	err := r.db.WithContext(ctx).
		Where("moderation_status = ? AND is_active = ?", campaigndomain.ModerationPending, false).
		Order("created_at asc").
		Find(&rows).Error
	if err != nil {
		return nil, err
	}
	return toDomainCampaigns(rows)
}

func (r *Repository) Create(ctx context.Context, c *campaigndomain.Campaign) error {
	row, err := persistence.CampaignRowFromDomain(c)
	if err != nil {
		return err
	}
	if err := r.db.WithContext(ctx).Create(row).Error; err != nil {
		return err
	}
	c.ID = row.ID
	c.CreatedAt = row.CreatedAt
	c.UpdatedAt = row.UpdatedAt
	return nil
}

// CreateTx inserts a campaign inside an existing database transaction.
func (r *Repository) CreateTx(ctx context.Context, tx any, c *campaigndomain.Campaign) error {
	db, ok := tx.(*gorm.DB)
	if !ok || db == nil {
		return gorm.ErrInvalidTransaction
	}
	row, err := persistence.CampaignRowFromDomain(c)
	if err != nil {
		return err
	}
	if err := db.WithContext(ctx).Create(row).Error; err != nil {
		return err
	}
	c.ID = row.ID
	c.CreatedAt = row.CreatedAt
	c.UpdatedAt = row.UpdatedAt
	return nil
}

// Transaction runs fn inside a single database transaction.
func (r *Repository) Transaction(ctx context.Context, fn func(tx any) error) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		return fn(tx)
	})
}

// CountActiveByAdvertiser counts is_active campaigns for subscription quota enforcement.
func (r *Repository) CountActiveByAdvertiser(ctx context.Context, advertiserID string) (int, error) {
	var n int64
	err := r.db.WithContext(ctx).Model(&persistence.CampaignRow{}).
		Where("advertiser_id = ? AND is_active = ?", advertiserID, true).
		Count(&n).Error
	return int(n), err
}

func (r *Repository) Get(ctx context.Context, id string) (*campaigndomain.Campaign, error) {
	var row persistence.CampaignRow
	if err := r.db.WithContext(ctx).Where("id = ?", id).First(&row).Error; err != nil {
		return nil, err
	}
	return row.ToDomain()
}

func (r *Repository) ListByAdvertiser(ctx context.Context, advertiserID string) ([]campaigndomain.Campaign, error) {
	var rows []persistence.CampaignRow
	err := r.db.WithContext(ctx).Where("advertiser_id = ?", advertiserID).Order("created_at desc").Find(&rows).Error
	if err != nil {
		return nil, err
	}
	return toDomainCampaigns(rows)
}

func (r *Repository) ListAll(ctx context.Context) ([]campaigndomain.Campaign, error) {
	var rows []persistence.CampaignRow
	err := r.db.WithContext(ctx).Order("created_at desc").Find(&rows).Error
	if err != nil {
		return nil, err
	}
	return toDomainCampaigns(rows)
}

func (r *Repository) Update(ctx context.Context, c *campaigndomain.Campaign) error {
	row, err := persistence.CampaignRowFromDomain(c)
	if err != nil {
		return err
	}
	return r.db.WithContext(ctx).Save(row).Error
}

// FindActiveForIntent returns the newest active campaign matching intent and channel code.
func (r *Repository) FindActiveForIntent(ctx context.Context, targetIntent, channelCode string) (*campaigndomain.Campaign, error) {
	var row persistence.CampaignRow
	q := r.db.WithContext(ctx).
		Table("campaigns").
		Select("campaigns.*").
		Joins("JOIN channels ON channels.id = campaigns.channel_id").
		Where("campaigns.target_intent = ? AND campaigns.is_active = ? AND campaigns.validation_status = ? AND campaigns.moderation_status = ?",
			targetIntent, true, "passed", campaigndomain.ModerationApproved)
	if channelCode != "" {
		q = q.Where("channels.code = ?", channelCode)
	}
	if err := q.Order("campaigns.created_at desc").First(&row).Error; err != nil {
		return nil, err
	}
	return row.ToDomain()
}

// ListEligibleForDelivery returns active campaigns for an intent/channel without SQL ordering.
// Plan-tier ranking happens in Go memory inside CachedCampaignRepository.
func (r *Repository) ListEligibleForDelivery(
	ctx context.Context,
	intentName, channelCode string,
) ([]campaigndomain.Campaign, error) {
	now := time.Now().UTC()
	var rows []eligibleCampaignScan
	q := r.db.WithContext(ctx).
		Table("campaigns").
		Select(`campaigns.*,
			sp.id AS plan_id,
			sp.name AS plan_name,
			sp.monthly_fee_etb AS plan_monthly_fee_etb,
			channels.code AS channel_code`).
		Joins("JOIN channels ON channels.id = campaigns.channel_id").
		Joins(`JOIN advertiser_subscriptions sub ON sub.advertiser_id = campaigns.advertiser_id AND sub.status = 'active'`).
		Joins(`JOIN subscription_plans sp ON sp.id = sub.plan_id AND sp.is_active = true`).
		Where(`campaigns.target_intent = ? AND campaigns.is_active = ? AND campaigns.validation_status = ? AND campaigns.moderation_status = ?
			AND sub.current_period_start <= ? AND sub.current_period_end >= ?`,
			intentName, true, "passed", campaigndomain.ModerationApproved, now, now)
	if channelCode != "" {
		q = q.Where("channels.code = ?", channelCode)
	}
	if err := q.Find(&rows).Error; err != nil {
		return nil, err
	}
	return toEligibleDomainCampaigns(rows)
}

// ListActiveMaster returns all active, approved campaigns with live subscriptions (all intents).
func (r *Repository) ListActiveMaster(ctx context.Context) ([]campaigndomain.Campaign, error) {
	now := time.Now().UTC()
	var rows []eligibleCampaignScan
	err := r.db.WithContext(ctx).
		Table("campaigns").
		Select(`campaigns.*,
			sp.id AS plan_id,
			sp.name AS plan_name,
			sp.monthly_fee_etb AS plan_monthly_fee_etb,
			channels.code AS channel_code`).
		Joins("JOIN channels ON channels.id = campaigns.channel_id").
		Joins(`JOIN advertiser_subscriptions sub ON sub.advertiser_id = campaigns.advertiser_id AND sub.status = 'active'`).
		Joins(`JOIN subscription_plans sp ON sp.id = sub.plan_id AND sp.is_active = true`).
		Where(`campaigns.is_active = ? AND campaigns.validation_status = ? AND campaigns.moderation_status = ?
			AND sub.current_period_start <= ? AND sub.current_period_end >= ?`,
			true, "passed", campaigndomain.ModerationApproved, now, now).
		Find(&rows).Error
	if err != nil {
		return nil, err
	}
	return toEligibleDomainCampaigns(rows)
}

// ListActiveByIntent returns approved active campaigns matching a target intent.
func (r *Repository) ListActiveByIntent(ctx context.Context, intentName string) ([]campaigndomain.Campaign, error) {
	return r.ListEligibleForDelivery(ctx, intentName, "")
}

// func (r *Repository) LogDelivery(ctx context.Context, log *campaigndomain.DeliveryLog) error {
// 	row := persistence.DeliveryLogRowFromDomain(log)
// 	return r.db.WithContext(ctx).Create(row).Error
// }

// CampaignAdContent builds the SDK ad payload from a campaign.
//
// The advertiser's destination is passed through verbatim unless it is genuinely
// a Google Play listing, in which case a signed `referrer` is attached so an
// install can be attributed back to the campaign (see
// delivery/worker.validateInstallToken, the consumer side of that token).
//
// Wrapping anything else is a bug the user sees immediately: the SDK launches
// destination_url with LaunchMode.externalApplication, so a web landing page
// rewritten into a play.google.com link opens the Play Store instead of the
// browser.
func CampaignAdContent(c *campaigndomain.Campaign, channelCode string, linkBuilder ...*PlayLinkBuilder) (map[string]any, error) {
	destination := c.DestinationURL

	if packageID := playStorePackageID(destination); packageID != "" && c.ID != "" {
		if builder := resolvePlayLinkBuilder(linkBuilder); builder != nil {
			destination = builder.BuildConsentedInstallURL(packageID, c.ID)
		}
	}
	canvas := c.CanvasJSON
	if canvas == nil {
		canvas = map[string]any{}
	}
	content := map[string]any{
		"title":           c.Title,
		"body_text":       c.BodyText,
		"image_url":       c.ImageURL,
		"destination_url": destination,
		"channel_code":    channelCode,
		"canvas_json":     canvas,
	}
	return content, nil
}

// playStorePackageID returns the Android package id when raw is a Google Play
// listing — https://play.google.com/store/apps/details?id=… or
// market://details?id=… — and "" for everything else: web landing pages, deep
// links into the host app, other app stores.
//
// Matching is deliberately narrow on the host. A suffix check alone would accept
// a lookalike like play.google.com.attacker.net, so only the exact host and its
// regional subdomains (www./country prefixes) qualify.
func playStorePackageID(raw string) string {
	if raw == "" {
		return ""
	}
	parsed, err := url.Parse(raw)
	if err != nil {
		return ""
	}

	switch strings.ToLower(parsed.Scheme) {
	case "market":
		if !strings.EqualFold(parsed.Host, "details") {
			return ""
		}
	case "http", "https":
		host := strings.ToLower(parsed.Hostname())
		if host != "play.google.com" && !strings.HasSuffix(host, ".play.google.com") {
			return ""
		}
		// Tolerate a trailing slash; reject any other path.
		if strings.Trim(parsed.Path, "/") != "store/apps/details" {
			return ""
		}
	default:
		return ""
	}

	return parsed.Query().Get("id")
}

func resolvePlayLinkBuilder(linkBuilders []*PlayLinkBuilder) *PlayLinkBuilder {
	if len(linkBuilders) > 0 && linkBuilders[0] != nil {
		return linkBuilders[0]
	}
	secret := strings.TrimSpace(os.Getenv("CLICK_TOKEN_SECRET"))
	if secret == "" {
		return nil
	}
	return NewPlayLinkBuilder(secret)
}

func toEligibleDomainCampaigns(rows []eligibleCampaignScan) ([]campaigndomain.Campaign, error) {
	out := make([]campaigndomain.Campaign, len(rows))
	for i := range rows {
		c, err := rows[i].CampaignRow.ToDomain()
		if err != nil {
			return nil, err
		}
		c.PlanID = rows[i].PlanID
		c.PlanName = rows[i].PlanName
		c.PlanMonthlyFeeETB = rows[i].PlanMonthlyFee
		c.ChannelCode = rows[i].ChannelCode
		out[i] = *c
	}
	return out, nil
}

func toDomainCampaigns(rows []persistence.CampaignRow) ([]campaigndomain.Campaign, error) {
	out := make([]campaigndomain.Campaign, len(rows))
	for i := range rows {
		c, err := rows[i].ToDomain()
		if err != nil {
			return nil, fmt.Errorf("map campaign row: %w", err)
		}
		out[i] = *c
	}
	return out, nil
}
