package config

import (
	"os"
	"time"
	"github.com/joho/godotenv"
)

type Config struct {
	Port        string
	RedisAddr   string
	StreamName  string
	WorkerCount int
	WindowSize  time.Duration
}

func LoadConfig() *Config {
	_ = godotenv.Load() // Ignore error if .env doesn't exist

	windowSize, _ := time.ParseDuration(getEnv("WINDOW_SIZE", "30s"))

	return &Config{
		Port:        getEnv("PORT", "8080"),
		RedisAddr:   getEnv("REDIS_ADDR", "localhost:6379"),
		StreamName:  getEnv("STREAM_NAME", "events-stream"),
		WorkerCount: 5, // Simplified
		WindowSize:  windowSize,
	}
}

func getEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}
