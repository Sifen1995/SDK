package application

import (
	"context"
	"testing"
	"time"

	frauddomain "skykin-platform/internal/fraud/domain"

	"github.com/stretchr/testify/require"
)

type stubSyncRepository struct {
	sync func(context.Context, *time.Time, time.Time) (*frauddomain.SyncSnapshot, error)
}

func (s stubSyncRepository) Sync(
	ctx context.Context,
	since *time.Time,
	until time.Time,
) (*frauddomain.SyncSnapshot, error) {
	return s.sync(ctx, since, until)
}

func TestSyncUseCaseUsesSingleServerCursor(t *testing.T) {
	now := time.Date(2026, 8, 3, 8, 30, 0, 0, time.UTC)
	since := now.Add(-time.Hour)
	repository := stubSyncRepository{sync: func(
		_ context.Context,
		actualSince *time.Time,
		until time.Time,
	) (*frauddomain.SyncSnapshot, error) {
		require.NotNil(t, actualSince)
		require.Equal(t, since, *actualSince)
		require.Equal(t, now, until)
		return &frauddomain.SyncSnapshot{NextCursor: until, IsDelta: true}, nil
	}}
	useCase := NewSyncUseCase(repository)
	useCase.now = func() time.Time { return now }

	result, err := useCase.Execute(context.Background(), &since)
	require.NoError(t, err)
	require.Equal(t, now, result.NextCursor)
}

func TestSyncUseCaseRejectsFutureCursor(t *testing.T) {
	now := time.Date(2026, 8, 3, 8, 30, 0, 0, time.UTC)
	called := false
	repository := stubSyncRepository{sync: func(
		_ context.Context,
		_ *time.Time,
		_ time.Time,
	) (*frauddomain.SyncSnapshot, error) {
		called = true
		return nil, nil
	}}
	useCase := NewSyncUseCase(repository)
	useCase.now = func() time.Time { return now }
	future := now.Add(time.Second)

	_, err := useCase.Execute(context.Background(), &future)
	require.ErrorIs(t, err, ErrFutureCursor)
	require.False(t, called)
}
