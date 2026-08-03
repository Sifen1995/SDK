package infrastructure

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"log/slog"
	"math"
	"time"

	frauddomain "skykin-platform/internal/fraud/domain"
	platformredis "skykin-platform/internal/platform/redis"
)

const (
	reportThreshold = int64(10)
	reportWindow    = time.Hour
	reportWindowTTL = 2 * time.Hour
)

type reportRedis interface {
	BRPop(ctx context.Context, key string, timeout time.Duration) (string, error)
	ZAdd(ctx context.Context, key, member string, score float64) error
	ZRemRangeByScore(ctx context.Context, key string, min, max float64) error
	ZCard(ctx context.Context, key string) (int64, error)
	Expire(ctx context.Context, key string, ttl time.Duration) error
}

func StartThreatReportWorker(
	rdb *platformredis.RedisClient,
	reports frauddomain.ThreatReportRepository,
	promotions frauddomain.PromotionRepository,
	logger *slog.Logger,
) {
	if rdb == nil || reports == nil || promotions == nil {
		return
	}
	if logger == nil {
		logger = slog.Default()
	}
	go runThreatReportWorker(
		context.Background(),
		rdb,
		reports,
		promotions,
		logger,
		time.Now,
	)
}

func runThreatReportWorker(
	ctx context.Context,
	rdb reportRedis,
	reports frauddomain.ThreatReportRepository,
	promotions frauddomain.PromotionRepository,
	logger *slog.Logger,
	now func() time.Time,
) {
	for {
		if ctx.Err() != nil {
			return
		}
		raw, err := rdb.BRPop(ctx, ThreatReportQueueKey, 2*time.Second)
		if err != nil || raw == "" {
			continue
		}
		report, err := DecodeThreatReportQueuePayload(raw)
		if err != nil {
			logger.Warn("fraud report worker: invalid payload", "error", err)
			continue
		}
		if err := processThreatReport(
			ctx,
			rdb,
			reports,
			promotions,
			report,
			now().UTC(),
		); err != nil {
			logger.Error("fraud report worker: aggregation failed",
				"report_id", report.ID,
				"error", err,
			)
		}
	}
}

func processThreatReport(
	ctx context.Context,
	rdb reportRedis,
	reports frauddomain.ThreatReportRepository,
	promotions frauddomain.PromotionRepository,
	report *frauddomain.ThreatReport,
	now time.Time,
) error {
	var result error
	if report.URLDomain != nil {
		result = errors.Join(result, processIndicator(
			ctx, rdb, reports, promotions, report,
			"domain", *report.URLDomain, now,
		))
	}
	if report.SenderHash != nil {
		result = errors.Join(result, processIndicator(
			ctx, rdb, reports, promotions, report,
			"sender", *report.SenderHash, now,
		))
	}
	return result
}

func processIndicator(
	ctx context.Context,
	rdb reportRedis,
	reports frauddomain.ThreatReportRepository,
	promotions frauddomain.PromotionRepository,
	report *frauddomain.ThreatReport,
	kind, indicator string,
	now time.Time,
) error {
	windowKey := indicatorRedisKey("fraud:reports", kind, indicator)
	score := float64(report.ReportedAt.UnixMilli())
	cutoff := float64(now.Add(-reportWindow).UnixMilli())

	if err := rdb.ZAdd(ctx, windowKey, report.ID, score); err != nil {
		return err
	}
	if err := rdb.ZRemRangeByScore(ctx, windowKey, -math.MaxFloat64, cutoff); err != nil {
		return err
	}
	count, err := rdb.ZCard(ctx, windowKey)
	if err != nil {
		return err
	}
	if err := rdb.Expire(ctx, windowKey, reportWindowTTL); err != nil {
		return err
	}
	if count <= reportThreshold {
		return nil
	}

	windowStart := now.Add(-reportWindow)
	var severity string
	if kind == "domain" {
		severity, err = reports.HighestSeverityForDomain(ctx, indicator, windowStart)
		if err == nil {
			err = promotions.PromoteDomain(
				ctx, indicator, report.ThreatType, severity, now,
			)
		}
	} else {
		severity, err = reports.HighestSeverityForSender(ctx, indicator, windowStart)
		if err == nil {
			err = promotions.PromoteSender(
				ctx, indicator, report.ThreatType, severity, now,
			)
		}
	}
	return err
}

func indicatorRedisKey(prefix, kind, indicator string) string {
	sum := sha256.Sum256([]byte(indicator))
	return prefix + ":" + kind + ":" + hex.EncodeToString(sum[:])
}
