package middleware

import "testing"

func TestBearerTokenFromHeader(t *testing.T) {
	tests := []struct {
		header string
		want   string
		ok     bool
	}{
		{"Bearer eyJhbGciOiJIUzI1NiJ9", "eyJhbGciOiJIUzI1NiJ9", true},
		{"bearer eyJhbGciOiJIUzI1NiJ9", "eyJhbGciOiJIUzI1NiJ9", true},
		{"eyJhbGciOiJIUzI1NiJ9", "eyJhbGciOiJIUzI1NiJ9", true},
		{"Bearer Bearer eyJhbGciOiJIUzI1NiJ9", "eyJhbGciOiJIUzI1NiJ9", true},
		{"", "", false},
		{"Bearer ", "", false},
		{"Bearer", "", false},
	}
	for _, tt := range tests {
		got, ok := bearerTokenFromHeader(tt.header)
		if ok != tt.ok || got != tt.want {
			t.Errorf("bearerTokenFromHeader(%q) = %q, %v; want %q, %v", tt.header, got, ok, tt.want, tt.ok)
		}
	}
}
