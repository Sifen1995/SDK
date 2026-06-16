package application

import (
	"context"
	"time"

	analyticsdomain "skykin-platform/internal/analytics/domain"
)

// Service provides read-only analytics for the operator admin console.
type Service struct {
	repo analyticsdomain.AnalyticsRepository
}

func NewService(repo analyticsdomain.AnalyticsRepository) *Service {
	return &Service{repo: repo}
}

func (s *Service) Overview(ctx context.Context) (*analyticsdomain.OverviewStats, error) {
	return s.repo.Overview(ctx)
}

func (s *Service) Campaigns(ctx context.Context) ([]analyticsdomain.CampaignPerformance, error) {
	return s.repo.CampaignPerformance(ctx)
}

func (s *Service) CampaignDetail(ctx context.Context, id string) (*analyticsdomain.CampaignDetail, error) {
	return s.repo.CampaignDetail(ctx, id)
}

func (s *Service) Delivery(ctx context.Context) (*analyticsdomain.DeliveryAnalytics, error) {
	since := time.Now().UTC().AddDate(0, 0, -30)
	return s.repo.DeliveryAnalytics(ctx, since)
}

func (s *Service) Revenue(ctx context.Context) (*analyticsdomain.RevenueOverview, error) {
	since := time.Now().UTC().AddDate(0, 0, -30)
	return s.repo.RevenueOverview(ctx, since)
}

func (s *Service) Advertisers(ctx context.Context) ([]analyticsdomain.AdvertiserSummary, error) {
	return s.repo.AdvertiserSummaries(ctx)
}
