package configs

import (
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	DBHost       string
	DBPort       string
	DBUser       string
	DBPassword   string
	DBName       string
	JwtSecret    string
	Port         string
	MLServiceURL string
	JwtSecretKey string
}

func LoadConfig() (*Config, error) {
	// Load environment variables
	_ = godotenv.Load()

	port := os.Getenv("PORT")
	if port == "" {
		port = "8081"
	}
	return &Config{
		DBHost:       os.Getenv("DB_HOST"),
		DBPort:       os.Getenv("DB_PORT"),
		DBUser:       os.Getenv("DB_USER"),
		DBPassword:   os.Getenv("DB_PASSWORD"),
		DBName:       os.Getenv("DB_NAME"),
		JwtSecret:    os.Getenv("JWT_SECRET"),
		Port:         port,
		MLServiceURL: os.Getenv("ML_SERVICE_URL"),
		JwtSecretKey: os.Getenv("JWT_SECRET_KEY"),
	}, nil
}
