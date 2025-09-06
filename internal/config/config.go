package config

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	Port              string
	GCSProjectID      string
	GCSBucketName     string
	GCSServiceAccount string
	FrontEndURL       string
	JWTPublicKey      string
}

func Load() *Config {
	if os.Getenv("GO_ENV") != "production" {
		err := godotenv.Load()
		if err != nil {
			log.Println("Warning: Could not find .env file, using environment variables.")
		}
	}

	return &Config{
		Port:              getEnv("PORT", "4001"),
		GCSProjectID:      getEnv("GCS_PROJECT_ID", ""),
		GCSBucketName:     getEnv("GCS_BUCKET_NAME", ""),
		GCSServiceAccount: getEnv("GCS_SERVICE_ACCOUNT_FILE", ""),
		FrontEndURL:       getEnv("FRONT_END_URL", "http://localhost:5173"),
		JWTPublicKey:      getEnv("JWT_PUBLIC_KEY", ""),
	}
}

func getEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}