package main

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-telegram/bot"

	"github.com/flykby/anonimus_chat/bot/internal/config"
	"github.com/flykby/anonimus_chat/bot/internal/handlers"
)

func main() {
	cfg := config.Load()
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: cfg.LogLevel,
	}))

	if err := cfg.Validate(); err != nil {
		logger.Error(config.FormatStartupError(err))
		os.Exit(1)
	}

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	mux := http.NewServeMux()
	mux.HandleFunc("GET /health", healthHandler)

	srv := &http.Server{
		Addr:              cfg.HTTPAddr,
		Handler:           mux,
		ReadHeaderTimeout: 5 * time.Second,
	}

	go func() {
		logger.Info("health server listening", "addr", cfg.HTTPAddr)
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			logger.Error("health server failed", "err", err)
			os.Exit(1)
		}
	}()

	echo := &handlers.Echo{Logger: logger}
	tg, err := bot.New(cfg.BotToken,
		bot.WithDefaultHandler(echo.Default),
	)
	if err != nil {
		logger.Error("failed to create telegram bot", "err", err)
		os.Exit(1)
	}
	echo.Register(tg)

	if _, err := tg.GetMe(ctx); err != nil {
		logger.Error("invalid BOT_TOKEN or Telegram API unavailable", "err", err)
		os.Exit(1)
	}

	go func() {
		logger.Info("telegram long polling started")
		tg.Start(ctx)
	}()

	logger.Info("echo bot ready", "mode", "long_polling")

	<-ctx.Done()
	logger.Info("shutdown signal received")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := srv.Shutdown(shutdownCtx); err != nil {
		logger.Error("health server shutdown failed", "err", err)
		os.Exit(1)
	}
	logger.Info("bot stopped")
}

func healthHandler(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte(`{"status":"ok","service":"bot"}`))
}
