package domain

import "time"

// User is an SDK end-user. Identity linkage to Flutter lives in
// consent.pseudonymous_mappings — never on this row.
type User struct {
	ID        int64
	CreatedAt time.Time
}
