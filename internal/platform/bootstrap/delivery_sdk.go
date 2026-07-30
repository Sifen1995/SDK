package bootstrap

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"skykin-platform/configs"
	billingApp "skykin-platform/internal/billing/application"
	billingInfra "skykin-platform/internal/billing/infrastructure"
	billingWorker "skykin-platform/internal/billing/worker"
	campaignApp "skykin-platform/internal/campaigns/application"
	campaigndomain "skykin-platform/internal/campaigns/domain"
	campaignInfra "skykin-platform/internal/campaigns/infrastructure"
	deliveryApp "skykin-platform/internal/delivery/application"
	deliveryConsumers "skykin-platform/internal/delivery/consumers"
	deliveryHTTP "skykin-platform/internal/delivery/http"
	deliveryInfra "skykin-platform/internal/delivery/infrastructure"
	deliveryWorker "skykin-platform/internal/delivery/worker"
	"skykin-platform/internal/platform/messaging"
	platformredis "skykin-platform/internal/platform/redis"

	"gorm.io/gorm"
)

// DeliverySDKHandlers holds SDK delivery + telemetry HTTP handlers.
type DeliverySDKHandlers struct {
	Campaigns   *deliveryHTTP.CampaignHandler
	Telemetry   *deliveryHTTP.TelemetryHandler
	CPC         *deliveryHTTP.CPCClickHandler
	SMSClick    *deliveryHTTP.SMSClickHandler
	Twilio      *deliveryHTTP.TwilioWebhookHandler
	SMSDebug    *deliveryHTTP.SMSDebugHandler
	SMSDispatch *deliveryApp.SMSDispatchService
}

// NewDeliverySDKSystem wires anonymous campaigns and telemetry track ingest.
func NewDeliverySDKSystem(db *gorm.DB, cfg *configs.Config, logger *slog.Logger) *DeliverySDKHandlers {
	if logger == nil {
		logger = slog.Default()
	}

	campaignRepo := campaignInfra.NewRepository(db)

	var platformRDB *platformredis.RedisClient
	var redisCampaign *campaignInfra.RedisCampaignRepository
	if addr := strings.TrimSpace(cfg.RedisAddr); addr != "" {
		if rdb, err := platformredis.NewRedisClient(addr); err == nil {
			platformRDB = rdb
			redisCampaign = campaignInfra.NewRedisCampaignRepository(rdb)
			logger.Info("delivery sdk: redis enabled", "addr", addr)
		} else {
			logger.Warn("delivery sdk: redis unavailable", "error", err)
		}
	}

	cached := campaignInfra.NewCachedCampaignRepository(campaignRepo, redisCampaign, platformRDB)
	secretKey := cfg.ClickTokenSecret
	anonSvc := campaignApp.NewAnonymousCampaignService(cached, secretKey)
	sendAttempts := deliveryInfra.NewSMSSendAttemptRepository(db)
	recipients := deliveryInfra.NewDemoSMSRecipientRepository(db)
	logs := deliveryInfra.NewDeliveryLogRepository(db)
	campaigns := &smsCampaignReader{repo: campaignRepo}
	smsProvider := buildSMSProvider(cfg)
	var smsDispatch *deliveryApp.SMSDispatchService
	if strings.TrimSpace(cfg.SMSClickSecret) != "" {
		smsDispatch = deliveryApp.NewSMSDispatchService(
			campaigns,
			recipients,
			sendAttempts,
			logs,
			smsProvider,
			cfg.SMSBaseURL,
			cfg.SMSClickSecret,
		)
	}

	out := &DeliverySDKHandlers{
		Campaigns:   deliveryHTTP.NewCampaignHandler(anonSvc),
		SMSDebug:    deliveryHTTP.NewSMSDebugHandler(sendAttempts),
		SMSDispatch: smsDispatch,
	}
	if platformRDB != nil {
		out.Telemetry = deliveryHTTP.NewTelemetryHandler(deliveryApp.NewTelemetryIngestService(platformRDB))
		if strings.TrimSpace(secretKey) != "" {
			cpcService := deliveryApp.NewCPCClickService(secretKey, platformRDB)
			out.CPC = deliveryHTTP.NewCPCClickHandler(cpcService)
		} else {
			logger.Warn("delivery sdk: anonymous CPC click handler disabled (click token secret required)")
		}
		if smsDispatch != nil {
			out.SMSClick = deliveryHTTP.NewSMSClickHandler(deliveryApp.NewSMSClickService(sendAttempts, platformRDB))
			if strings.EqualFold(cfg.SMSProvider, "twilio") && strings.TrimSpace(cfg.TwilioAuthToken) != "" {
				out.Twilio = deliveryHTTP.NewTwilioWebhookHandler(
					deliveryApp.NewTwilioStatusIngestService(sendAttempts, cfg.TwilioAuthToken),
				)
			}
		} else {
			logger.Warn("delivery sdk: sms click tracking disabled (sms click secret required)")
		}
	} else {
		logger.Warn("delivery sdk: telemetry track disabled (redis required)")
	}
	return out
}

