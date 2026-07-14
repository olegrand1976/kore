package app

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/kore/kore/internal/modules/cra/domain"
	notifports "github.com/kore/kore/internal/modules/notifications/ports"
	orgports "github.com/kore/kore/internal/modules/org/ports"
	"github.com/kore/kore/internal/platform/cache"
	"github.com/kore/kore/internal/platform/logging"
	"github.com/kore/kore/pkg/kernel"
)

const ReminderWorkerInterval = time.Hour

const reminderSentTTL = 35 * 24 * time.Hour

type ReminderWorker struct {
	cra      *Service
	org      orgports.OrganizationRepository
	notifier notifports.TransactionalNotifier
	cache    cache.Cache
	keys     cache.KeyBuilder
	log      *logging.Logger
	clock    func() time.Time
}

func NewReminderWorker(
	cra *Service,
	org orgports.OrganizationRepository,
	notifier notifports.TransactionalNotifier,
	appCache cache.Cache,
	keys cache.KeyBuilder,
	log *logging.Logger,
) *ReminderWorker {
	if log == nil {
		log = logging.New("info")
	}
	return &ReminderWorker{
		cra:      cra,
		org:      org,
		notifier: notifier,
		cache:    appCache,
		keys:     keys,
		log:      log,
		clock:    time.Now,
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

	targets, err := w.org.ListSocietesCraMailAuto(ctx)
	if err != nil {
		w.log.Warn("cra reminder worker: list societes", "error", err)
		return
	}
	month := domain.Month(now.Format("2006-01"))
	sent := 0
	for _, target := range targets {
		if w.alreadySent(ctx, target, month) {
			continue
		}
		if err := w.notifyTarget(ctx, target, month); err != nil {
			w.log.Warn("cra reminder worker: target", "tenant", target.TenantID, "societe", target.SocieteID, "error", err)
			continue
		}
		w.markSent(ctx, target, month)
		sent++
	}
	if sent > 0 {
		w.log.Info("cra reminder worker dispatched", "targets", sent, "month", month)
	}
}

func (w *ReminderWorker) reminderKey(target orgports.CraMailReminderTarget, month domain.Month) string {
	return w.keys.Key(target.TenantID, "cra", "reminder", target.SocieteID.String(), string(month))
}

func (w *ReminderWorker) alreadySent(ctx context.Context, target orgports.CraMailReminderTarget, month domain.Month) bool {
	if w.cache == nil || w.keys == nil {
		return false
	}
	var marker string
	found, err := w.cache.Get(ctx, w.reminderKey(target, month), &marker)
	return err == nil && found
}

func (w *ReminderWorker) markSent(ctx context.Context, target orgports.CraMailReminderTarget, month domain.Month) {
	if w.cache == nil || w.keys == nil {
		return
	}
	_ = w.cache.Set(ctx, w.reminderKey(target, month), "1", reminderSentTTL)
}

func (w *ReminderWorker) notifyTarget(ctx context.Context, target orgports.CraMailReminderTarget, month domain.Month) error {
	pending, err := w.cra.SendMonthlyReminders(ctx, target.TenantID, month)
	if err != nil {
		return err
	}
	pending = filterUsersForSociete(ctx, w.org, target.TenantID, target.SocieteID, pending)
	if len(pending) == 0 {
		return nil
	}
	userEmails, err := w.org.ResolveUserEmails(ctx, target.TenantID, pending)
	if err != nil {
		return err
	}
	recipients := uniqueEmails(append(append([]string{}, target.Recipients...), userEmails...))
	if len(recipients) == 0 {
		return fmt.Errorf("no recipients for societe %s", target.SocieteID)
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

func filterUsersForSociete(
	ctx context.Context,
	org orgports.OrganizationRepository,
	tenant kernel.TenantID,
	societeID uuid.UUID,
	userIDs []uuid.UUID,
) []uuid.UUID {
	if org == nil {
		return userIDs
	}
	out := make([]uuid.UUID, 0, len(userIDs))
	for _, userID := range userIDs {
		resolved, err := org.ResolveSocieteIDForUser(ctx, tenant, userID)
		if err != nil || resolved != societeID {
			continue
		}
		out = append(out, userID)
	}
	return out
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
