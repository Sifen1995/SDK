package validation

import (
	"errors"
	"strings"

	platformvalidation "skykin-platform/internal/platform/validation"
)

// DeveloperRegister validates developer portal registration.
func DeveloperRegister(name, email, password string) (string, error) {
	if strings.TrimSpace(name) == "" {
		return "", errors.New("name is required")
	}
	if len(strings.TrimSpace(name)) < 2 {
		return "", errors.New("name must be at least 2 characters")
	}
	normalizedEmail, err := platformvalidation.Email(email)
	if err != nil {
		return "", err
	}
	if err := platformvalidation.Password(password); err != nil {
		return "", err
	}
	return normalizedEmail, nil
}

// DeveloperLogin validates developer portal login credentials.
func DeveloperLogin(email, password string) (string, error) {
	return platformvalidation.LoginCredentials(email, password)
}
