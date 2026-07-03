package config

import (
	"errors"
	"log/slog"
	"testing"
)

func TestValidateRequiresBotToken(t *testing.T) {
	t.Parallel()

	cfg := Config{}
	if err := cfg.Validate(); !errors.Is(err, ErrMissingBotToken) {
		t.Fatalf("Validate() = %v, want ErrMissingBotToken", err)
	}

	cfg.BotToken = "123:ABC"
	if err := cfg.Validate(); err != nil {
		t.Fatalf("Validate() = %v, want nil", err)
	}
}

func TestValidateHealthOnlySkipsToken(t *testing.T) {
	t.Parallel()

	cfg := Config{HealthOnly: true}
	if err := cfg.Validate(); err != nil {
		t.Fatalf("Validate() = %v, want nil for health-only mode", err)
	}
}

func TestLoadDefaults(t *testing.T) {
	t.Parallel()

	cfg := Load()
	if cfg.HTTPAddr != ":8080" {
		t.Errorf("HTTPAddr = %q, want :8080", cfg.HTTPAddr)
	}
	if cfg.LogLevel != slog.LevelInfo {
		t.Errorf("LogLevel = %v, want Info", cfg.LogLevel)
	}
}
