package redistest

import (
	"context"
	"testing"

	goredis "github.com/redis/go-redis/v9"
	"github.com/alicebob/miniredis/v2"
)

func NewTestClient(t *testing.T) (*miniredis.Miniredis, *goredis.Client) {
	t.Helper()

	mr, err := miniredis.Run()
	if err != nil {
		t.Fatalf("miniredis: %v", err)
	}
	t.Cleanup(mr.Close)

	client := goredis.NewClient(&goredis.Options{Addr: mr.Addr()})
	t.Cleanup(func() { _ = client.Close() })

	ctx := context.Background()
	if err := client.Ping(ctx).Err(); err != nil {
		t.Fatalf("ping: %v", err)
	}
	return mr, client
}
