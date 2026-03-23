package config

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	ServerPort string
	DSN        string
}

func LoadConfig() *Config {
	err := godotenv.Load()
	if err != nil {
		log.Println("No .env file found, relying on environment variables.")
	}

	return &Config{
		ServerPort: getEnv("SERVER_PORT", "8080"),
		DSN:        getEnv("DB_DSN", "root:rootpass@tcp(127.0.0.1:3306)/taipflow?charset=utf8mb4&parseTime=True&loc=Local"),
	}
}

func getEnv(key string, fallback string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return fallback
}
