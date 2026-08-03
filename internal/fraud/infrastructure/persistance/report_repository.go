package persistance

import (
	"context"
	"database/sql"
	"errors"
	"time"

	frauddomain "skykin-platform/internal/fraud/domain"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type ReportRepository struct {
	db *gorm.DB
}

func NewReportRepository(db *gorm.DB) *ReportRepository {
	return &ReportRepository{db: db}
}

var (
	_ frauddomain.ThreatReportRepository = (*ReportRepository)(nil)
	_ frauddomain.PromotionRepository    = (*ReportRepository)(nil)
)

func (r *ReportRepository) Create(ctx context.Context, report *frauddomain.ThreatReport) error {
	row := ThreatReportsRowFromDomain(*report)
	if err := r.db.WithContext(ctx).Create(&row).Error; err != nil {
		return err
	}
	*report = row.ToDomain()
	return nil
}

func (r *ReportRepository) HighestSeverityForDomain(
	ctx context.Context,
	domain string,
	since time.Time,
) (string, error) {
	return r.highestSeverity(ctx, "url_domain", domain, since)
}

func (r *ReportRepository) HighestSeverityForSender(
	ctx context.Context,
	senderHash string,
	since time.Time,
) (string, error) {
	return r.highestSeverity(ctx, "sender_hash", senderHash, since)
}

func (r *ReportRepository) highestSeverity(
	ctx context.Context,
	column, value string,
	since time.Time,
) (string, error) {
	var row ThreatReportsRow
	err := r.db.WithContext(ctx).
		Select("severity").
		Where(column+" = ? AND reported_at >= ?", value, since).
		Order(`CASE severity
			WHEN 'critical' THEN 4
			WHEN 'high' THEN 3
			WHEN 'medium' THEN 2
			WHEN 'low' THEN 1
			ELSE 0 END DESC`).
		First(&row).Error
	if err != nil {
		return "", err
	}
	return row.Severity, nil
}

func (r *ReportRepository) PromoteDomain(
	ctx context.Context,
	domain, threatType, severity string,
	now time.Time,
) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var row BlockedDomainsRow
		err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
			Where("domain = ?", domain).
			First(&row).Error
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return tx.Create(&BlockedDomainsRow{
				Domain:     domain,
				ThreatType: threatType,
				Severity:   severity,
				Source:     "community_report",
				Status:     frauddomain.StatusActive,
				CreatedAt:  now,
				ExpiresAt:  sql.NullTime{Time: now.AddDate(0, 0, 30), Valid: true},
				UpdatedAt:  now,
			}).Error
		}
		if err != nil {
			return err
		}
		row.ThreatType = threatType
		row.Source = "community_report"
		row.Status = frauddomain.StatusActive
		row.ExpiresAt = sql.NullTime{Time: now.AddDate(0, 0, 30), Valid: true}
		row.UpdatedAt = now
		if severityRank(severity) > severityRank(row.Severity) {
			row.Severity = severity
		}
		return tx.Save(&row).Error
	})
}

func (r *ReportRepository) PromoteSender(
	ctx context.Context,
	senderHash, threatType, severity string,
	now time.Time,
) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var row BlockedSendersRow
		err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
			Where("sender_hash = ?", senderHash).
			First(&row).Error
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return tx.Create(&BlockedSendersRow{
				SenderHash: senderHash,
				ThreatType: threatType,
				Severity:   severity,
				Source:     "community_report",
				Status:     frauddomain.StatusActive,
				CreatedAt:  now,
				UpdatedAt:  now,
			}).Error
		}
		if err != nil {
			return err
		}
		row.ThreatType = threatType
		row.Source = "community_report"
		row.Status = frauddomain.StatusActive
		row.UpdatedAt = now
		if severityRank(severity) > severityRank(row.Severity) {
			row.Severity = severity
		}
		return tx.Save(&row).Error
	})
}

func severityRank(value string) int {
	switch value {
	case "critical":
		return 4
	case "high":
		return 3
	case "medium":
		return 2
	case "low":
		return 1
	default:
		return 0
	}
}
