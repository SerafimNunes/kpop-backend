package config

import (
	"os"
)

// Config holds runtime configuration values
type Config struct {
	Port         string
	DatabaseURL  string
	GeminiAPIKey string
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

// Load reads environment variables and returns a Config
func Load() Config {
	LoadEnv()

	return Config{
		Port:         getEnv("PORT", "8080"),
		DatabaseURL:  getEnv("DATABASE_URL", ""),
		GeminiAPIKey: getEnv("GEMINI_API_KEY", ""),
	}
}