func RegisterDeliveryEventConsumers(
	bus *messaging.Bus,
	dispatch *deliveryApp.SMSDispatchService,
	logger *slog.Logger,
) {
	if bus == nil || dispatch == nil {
		return
	}
	deliveryConsumers.NewSMSPlusConsumer(dispatch, logger).Register(bus)
}

// StartBillingStreamWorker launches the Redis Streams write-behind billing consumer.
func StartBillingStreamWorker(db *gorm.DB, cfg *configs.Config, logger *slog.Logger) {
	if logger == nil {
		logger = slog.Default()
	}
	addr := strings.TrimSpace(cfg.RedisAddr)
	if addr == "" {
		return
	}
	rdb, err := platformredis.NewRedisClient(addr)
	if err != nil {
		logger.Warn("billing stream worker: redis unavailable", "error", err)
		return
	}
	campaignReader := &campaignBillingReader{repo: campaignInfra.NewRepository(db)}
	subscriptions := billingInfra.NewSubscriptionRepository(db)
	rates := billingInfra.NewBillingRateRepository(db)
	events := billingInfra.NewBillingEventRepository(db)
	spend := &redisDailySpendTracker{rdb: rdb}
	marker := &budgetExhaustionMarker{flags: campaignInfra.NewRedisCampaignRepository(rdb)}
	processor := billingApp.NewEventProcessor(campaignReader, subscriptions, rates, events, spend, marker)
	billingWorker.StartBillingConsumer(processor, rdb, logger)
}

// StartDeliveryLogStreamWorker launches the delivery-module consumer for campaign_delivery_logs.
func StartDeliveryLogStreamWorker(db *gorm.DB, cfg *configs.Config, logger *slog.Logger) {
	if logger == nil {
		logger = slog.Default()
	}
	addr := strings.TrimSpace(cfg.RedisAddr)
	if addr == "" {
		return
	}
	rdb, err := platformredis.NewRedisClient(addr)
	if err != nil {
		logger.Warn("delivery log stream worker: redis unavailable", "error", err)
		return
	}
	deliveryWorker.StartDeliveryLogConsumer(db, rdb, logger)
}

type campaignBillingReader struct {
	repo *campaignInfra.Repository
}

func (r *campaignBillingReader) GetCampaignBillingInfo(
	ctx context.Context,
	campaignID string,
) (*billingApp.CampaignBillingInfo, error) {
	campaign, err := r.repo.Get(ctx, campaignID)
	if err != nil {
		return nil, err
	}
	if campaign == nil {
		return nil, fmt.Errorf("campaign not found: %s", campaignID)
	}
	return &billingApp.CampaignBillingInfo{
		ID:               campaign.ID,
		AdvertiserID:     campaign.AdvertiserID,
		DailyBudgetCap:   campaign.DailyBudgetCap,
		TotalBudgetCap:   campaign.TotalBudgetCap,
		CurrentBudgetUse: campaign.BudgetSpent,
	}, nil
}

type redisDailySpendTracker struct {
	rdb *platformredis.RedisClient
}

func (t *redisDailySpendTracker) Add(
	ctx context.Context,
	campaignID string,
	amount float64,
	ttl time.Duration,
) (float64, error) {
	today := time.Now().UTC().Format("2006-01-02")
	key := fmt.Sprintf("budget:spent:%s:%s", campaignID, today)
	spent, err := t.rdb.IncrByFloat(ctx, key, amount)
	if err != nil {
		return 0, err
	}
	if ttl > 0 {
		_ = t.rdb.Expire(ctx, key, ttl)
	}
	return spent, nil
}

type budgetExhaustionMarker struct {
	flags *campaignInfra.RedisCampaignRepository
}

func (m *budgetExhaustionMarker) MarkExhausted(ctx context.Context, campaignID string) error {
	return m.flags.SetBudgetExhausted(ctx, campaignID, 0)
}

type smsCampaignReader struct {
	repo campaigndomain.CampaignRepository
}

func (r *smsCampaignReader) GetSMSCampaign(
	ctx context.Context,
	campaignID string,
) (*deliveryApp.SMSCampaign, error) {
	campaign, err := r.repo.Get(ctx, campaignID)
	if err != nil {
		return nil, err
	}
	if campaign == nil {
		return nil, fmt.Errorf("campaign not found: %s", campaignID)
	}
	return &deliveryApp.SMSCampaign{
		ID:             campaign.ID,
		Title:          campaign.Title,
		BodyText:       campaign.BodyText,
		DestinationURL: campaign.DestinationURL,
	}, nil
}

func buildSMSProvider(cfg *configs.Config) deliveryApp.SMSProvider {
	if strings.EqualFold(strings.TrimSpace(cfg.SMSProvider), "twilio") {
		return deliveryInfra.NewTwilioSMSProvider(
			cfg.TwilioAccountSID,
			cfg.TwilioAuthToken,
			cfg.TwilioFromNumber,
			cfg.TwilioMessagingServiceSID,
			&http.Client{Timeout: 15 * time.Second},
		)
	}
	return deliveryInfra.NewMockSMSProvider("")
}
