package app

import (
	"context"
	"testing"

	"github.com/kore/kore/internal/platform/cache"
)

func TestSnapshotWorker_TryAcquireSkipsWhenLocked(t *testing.T) {
	mem := cache.NewInMemoryCache()
	keys := cache.NewKeyBuilder("kore")
	w := NewSnapshotWorker(nil, nil, mem, keys, nil)
	ctx := context.Background()
	if !w.tryAcquire(ctx) {
		t.Fatal("first acquire should succeed")
	}
	if w.tryAcquire(ctx) {
		t.Fatal("second acquire should skip while lock held")
	}
}
