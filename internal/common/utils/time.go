package utils

import "time"

// FormatISO8601 formats a time as ISO8601 string.
func FormatISO8601(t time.Time) string {
	return t.Format(time.RFC3339)
}
