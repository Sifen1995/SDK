package middleware

import "strings"

// bearerTokenFromHeader extracts a JWT from Authorization.
// Swagger UI (apiKey in Authorization) often sends the raw token without a "Bearer " prefix.
func bearerTokenFromHeader(authHeader string) (string, bool) {
	token := strings.TrimSpace(authHeader)
	for strings.HasPrefix(strings.ToLower(token), "bearer ") {
		token = strings.TrimSpace(token[7:])
	}
	return token, token != ""
}
