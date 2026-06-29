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

func TestParseLogLevel(t *testing.T) {
	t.Parallel()

	cases := map[string]slog.Level{
		"debug":   slog.LevelDebug,
		"WARN":    slog.LevelWarn,
		"error":   slog.LevelError,
		"":        slog.LevelInfo,
		"unknown": slog.LevelInfo,
	}

	for input, want := range cases {
		if got := parseLogLevel(input); got != want {
			t.Errorf("parseLogLevel(%q) = %v, want %v", input, got, want)
		}
	}
}

func TestLoadDefaults(t *testing.T) {
	t.Parallel()

	cfg := Load()
	if cfg.HTTPAddr != ":8080" {
		t.Errorf("HTTPAddr = %q, want :8080", cfg.HTTPAddr)
	}
}
