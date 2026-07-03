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

	addr := env.PortAddr("AI_PORT", 8001)
	if v := env.Get("HTTP_ADDR", ""); v != "" {
		addr = v
	}

	health := httputil.HealthResponse{
		"status":                      "ok",
		"service":                     "ai",
		"runpod_llm_configured":       env.Set("RUNPOD_LLM_URL"),
		"runpod_embedding_configured": env.Set("RUNPOD_EMBEDDING_URL"),
	}

	if err := httputil.Run(context.Background(), logger, addr, health); err != nil {
		logger.Error("ai server failed", "err", err)
		os.Exit(1)
	}
}
