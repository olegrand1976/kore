package notifications

import (
	"context"

	"github.com/kore/kore/internal/modules/invoicing/ports"
	notifports "github.com/kore/kore/internal/modules/notifications/ports"
)

type MailSender struct {
	notifier notifports.TransactionalNotifier
}

func NewMailSender(notifier notifports.TransactionalNotifier) *MailSender {
	return &MailSender{notifier: notifier}
}

func (m *MailSender) Send(ctx context.Context, to, subject, body string) error {
	if m == nil || m.notifier == nil {
		return nil
	}
	return m.notifier.NotifyTransactional(ctx, notifports.TransactionalMessage{
		Recipients:    []string{to},
		Subject:       subject,
		Body:          body,
		SkipSignature: true,
	})
}

var _ ports.MailSender = (*MailSender)(nil)
