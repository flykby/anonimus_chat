package config

import (
	"errors"
	"fmt"
	"log/slog"
	"strings"

	"github.com/flykby/anonimus_chat/internal/platform/env"
)

var (
	ErrMissingBotToken = errors.New("BOT_TOKEN is required")
	ErrMissingAPIURL   = errors.New("API_URL is required")
	ErrMissingRedisURL = errors.New("REDIS_URL is required")
)

type Config struct {
	BotToken   string
	APIURL     string
	RedisURL   string
	ReportChatID int64
	HTTPAddr   string
	LogLevel   slog.Level
	HealthOnly bool
}

func Load() Config {
	return Config{
		BotToken:   env.Get("BOT_TOKEN", ""),
		APIURL:     env.Get("API_URL", "http://api:8000"),
		RedisURL:   env.Get("REDIS_URL", ""),
		ReportChatID: env.Int64("REPORT_CHAT_ID", 0),
		HTTPAddr:   env.Get("HTTP_ADDR", ":8080"),
		LogLevel:   env.ParseLogLevel(env.Get("LOG_LEVEL", "")),
		HealthOnly: env.Bool("BOT_HEALTH_ONLY"),
	}
}

func (c Config) Validate() error {
	if c.HealthOnly {
		return nil
	}
	if strings.TrimSpace(c.BotToken) == "" {
		return ErrMissingBotToken
	}
	if strings.TrimSpace(c.APIURL) == "" {
		return ErrMissingAPIURL
	}
	if strings.TrimSpace(c.RedisURL) == "" {
		return ErrMissingRedisURL
	}
	return nil
}

func FormatStartupError(err error) string {
	switch {
	case errors.Is(err, ErrMissingBotToken):
		return "BOT_TOKEN is not set; copy .env.example to .env and set your Telegram bot token"
	case errors.Is(err, ErrMissingAPIURL):
		return "API_URL is not set"
	case errors.Is(err, ErrMissingRedisURL):
		return "REDIS_URL is not set"
	default:
		return fmt.Sprintf("configuration error: %v", err)
	}
}
