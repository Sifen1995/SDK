package bootstrap

import (
	"log/slog"
	"strings"

	"skykin-platform/configs"
	fraudapp "skykin-platform/internal/fraud/application"
	fraudinfra "skykin-platform/internal/fraud/infrastructure"
	fraudpersistance "skykin-platform/internal/fraud/infrastructure/persistance"
	fraudHTTP "skykin-platform/internal/fraud/interfaces/http"
	platformredis "skykin-platform/internal/platform/redis"

	"gorm.io/gorm"
)

// NewFraudSystem wires mobile sync and anonymous threat-report ingestion.
func NewFraudSystem(
	db *gorm.DB,
	cfg *configs.Config,
	logger *slog.Logger,
) *fraudHTTP.Handler {
	if logger == nil {
		logger = slog.Default()
	}
	syncRepository := fraudpersistance.NewSyncRepository(db)
	syncUseCase := fraudapp.NewSyncUseCase(syncRepository)

	reportRepository := fraudpersistance.NewReportRepository(db)
	var queue *fraudinfra.ThreatReportQueue
	if cfg != nil {
		if addr := strings.TrimSpace(cfg.RedisAddr); addr != "" {
			if rdb, err := platformredis.NewRedisClient(addr); err == nil {
				queue = fraudinfra.NewThreatReportQueue(rdb)
				fraudinfra.StartThreatReportWorker(
					rdb,
					reportRepository,
					reportRepository,
					logger,
				)
				logger.Info("fraud report aggregation: redis enabled", "addr", addr)
			} else {
				logger.Warn("fraud report aggregation: redis unavailable", "error", err)
			}
		}
	}

	reportUseCase := fraudapp.NewIngestReportUseCase(reportRepository, queue)
	return fraudHTTP.NewHandler(syncUseCase, reportUseCase)
}
