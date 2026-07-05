package redistest

import (
	"testing"

	goredis "github.com/redis/go-redis/v9"
	"github.com/alicebob/miniredis/v2"
)

func ReopenClient(t *testing.T, mr *miniredis.Miniredis) *goredis.Client {
	t.Helper()
	client := goredis.NewClient(&goredis.Options{Addr: mr.Addr()})
	t.Cleanup(func() { _ = client.Close() })
	return client
}
