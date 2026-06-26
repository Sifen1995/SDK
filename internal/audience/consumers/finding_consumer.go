package consumers

import (
	"context"
	"log/slog"

	analyticsdomain "skykin-platform/internal/analytics/domain"
	audienceApp "skykin-platform/internal/audience/application"
	"skykin-platform/internal/platform/messaging"
)

// FindingConsumer persists segment candidates when analytics publishes intent consistency findings.
type FindingConsumer struct {
	save *audienceApp.SaveCandidateFromFindingUseCase
	log  *slog.Logger
}

func NewFindingConsumer(save *audienceApp.SaveCandidateFromFindingUseCase, log *slog.Logger) *FindingConsumer {
	if log == nil {
		log = slog.Default()
	}
	return &FindingConsumer{save: save, log: log}
}

func (c *FindingConsumer) Register(bus *messaging.Bus) {
	messaging.Register(bus, analyticsdomain.TopicIntentConsistencyFinding, c.handle)
}

func (c *FindingConsumer) handle(e messaging.Event) {
	finding, ok := e.Payload.(analyticsdomain.IntentConsistencyFinding)
	if !ok {
		c.log.Error("invalid intent consistency finding payload")
		return
	}
	ctx := e.Ctx
	if ctx == nil {
		ctx = context.Background()
	}
	if err := c.save.Execute(ctx, finding); err != nil {
		c.log.Error("save candidate from finding failed", "intent", finding.IntentName, "error", err)
	}
}
