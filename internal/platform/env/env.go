package env

import (
	"log/slog"
	"os"
	"strconv"
	"strings"
)

func Get(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func Bool(key string) bool {
	switch strings.ToLower(strings.TrimSpace(os.Getenv(key))) {
	case "1", "true", "yes", "on":
		return true
	default:
		return false
	}
}

func Set(key string) bool {
	return strings.TrimSpace(os.Getenv(key)) != ""
}

func ParseLogLevel(raw string) slog.Level {
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

func PortAddr(name string, defaultPort int) string {
	if v := os.Getenv(name); v != "" {
		if strings.HasPrefix(v, ":") {
			return v
		}
		if _, err := strconv.Atoi(v); err == nil {
			return ":" + v
		}
		return v
	}
	return ":" + strconv.Itoa(defaultPort)
}
