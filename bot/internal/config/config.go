package config

import (
	"log/slog"
	"os"
	"strings"
)

type Config struct {
	BotToken string
	HTTPAddr string
	LogLevel slog.Level
}

func Load() Config {
	return Config{
		BotToken: os.Getenv("BOT_TOKEN"),
		HTTPAddr: envOrDefault("HTTP_ADDR", ":8080"),
		LogLevel: parseLogLevel(os.Getenv("LOG_LEVEL")),
	}
}

func envOrDefault(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func parseLogLevel(raw string) slog.Level {
	switch strings.ToLower(strings.TrimSpace(raw)) {
	case "debug":
		return slog.LevelDebug
	case "warn", "warning":
		return slog.LevelWarn
	case "error":
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}
