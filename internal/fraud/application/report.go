package application

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/url"
	"regexp"
	"strings"
	"time"

	frauddomain "skykin-platform/internal/fraud/domain"

	"github.com/google/uuid"
)

var (
	ErrInvalidReport      = errors.New("invalid threat report")
	ErrQueueUnavailable   = errors.New("threat report aggregation queue unavailable")
	senderHashPattern     = regexp.MustCompile(`^[0-9a-f]{64}$`)
	domainLabelPattern    = regexp.MustCompile(`^[a-z0-9](?:[a-z0-9-]{0,61}[a-z0-9])?$`)
	validThreatTypes      = map[string]struct{}{"url_phishing": {}, "financial_scam": {}, "brand_impersonation": {}}
	validSeverities       = map[string]struct{}{"low": {}, "medium": {}, "high": {}, "critical": {}}
	validDetectionSources = map[string]struct{}{"blocklist": {}, "pattern": {}, "ml": {}}
)

type ThreatReportQueue interface {
	Enqueue(ctx context.Context, report *frauddomain.ThreatReport) error
}

type IngestReportCommand struct {
	ThreatType      string
	Severity        string
	SenderHash      string
	URLDomain       string
	DetectionSource string
	SDKVersion      string
}

type IngestReportResult struct {
	ReportID   string
	ReportedAt time.Time
}

type IngestReportUseCase struct {
	repository frauddomain.ThreatReportRepository
	queue      ThreatReportQueue
	now        func() time.Time
	newID      func() string
}

func NewIngestReportUseCase(
	repository frauddomain.ThreatReportRepository,
	queue ThreatReportQueue,
) *IngestReportUseCase {
	return &IngestReportUseCase{
		repository: repository,
		queue:      queue,
		now:        time.Now,
		newID:      uuid.NewString,
	}
}

func (uc *IngestReportUseCase) Execute(
	ctx context.Context,
	command IngestReportCommand,
) (*IngestReportResult, error) {
	report, err := buildThreatReport(command)
	if err != nil {
		return nil, err
	}
	report.ID = uc.newID()
	report.ReportedAt = uc.now().UTC()

	if err := uc.repository.Create(ctx, report); err != nil {
		return nil, fmt.Errorf("persist threat report: %w", err)
	}

	result := &IngestReportResult{ReportID: report.ID, ReportedAt: report.ReportedAt}
	if uc.queue == nil {
		return result, ErrQueueUnavailable
	}
	if err := uc.queue.Enqueue(ctx, report); err != nil {
		return result, fmt.Errorf("%w: %v", ErrQueueUnavailable, err)
	}
	return result, nil
}

func buildThreatReport(command IngestReportCommand) (*frauddomain.ThreatReport, error) {
	threatType := strings.ToLower(strings.TrimSpace(command.ThreatType))
	severity := strings.ToLower(strings.TrimSpace(command.Severity))
	detectionSource := strings.ToLower(strings.TrimSpace(command.DetectionSource))
	sdkVersion := strings.TrimSpace(command.SDKVersion)

	if _, ok := validThreatTypes[threatType]; !ok {
		return nil, fmt.Errorf("%w: unsupported threat_type", ErrInvalidReport)
	}
	if _, ok := validSeverities[severity]; !ok {
		return nil, fmt.Errorf("%w: unsupported severity", ErrInvalidReport)
	}
	if _, ok := validDetectionSources[detectionSource]; !ok {
		return nil, fmt.Errorf("%w: unsupported detection_source", ErrInvalidReport)
	}
	if sdkVersion == "" || len(sdkVersion) > 32 {
		return nil, fmt.Errorf("%w: sdk_version is required and must be at most 32 characters", ErrInvalidReport)
	}

	var senderHash *string
	if raw := strings.TrimSpace(command.SenderHash); raw != "" {
		if !senderHashPattern.MatchString(raw) {
			return nil, fmt.Errorf("%w: sender_hash must be lowercase SHA-256 hex", ErrInvalidReport)
		}
		senderHash = &raw
	}

	var urlDomain *string
	if raw := strings.TrimSpace(command.URLDomain); raw != "" {
		normalized, err := normalizeDomain(raw)
		if err != nil {
			return nil, fmt.Errorf("%w: %v", ErrInvalidReport, err)
		}
		urlDomain = &normalized
	}
	if senderHash == nil && urlDomain == nil {
		return nil, fmt.Errorf("%w: sender_hash or url_domain is required", ErrInvalidReport)
	}

	return &frauddomain.ThreatReport{
		ThreatType:      threatType,
		Severity:        severity,
		SenderHash:      senderHash,
		URLDomain:       urlDomain,
		DetectionSource: detectionSource,
		SDKVersion:      sdkVersion,
	}, nil
}

func normalizeDomain(raw string) (string, error) {
	candidate := strings.TrimSpace(raw)
	if candidate == "" {
		return "", errors.New("url_domain is required")
	}

	var parsed *url.URL
	var err error
	if strings.Contains(candidate, "://") {
		parsed, err = url.Parse(candidate)
	} else {
		parsed, err = url.Parse("//" + candidate)
	}
	if err != nil || parsed == nil || parsed.Hostname() == "" {
		return "", errors.New("url_domain must contain a valid host")
	}
	if parsed.User != nil {
		return "", errors.New("url_domain must not contain credentials")
	}

	host := strings.TrimSuffix(strings.ToLower(parsed.Hostname()), ".")
	if len(host) == 0 || len(host) > 253 {
		return "", errors.New("url_domain host length is invalid")
	}
	if net.ParseIP(host) != nil {
		return host, nil
	}
	for _, label := range strings.Split(host, ".") {
		if !domainLabelPattern.MatchString(label) {
			return "", errors.New("url_domain contains an invalid host label")
		}
	}
	return host, nil
}
