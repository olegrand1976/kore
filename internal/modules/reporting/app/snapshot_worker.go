package app

import (
	"context"
	"time"

	"github.com/kore/kore/internal/modules/reporting/ports"
	"github.com/kore/kore/internal/platform/cache"
	"github.com/kore/kore/internal/platform/logging"
	"github.com/kore/kore/pkg/kernel"
)

const SnapshotWorkerInterval = time.Hour

const snapshotWorkerLockTTL = 50 * time.Minute

type SnapshotWorker struct {
	svc   ports.ReportingService
	repo  ports.ReportingRepository
	cache cache.Cache
	keys  cache.KeyBuilder
	log   *logging.Logger
}

func NewSnapshotWorker(
	svc ports.ReportingService,
	repo ports.ReportingRepository,
	appCache cache.Cache,
	keys cache.KeyBuilder,
	log *logging.Logger,
) *SnapshotWorker {
	return &SnapshotWorker{svc: svc, repo: repo, cache: appCache, keys: keys, log: log}
}

func StartSnapshotWorker(ctx context.Context, worker *SnapshotWorker, interval time.Duration) context.CancelFunc {
	if interval <= 0 {
		interval = SnapshotWorkerInterval
	}
	runCtx, cancel := context.WithCancel(ctx)
	go func() {
		ticker := time.NewTicker(interval)
		defer ticker.Stop()
		worker.runOnce(runCtx)
		for {
			select {
			case <-runCtx.Done():
				return
			case <-ticker.C:
				worker.runOnce(runCtx)
			}
		}
	}()
	return cancel
}

func (w *SnapshotWorker) runOnce(ctx context.Context) {
	if w == nil || w.svc == nil || w.repo == nil {
		return
	}
	if !w.tryAcquire(ctx) {
		return
	}
	tenants, err := w.repo.ListTenantIDsForSnapshotRefresh(ctx)
	if err != nil {
		if w.log != nil {
			w.log.Error("reporting snapshot worker list tenants", "error", err)
		}
		return
	}
	for _, tenant := range tenants {
		w.refreshTenant(ctx, tenant)
	}
}

func (w *SnapshotWorker) tryAcquire(ctx context.Context) bool {
	if w.cache == nil || w.keys == nil {
		return true
	}
	key := w.keys.PublicKey("reporting", "snapshot-worker", "lock")
	var marker string
	found, err := w.cache.Get(ctx, key, &marker)
	if err == nil && found {
		return false
	}
	_ = w.cache.Set(ctx, key, "1", snapshotWorkerLockTTL)
	return true
}

func (w *SnapshotWorker) refreshTenant(ctx context.Context, tenant kernel.TenantID) {
	if err := w.svc.RefreshDashboardSnapshot(ctx, tenant, "cra"); err != nil && w.log != nil {
		w.log.Error("reporting snapshot refresh", "tenant", tenant.String(), "code", "cra", "error", err)
	}
}
