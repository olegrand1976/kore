package notifications

import (
	"context"

	notifdomain "github.com/kore/kore/internal/modules/notifications/domain"
	notifports "github.com/kore/kore/internal/modules/notifications/ports"
	"github.com/kore/kore/internal/modules/publicsite/ports"
)

type NotifierAdapter struct {
	notifier notifports.TransactionalNotifier
}

func NewNotifierAdapter(notifier notifports.TransactionalNotifier) ports.TransactionalNotifier {
	return &NotifierAdapter{notifier: notifier}
}

func (a *NotifierAdapter) NotifyTransactional(ctx context.Context, msg ports.TransactionalMessage) error {
	attachments := make([]notifdomain.Attachment, len(msg.Attachments))
	for i, att := range msg.Attachments {
		attachments[i] = notifdomain.Attachment{
			Filename:    att.Filename,
			Content:     att.Content,
			ContentType: att.ContentType,
		}
	}
	return a.notifier.NotifyTransactional(ctx, notifports.TransactionalMessage{
		Subject:     msg.Subject,
		Body:        msg.Body,
		Recipients:  msg.To,
		Attachments: attachments,
	})
}
