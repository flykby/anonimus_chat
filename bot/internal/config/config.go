package config

import (
	"errors"
	"fmt"
	"log/slog"
	"os"
	"strings"
)

var ErrMissingBotToken = errors.New("BOT_TOKEN is required")

type Config struct {
	BotToken   string
	HTTPAddr   string
	LogLevel   slog.Level
	HealthOnly bool
}

func Load() Config {
	return Config{
		BotToken:   os.Getenv("BOT_TOKEN"),
		HTTPAddr:   envOrDefault("HTTP_ADDR", ":8080"),
		LogLevel:   parseLogLevel(os.Getenv("LOG_LEVEL")),
		HealthOnly: envBool("BOT_HEALTH_ONLY"),
	}
}

func (c Config) Validate() error {
	if c.HealthOnly {
		return nil
	}
	if strings.TrimSpace(c.BotToken) == "" {
		return ErrMissingBotToken
	}
	return nil
}

func envOrDefault(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func envBool(key string) bool {
	switch strings.ToLower(strings.TrimSpace(os.Getenv(key))) {
	case "1", "true", "yes", "on":
		return true
	default:
		return false
	}
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

func FormatStartupError(err error) string {
	if errors.Is(err, ErrMissingBotToken) {
		return "BOT_TOKEN is not set; copy .env.example to .env and set your Telegram bot token"
	}
	return fmt.Sprintf("configuration error: %v", err)
}
