package main

import (
	"context"
	"crypto/tls"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-telegram/bot"

	"github.com/flykby/anonimus_chat/internal/bot/apiclient"
	"github.com/flykby/anonimus_chat/internal/bot/config"
	"github.com/flykby/anonimus_chat/internal/bot/handlers"
	"github.com/flykby/anonimus_chat/internal/bot/webhook"
	iredis "github.com/flykby/anonimus_chat/internal/redis"
	"github.com/flykby/anonimus_chat/internal/redis/fsm"
	"github.com/flykby/anonimus_chat/internal/redis/navscreen"
	"github.com/flykby/anonimus_chat/internal/redis/regdraft"
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

	if cfg.HealthOnly {
		srv := newHealthServer(cfg.HTTPAddr)
		go runHealthServer(ctx, logger, srv)
		logger.Info("bot ready", "mode", "health_only")
		<-ctx.Done()
		shutdownHealthServer(logger, srv)
		return
	}

	if cfg.UseWebhook() {
		runTelegramBot(ctx, logger, cfg)
		return
	}

	srv := newHealthServer(cfg.HTTPAddr)
	go runHealthServer(ctx, logger, srv)
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
	rdb, err := iredis.Open(ctx, cfg.RedisURL)
	if err != nil {
		logger.Error("redis unavailable", "err", err)
		os.Exit(1)
	}
	defer rdb.Close()

	app := &handlers.App{
		Logger:              logger,
		FSM:                 fsm.New(rdb),
		Draft:               regdraft.New(rdb),
		NavScreen:           navscreen.New(rdb),
		API:                 apiclient.NewClient(cfg.APIURL),
		ReportChatID:        cfg.ReportChatID,
		PremiumPriceStars:   cfg.PremiumPriceStars,
		PremiumDurationDays: cfg.PremiumDurationDays,
	}

	tg, err := bot.New(cfg.BotToken, bot.WithDefaultHandler(app.Default))
	if err != nil {
		logger.Error("failed to create telegram bot", "err", err)
		os.Exit(1)
	}
	app.Register(tg)

	if _, err := tg.GetMe(ctx); err != nil {
		logger.Error("invalid BOT_TOKEN or Telegram API unavailable", "err", err)
		os.Exit(1)
	}

	if cfg.UseWebhook() {
		runWebhookMode(ctx, logger, cfg, tg)
	} else {
		runPollingMode(ctx, logger, tg)
	}
}

func runPollingMode(ctx context.Context, logger *slog.Logger, tg *bot.Bot) {
	go func() {
		logger.Info("telegram long polling started")
		tg.Start(ctx)
	}()

	logger.Info("bot ready", "mode", "long_polling")
	<-ctx.Done()
}

func runWebhookMode(ctx context.Context, logger *slog.Logger, cfg config.Config, tg *bot.Bot) {
	wh := webhook.New(tg, cfg.WebhookSecret, logger)

	if err := wh.Register(ctx, cfg.WebhookURL, cfg.WebhookCertPath); err != nil {
		logger.Error("failed to register webhook", "err", err)
		os.Exit(1)
	}
	logger.Info("webhook registered", "url", cfg.WebhookURL)

	mux := http.NewServeMux()
	mux.HandleFunc("GET /health", healthHandler)
	mux.HandleFunc("POST /telegram/webhook", wh.Handler)

	srv := &http.Server{
		Addr:              cfg.HTTPAddr,
		Handler:           mux,
		ReadHeaderTimeout: 5 * time.Second,
	}

	go func() {
		var err error
		if cfg.WebhookCertPath != "" && cfg.WebhookKeyPath != "" {
			logger.Info("webhook server listening (TLS)", "addr", cfg.HTTPAddr)
			srv.TLSConfig = &tls.Config{MinVersion: tls.VersionTLS12}
			err = srv.ListenAndServeTLS(cfg.WebhookCertPath, cfg.WebhookKeyPath)
		} else {
			logger.Info("webhook server listening", "addr", cfg.HTTPAddr)
			err = srv.ListenAndServe()
		}
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			logger.Error("webhook server failed", "err", err)
			os.Exit(1)
		}
	}()

	logger.Info("bot ready", "mode", "webhook")
	<-ctx.Done()

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := srv.Shutdown(shutdownCtx); err != nil {
		logger.Error("webhook server shutdown failed", "err", err)
	}
}

func healthHandler(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte(`{"status":"ok","service":"bot"}`))
}
