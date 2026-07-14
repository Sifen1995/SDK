package validation

func ValidateConsentLevel(level string) bool {
	switch level {
	case "individual", "aggregate", "none":
		return true
	default:
		return false
	}
}
