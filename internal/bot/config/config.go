package config

import (
	"errors"
	"fmt"
	"log/slog"
	"strings"

	"github.com/flykby/anonimus_chat/internal/platform/env"
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
		BotToken:   env.Get("BOT_TOKEN", ""),
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
	return nil
}

func FormatStartupError(err error) string {
	if errors.Is(err, ErrMissingBotToken) {
		return "BOT_TOKEN is not set; copy .env.example to .env and set your Telegram bot token"
	}
	return fmt.Sprintf("configuration error: %v", err)
}
