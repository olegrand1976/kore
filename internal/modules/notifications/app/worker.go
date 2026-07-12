package app

import (
	"context"
	"time"

	"github.com/kore/kore/internal/platform/logging"
)

const defaultWorkerInterval = 60 * time.Second

// StartWorker polls pending notification messages and dispatches them.
func StartWorker(ctx context.Context, svc *Service, log *logging.Logger, interval time.Duration) {
	if interval <= 0 {
		interval = defaultWorkerInterval
	}
	if log == nil {
		log = logging.New("info")
	}
	ticker := time.NewTicker(interval)
	go func() {
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				n, err := svc.ProcessPending(ctx)
				if err != nil {
					log.Warn("notification worker error", "error", err)
					continue
				}
				if n > 0 {
					log.Info("notification worker dispatched", "count", n)
				}
			}
		}
	}()
}
