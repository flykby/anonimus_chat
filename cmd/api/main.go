package main

import (
	"context"
	"log/slog"
	"os"

	"github.com/flykby/anonimus_chat/internal/platform/env"
	"github.com/flykby/anonimus_chat/internal/platform/httputil"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: env.ParseLogLevel(env.Get("LOG_LEVEL", "")),
	}))

	addr := env.PortAddr("API_PORT", 8000)
	if v := env.Get("HTTP_ADDR", ""); v != "" {
		addr = v
	}

	health := httputil.HealthResponse{
		"status":              "ok",
		"service":             "api",
		"database_configured": env.Set("DATABASE_URL"),
		"redis_configured":    env.Set("REDIS_URL"),
	}

	if err := httputil.Run(context.Background(), logger, addr, health); err != nil {
		logger.Error("api server failed", "err", err)
		os.Exit(1)
	}
}
