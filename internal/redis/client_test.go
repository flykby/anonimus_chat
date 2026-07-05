package redis_test

import (
	"context"
	"os"
	"testing"

	iredis "github.com/flykby/anonimus_chat/internal/redis"
)

func TestOpenRequiresURL(t *testing.T) {
	t.Parallel()
	_, err := iredis.Open(context.Background(), "")
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestIntegrationPing(t *testing.T) {
	redisURL := os.Getenv("REDIS_URL")
	if redisURL == "" {
		t.Skip("REDIS_URL not set")
	}

	ctx := context.Background()
	client, err := iredis.Open(ctx, redisURL)
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	t.Cleanup(func() { _ = client.Close() })

	if err := iredis.Ping(ctx, client); err != nil {
		t.Fatalf("Ping: %v", err)
	}
}
