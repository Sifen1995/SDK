package application

import "context"

// IsDuplicate checks whether an event_id was seen within the dedup window.
func IsDuplicate(ctx context.Context, store DedupStore, eventID string) (bool, error) {
	if store == nil {
		return false, nil
	}
	acquired, err := store.TryAcquire(ctx, eventID)
	if err != nil {
		return false, err
	}
	return !acquired, nil
}
