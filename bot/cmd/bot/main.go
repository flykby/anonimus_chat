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

	srv := newHealthServer(cfg.HTTPAddr)
	go runHealthServer(ctx, logger, srv)

	if cfg.HealthOnly {
		logger.Info("bot ready", "mode", "health_only")
		<-ctx.Done()
		shutdownHealthServer(logger, srv)
		return
	}

	runTelegramBot(ctx, logger, cfg)
	shutdownHealthServer(logger, srv)
}

func newHealthServer(addr string) *http.Server {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /health", healthHandler)
	return &http.Server{
		Addr:              addr,
		Handler:           mux,
		ReadHeaderTimeout: 5 * time.Second,
	}
}

func runHealthServer(ctx context.Context, logger *slog.Logger, srv *http.Server) {
	logger.Info("health server listening", "addr", srv.Addr)
	if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		logger.Error("health server failed", "err", err)
		os.Exit(1)
	}
}

func shutdownHealthServer(logger *slog.Logger, srv *http.Server) {
	logger.Info("shutdown signal received")
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := srv.Shutdown(shutdownCtx); err != nil {
		logger.Error("health server shutdown failed", "err", err)
		os.Exit(1)
	}
	logger.Info("bot stopped")
}

func runTelegramBot(ctx context.Context, logger *slog.Logger, cfg config.Config) {
	echo := &handlers.Echo{Logger: logger}
	tg, err := bot.New(cfg.BotToken, bot.WithDefaultHandler(echo.Default))
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
}

func healthHandler(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte(`{"status":"ok","service":"bot"}`))
}
