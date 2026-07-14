package application

import (
	"context"
	"errors"
	"log/slog"
	"time"

	"skykin-platform/internal/consent/domain"
	"skykin-platform/internal/consent/validation"
	"skykin-platform/internal/platform/messaging"

	"github.com/google/uuid"
)

type consentWriter interface {
	Create(ctx context.Context, consent *domain.Consent) error
}

type mappingWriter interface {
	Create(ctx context.Context, mapping *domain.PseudonymousMapping) error
}

// CompleteConsentRegistrationUseCase persists mapping + consent after a user is provisioned.
type CompleteConsentRegistrationUseCase struct {
	consents consentWriter
	mappings mappingWriter
	bus      interface {
		Publish(event messaging.Event)
	}
	logger *slog.Logger
}

func NewCompleteConsentRegistrationUseCase(
	consents consentWriter,
	mappings mappingWriter,
	bus interface {
		Publish(event messaging.Event)
	},
	logger *slog.Logger,
) *CompleteConsentRegistrationUseCase {
	if logger == nil {
		logger = slog.Default()
	}
	return &CompleteConsentRegistrationUseCase{
		consents: consents,
		mappings: mappings,
		bus:      bus,
		logger:   logger,
	}
}

func (uc *CompleteConsentRegistrationUseCase) Execute(
	ctx context.Context,
	userID int64,
	pseudonymousID string,
	consentLevel string,
	sdkVersion string,
) error {
	if userID == 0 {
		return errors.New("user_id is required")
	}
	if !validation.ValidateConsentLevel(consentLevel) {
		return errors.New("invalid consent level: " + consentLevel)
	}
	pseudoUUID, err := uuid.Parse(pseudonymousID)
	if err != nil {
		return errors.New("invalid pseudonymous_id")
	}

	mapping := &domain.PseudonymousMapping{
		UserID:         userID,
		PseudonymousID: pseudoUUID,
	}
	if err := uc.mappings.Create(ctx, mapping); err != nil {
		return err
	}

	now := time.Now().UTC()
	consent := &domain.Consent{
		UserID:       userID,
		ConsentLevel: consentLevel,
		IsActive:     true,
		GrantedAt:    &now,
		SDKVersion:   sdkVersion,
	}
	if err := uc.consents.Create(ctx, consent); err != nil {
		return err
	}

	uc.bus.Publish(messaging.Event{
		Name: domain.EventConsentCreated,
		Payload: domain.ConsentCreatedPayload{
			ConsentID:      consent.ID,
			UserID:         userID,
			PseudonymousID: pseudonymousID,
			ConsentLevel:   consentLevel,
			SDKVersion:     sdkVersion,
		},
		Ctx: ctx,
	})

	uc.logger.Info("consent registration completed",
		"consent_id", consent.ID,
		"user_id", userID,
		"pseudonymous_id", pseudonymousID,
		"consent_level", consentLevel,
	)
	return nil
}
