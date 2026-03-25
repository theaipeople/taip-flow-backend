package config

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	ServerPort      string
	DSN             string
	JWTSecret       string
	JWTAccessMins   int
	JWTRefreshDays  int
	CookieDomain    string
	CookieSecure    bool
	GoogleClientID  string
	GoogleClientSecret string
}

func LoadConfig() *Config {
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, relying on environment variables.")
	}

	return &Config{
		ServerPort:         getEnv("SERVER_PORT", "8080"),
		DSN:                getEnv("DB_DSN", "root:rootpass@tcp(127.0.0.1:3306)/taipflow?charset=utf8mb4&parseTime=True&loc=Local"),
		JWTSecret:          getEnv("JWT_SECRET", "change-me-in-production-use-32-chars-min"),
		JWTAccessMins:      15,
		JWTRefreshDays:     7,
		CookieDomain:       getEnv("COOKIE_DOMAIN", "localhost"),
		CookieSecure:       getEnv("COOKIE_SECURE", "false") == "true",
		GoogleClientID:     getEnv("GOOGLE_CLIENT_ID", ""),
		GoogleClientSecret: getEnv("GOOGLE_CLIENT_SECRET", ""),
	}
}

func getEnv(key, fallback string) string {
	if v, ok := os.LookupEnv(key); ok {
		return v
	}
	return fallback
}
