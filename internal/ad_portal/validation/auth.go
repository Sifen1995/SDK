package validation

import (
	"errors"
	"strings"

	platformvalidation "skykin-platform/internal/platform/validation"
)

// Register validates ad portal self-registration fields.
func Register(name, email, password string) (normalizedEmail string, err error) {
	if strings.TrimSpace(name) == "" {
		return "", errors.New("name is required")
	}
	normalizedEmail, err = platformvalidation.Email(email)
	if err != nil {
		return "", err
	}
	if err := platformvalidation.Password(password); err != nil {
		return "", err
	}
	return normalizedEmail, nil
}

// Login validates ad portal login credentials.
func Login(email, password string) (string, error) {
	return platformvalidation.LoginCredentials(email, password)
}

// CreateUser validates operator-created portal users.
func CreateUser(name, email, password, role, company string) (string, error) {
	if strings.TrimSpace(name) == "" {
		return "", errors.New("name is required")
	}
	normalizedEmail, err := platformvalidation.Email(email)
	if err != nil {
		return "", err
	}
	if err := platformvalidation.Password(password); err != nil {
		return "", err
	}
	switch role {
	case "advertiser", "read_only_analyst", "operator_admin":
	default:
		return "", errors.New("invalid role")
	}
	if role == "advertiser" && strings.TrimSpace(company) == "" {
		return "", errors.New("company_name is required")
	}
	return normalizedEmail, nil
}
