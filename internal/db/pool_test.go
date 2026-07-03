package db

import (
	"context"
	"os"
	"testing"
)

func TestOpenRequiresURL(t *testing.T) {
	t.Parallel()

	_, err := Open(context.Background(), "")
	if err == nil {
		t.Fatal("expected error for empty database URL")
	}
}

func TestPingNilPool(t *testing.T) {
	t.Parallel()

	if err := Ping(context.Background(), nil); err == nil {
		t.Fatal("expected error for nil pool")
	}
}

func TestIntegrationOpenPing(t *testing.T) {
	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		t.Skip("DATABASE_URL not set")
	}

	ctx := context.Background()
	pool, err := Open(ctx, databaseURL)
	if err != nil {
		t.Fatalf("Open() error: %v", err)
	}
	t.Cleanup(pool.Close)

	if err := Ping(ctx, pool); err != nil {
		t.Fatalf("Ping() error: %v", err)
	}

	var version string
	if err := pool.QueryRow(ctx, "SELECT version()").Scan(&version); err != nil {
		t.Fatalf("query version: %v", err)
	}
	if version == "" {
		t.Fatal("expected postgres version string")
	}
}
