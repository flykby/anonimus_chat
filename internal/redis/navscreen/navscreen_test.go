package navscreen_test

import (
	"testing"

	"github.com/alicebob/miniredis/v2"

	iredis "github.com/flykby/anonimus_chat/internal/redis"
	"github.com/flykby/anonimus_chat/internal/redis/navscreen"
)

func TestNavScreenSetGetDelete(t *testing.T) {
	t.Parallel()

	mr, err := miniredis.Run()
	if err != nil {
		t.Fatal(err)
	}
	client, err := iredis.Open(t.Context(), "redis://"+mr.Addr()+"/0")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = client.Close() })

	store := navscreen.New(client)
	ids := []int64{10, 11, 12}
	if err := store.Set(t.Context(), 42, ids); err != nil {
		t.Fatal(err)
	}

	got, ok, err := store.Get(t.Context(), 42)
	if err != nil || !ok {
		t.Fatalf("get: ok=%v err=%v", ok, err)
	}
	if len(got) != 3 || got[0] != 10 {
		t.Fatalf("got=%v", got)
	}

	if err := store.Delete(t.Context(), 42); err != nil {
		t.Fatal(err)
	}
	_, ok, err = store.Get(t.Context(), 42)
	if err != nil || ok {
		t.Fatalf("expected empty after delete, ok=%v err=%v", ok, err)
	}
}
