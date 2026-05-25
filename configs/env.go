package configs

import "os"

// GetEnv retrieves the value of the environment variable named by the key.
// It returns a default value if the variable is not present.
func GetEnv(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}
