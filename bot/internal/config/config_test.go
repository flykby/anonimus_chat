package config

import (
	"log/slog"
	"testing"
)

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
	if cfg.LogLevel != slog.LevelInfo {
		t.Errorf("LogLevel = %v, want Info", cfg.LogLevel)
	}
}
