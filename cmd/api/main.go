package main

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	goredis "github.com/redis/go-redis/v9"

	"github.com/flykby/anonimus_chat/internal/api/users"
	"github.com/flykby/anonimus_chat/internal/db"
	"github.com/flykby/anonimus_chat/internal/platform/env"
	iredis "github.com/flykby/anonimus_chat/internal/redis"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: env.ParseLogLevel(env.Get("LOG_LEVEL", "")),
	}))

	addr := env.PortAddr("API_PORT", 8000)
	if v := env.Get("HTTP_ADDR", ""); v != "" {
		addr = v
	}

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	var pool *pgxpool.Pool
	databaseURL := env.Get("DATABASE_URL", "")
	if databaseURL != "" {
		p, err := db.Open(ctx, databaseURL)
		if err != nil {
			logger.Warn("database unavailable at startup", "err", err)
		} else {
			pool = p
			logger.Info("database connected")
			defer pool.Close()
		}
	}

	var rdb *goredis.Client
	redisURL := env.Get("REDIS_URL", "")
	if redisURL != "" {
		c, err := iredis.Open(ctx, redisURL)
		if err != nil {
			logger.Warn("redis unavailable at startup", "err", err)
		} else {
			rdb = c
			logger.Info("redis connected")
			defer rdb.Close()
		}
	}

	mux := http.NewServeMux()
	mux.HandleFunc("GET /health", healthHandler(pool, rdb))
	if pool != nil {
		(&users.Handler{Users: db.NewUsersRepo(pool)}).RegisterRoutes(mux)
	}

	srv := &http.Server{
		Addr:              addr,
		Handler:           mux,
		ReadHeaderTimeout: 5 * time.Second,
	}

	go func() {
		logger.Info("api server listening", "addr", addr)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Error("api server failed", "err", err)
			os.Exit(1)
		}
	}()

	<-ctx.Done()
	logger.Info("shutdown signal received")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := srv.Shutdown(shutdownCtx); err != nil {
		logger.Error("api shutdown failed", "err", err)
		os.Exit(1)
	}
	logger.Info("api stopped")
}

func healthHandler(pool *pgxpool.Pool, rdb *goredis.Client) http.HandlerFunc {
	return func(w http.ResponseWriter, _ *http.Request) {
		pingCtx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()

		dbOK := false
		if pool != nil {
			dbOK = db.Ping(pingCtx, pool) == nil
		}

		redisOK := false
		if rdb != nil {
			redisOK = iredis.Ping(pingCtx, rdb) == nil
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(map[string]any{
			"status":              "ok",
			"service":             "api",
			"database_configured": env.Set("DATABASE_URL"),
			"database_ok":         dbOK,
			"redis_configured":    env.Set("REDIS_URL"),
			"redis_ok":            redisOK,
		})
	}
}
