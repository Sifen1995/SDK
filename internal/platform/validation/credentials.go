package validation

import (
	"errors"
	"net/mail"
	"strings"
	"unicode"
)

const minPasswordLen = 8

// Email normalizes and validates an email address.
func Email(raw string) (string, error) {
	addr := strings.TrimSpace(strings.ToLower(raw))
	if addr == "" {
		return "", errors.New("email is required")
	}
	if _, err := mail.ParseAddress(addr); err != nil {
		return "", errors.New("email must be a valid address")
	}
	return addr, nil
}

// Password checks password strength for registration and account creation.
func Password(raw string) error {
	p := strings.TrimSpace(raw)
	if len(p) < minPasswordLen {
		return errors.New("password must be at least 8 characters")
	}
	var hasLetter, hasDigit bool
	for _, r := range p {
		switch {
		case unicode.IsLetter(r):
			hasLetter = true
		case unicode.IsDigit(r):
			hasDigit = true
		}
	}
	if !hasLetter || !hasDigit {
		return errors.New("password must include at least one letter and one number")
	}
	return nil
}

// LoginCredentials validates email format for login; password presence only.
func LoginCredentials(emailRaw, passwordRaw string) (string, error) {
	email, err := Email(emailRaw)
	if err != nil {
		return "", err
	}
	if strings.TrimSpace(passwordRaw) == "" {
		return "", errors.New("password is required")
	}
	return email, nil
}
