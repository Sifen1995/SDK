package events

import (
	"context"
	"log/slog"

	adminEvents "skykin-platform/internal/admin/events"
	audienceApp "skykin-platform/internal/audience/application"
	campaignEvents "skykin-platform/internal/campaigns/events"
	"skykin-platform/internal/platform/messaging"
)

// CandidateConsumer provisions audience data for admin and campaign lifecycle events.
type CandidateConsumer struct {
	approve         *audienceApp.ProcessApprovedCandidateUseCase
	reject          *audienceApp.RejectCandidateUseCase
	recordPurchase  *audienceApp.RecordSegmentPurchaseUseCase
	log             *slog.Logger
}

func NewCandidateConsumer(
	approve *audienceApp.ProcessApprovedCandidateUseCase,
	reject *audienceApp.RejectCandidateUseCase,
	recordPurchase *audienceApp.RecordSegmentPurchaseUseCase,
	log *slog.Logger,
) *CandidateConsumer {
	if log == nil {
		log = slog.Default()
	}
	return &CandidateConsumer{
		approve:        approve,
		reject:         reject,
		recordPurchase: recordPurchase,
		log:            log,
	}
}

func (c *CandidateConsumer) Register(bus *messaging.Bus) {
	messaging.Register(bus, adminEvents.TopicCandidateApproved, c.handleApproved)
	messaging.Register(bus, adminEvents.TopicCandidateRejected, c.handleRejected)
	messaging.Register(bus, campaignEvents.TopicCampaignCreated, c.handleCampaignCreated)
}

func (c *CandidateConsumer) handleApproved(e messaging.Event) {
	evt, ok := e.Payload.(adminEvents.CandidateApprovedEvent)
	if !ok {
		c.log.Error("invalid candidate approved payload")
		return
	}
	ctx := e.Ctx
	if ctx == nil {
		ctx = context.Background()
	}
	if err := c.approve.Execute(ctx, evt); err != nil {
		c.log.Error("process approved candidate failed", "candidate_id", evt.CandidateID, "error", err)
	}
}

func (c *CandidateConsumer) handleRejected(e messaging.Event) {
	evt, ok := e.Payload.(adminEvents.CandidateRejectedEvent)
	if !ok {
		c.log.Error("invalid candidate rejected payload")
		return
	}
	ctx := e.Ctx
	if ctx == nil {
		ctx = context.Background()
	}
	if err := c.reject.Execute(ctx, evt); err != nil {
		c.log.Error("reject candidate failed", "candidate_id", evt.CandidateID, "error", err)
	}
}

func (c *CandidateConsumer) handleCampaignCreated(e messaging.Event) {
	evt, ok := e.Payload.(campaignEvents.CampaignCreatedEvent)
	if !ok {
		c.log.Error("invalid campaign created payload")
		return
	}
	ctx := e.Ctx
	if ctx == nil {
		ctx = context.Background()
	}
	if err := c.recordPurchase.Execute(ctx, evt); err != nil {
		c.log.Error("record segment purchase failed", "campaign_id", evt.CampaignID, "error", err)
	}
}
