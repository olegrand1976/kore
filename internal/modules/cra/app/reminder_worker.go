package app

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/kore/kore/internal/modules/cra/domain"
	orgports "github.com/kore/kore/internal/modules/org/ports"
	notifports "github.com/kore/kore/internal/modules/notifications/ports"
	"github.com/kore/kore/internal/platform/logging"
)

const ReminderWorkerInterval = time.Hour

type ReminderWorker struct {
	cra      *Service
	org      orgports.OrganizationRepository
	notifier notifports.TransactionalNotifier
	log      *logging.Logger
	clock    func() time.Time
	lastRun  map[string]struct{}
}

func NewReminderWorker(
	cra *Service,
	org orgports.OrganizationRepository,
	notifier notifports.TransactionalNotifier,
	log *logging.Logger,
) *ReminderWorker {
	if log == nil {
		log = logging.New("info")
	}
	return &ReminderWorker{
		cra:      cra,
		org:      org,
		notifier: notifier,
		log:      log,
		clock:    time.Now,
		lastRun:  make(map[string]struct{}),
	}
}

func StartReminderWorker(ctx context.Context, worker *ReminderWorker, interval time.Duration) context.CancelFunc {
	if worker == nil {
		return func() {}
	}
	if interval <= 0 {
		interval = ReminderWorkerInterval
	}
	runCtx, cancel := context.WithCancel(ctx)
	ticker := time.NewTicker(interval)
	go func() {
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

func (w *ReminderWorker) runOnce(ctx context.Context) {
	now := w.clock().UTC()
	if now.Weekday() != time.Monday {
		return
	}
	lastMonday := LastMondayOfMonth(now.Year(), now.Month())
	if !sameDay(now, lastMonday) {
		return
	}
	runKey := fmt.Sprintf("%04d-%02d", now.Year(), now.Month())
	if _, done := w.lastRun[runKey]; done {
		return
	}

	targets, err := w.org.ListSocietesCraMailAuto(ctx)
	if err != nil {
		w.log.Warn("cra reminder worker: list societes", "error", err)
		return
	}
	month := domain.Month(now.Format("2006-01"))
	sent := 0
	for _, target := range targets {
		if err := w.notifyTenant(ctx, target, month); err != nil {
			w.log.Warn("cra reminder worker: tenant", "tenant", target.TenantID, "error", err)
			continue
		}
		sent++
	}
	w.lastRun[runKey] = struct{}{}
	if sent > 0 {
		w.log.Info("cra reminder worker dispatched", "tenants", sent, "month", month)
	}
}

func (w *ReminderWorker) notifyTenant(ctx context.Context, target orgports.CraMailReminderTarget, month domain.Month) error {
	pending, err := w.cra.SendMonthlyReminders(ctx, target.TenantID, month)
	if err != nil {
		return err
	}
	if len(pending) == 0 {
		return nil
	}
	userEmails, err := w.org.ResolveUserEmails(ctx, target.TenantID, pending)
	if err != nil {
		return err
	}
	recipients := append([]string{}, target.Recipients...)
	recipients = append(recipients, userEmails...)
	recipients = uniqueEmails(recipients)
	if len(recipients) == 0 {
		return fmt.Errorf("no recipients for tenant %s", target.TenantID)
	}
	body := fmt.Sprintf(
		"Rappel CRA : %d collaborateur(s) n'ont pas finalisé leur CRA pour %s.\n\nMerci de compléter vos feuilles de temps avant la fin du mois.",
		len(pending),
		month,
	)
	return w.notifier.NotifyTransactional(ctx, notifports.TransactionalMessage{
		Recipients: recipients,
		Subject:    fmt.Sprintf("Rappel CRA — %s", month),
		Body:       body,
	})
}

func sameDay(a, b time.Time) bool {
	ay, am, ad := a.Date()
	by, bm, bd := b.Date()
	return ay == by && am == bm && ad == bd
}

func uniqueEmails(items []string) []string {
	seen := make(map[string]struct{}, len(items))
	out := make([]string, 0, len(items))
	for _, raw := range items {
		email := strings.ToLower(strings.TrimSpace(raw))
		if email == "" {
			continue
		}
		if _, ok := seen[email]; ok {
			continue
		}
		seen[email] = struct{}{}
		out = append(out, email)
	}
	return out
}
