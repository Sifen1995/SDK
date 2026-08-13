package configs

import (
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	DBHost                    string
	DBPort                    string
	DBUser                    string
	DBPassword                string
	DBName                    string
	ClickTokenSecret          string
	SMSClickSecret            string
	SMSProvider               string
	SMSBaseURL                string
	TwilioAccountSID          string
	TwilioAuthToken           string
	TwilioFromNumber          string
	TwilioMessagingServiceSID string
	JwtSecret                 string
	Port                      string
	RedisAddr                 string // optional, e.g. redis:6379; empty uses in-memory dedup
	AdminEmail                string
	AdminPassword             string
}

func LoadConfig() (*Config, error) {
	// Load environment variables
	_ = godotenv.Load()

	port := os.Getenv("PORT")
	if port == "" {
		port = "8081"
	}
	return &Config{
		DBHost:                    os.Getenv("DB_HOST"),
		DBPort:                    os.Getenv("DB_PORT"),
		DBUser:                    os.Getenv("DB_USER"),
		DBPassword:                os.Getenv("DB_PASSWORD"),
		DBName:                    os.Getenv("DB_NAME"),
		ClickTokenSecret:          os.Getenv("CLICK_TOKEN_SECRET"),
		SMSClickSecret:            envOr("SMS_CLICK_SECRET", os.Getenv("CLICK_TOKEN_SECRET")),
		SMSProvider:               envOr("SMS_PROVIDER", "mock"),
		SMSBaseURL:                envOr("SMS_BASE_URL", "http://localhost:"+port),
		TwilioAccountSID:          os.Getenv("TWILIO_ACCOUNT_SID"),
		TwilioAuthToken:           os.Getenv("TWILIO_AUTH_TOKEN"),
		TwilioFromNumber:          os.Getenv("TWILIO_FROM_NUMBER"),
		TwilioMessagingServiceSID: os.Getenv("TWILIO_MESSAGING_SERVICE_SID"),
		JwtSecret:                 os.Getenv("JWT_SECRET"),
		Port:                      port,
		RedisAddr:                 os.Getenv("REDIS_ADDR"),
		AdminEmail:                envOr("ADMIN_EMAIL", "admin@skykin.com"),
		AdminPassword:             envOr("ADMIN_PASSWORD", "Admin12345!"),
	}, nil
}

func envOr(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
