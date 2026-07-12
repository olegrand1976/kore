package app

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/kore/kore/internal/modules/notifications/adapters/email"
	"github.com/kore/kore/internal/modules/notifications/domain"
	"github.com/kore/kore/internal/modules/notifications/ports"
	"github.com/kore/kore/pkg/kernel"
)

type Service struct {
	repo     ports.NotificationRepository
	sender   ports.EmailSender
	resolver ports.RecipientResolver
	clock    ports.Clock
}

func NewService(
	repo ports.NotificationRepository,
	sender ports.EmailSender,
	resolver ports.RecipientResolver,
) *Service {
	return &Service{
		repo:     repo,
		sender:   sender,
		resolver: resolver,
		clock:    realClock{},
	}
}

type realClock struct{}

func (realClock) Now() time.Time { return time.Now() }

func (s *Service) DefineRule(ctx context.Context, rule domain.NotificationRule) error {
	if _, err := domain.ParseFrequency(string(rule.Frequency)); err != nil {
		return err
	}
	if rule.ID == uuid.Nil {
		rule.ID = uuid.New()
	}
	return s.repo.SaveRule(ctx, rule)
}

func (s *Service) ListRules(ctx context.Context, tenant kernel.TenantID) ([]domain.NotificationRule, error) {
	return s.repo.ListRules(ctx, tenant)
}

func (s *Service) Notify(ctx context.Context, evt ports.NotificationEvent) error {
	return s.Publish(ctx, evt)
}

func (s *Service) Publish(ctx context.Context, evt ports.NotificationEvent) error {
	rule, err := s.repo.GetRuleByTrigger(ctx, evt.TenantID, evt.Trigger)
	if err != nil {
		if errors.Is(err, domain.ErrRuleNotFound) {
			return domain.ErrRuleNotFound
		}
		return err
	}

	recipients, err := s.resolveRecipients(ctx, evt.TenantID, rule.RecipientsPolicy)
	if err != nil {
		return err
	}
	if len(recipients) == 0 {
		return domain.ErrNoRecipients
	}

	subject := evt.Subject
	if subject == "" {
		subject = rule.Code
	}
	body := domain.ApplyTemplate(rule.Template, evt.Vars)
	body = domain.WithSignature(body, domain.DefaultSignature("", ""))

	msg := domain.NotificationMessage{
		ID:         uuid.New(),
		TenantID:   evt.TenantID,
		RuleCode:   rule.Code,
		Recipients: recipients,
		Subject:    subject,
		Body:       body,
		Status:     domain.MessageStatusPending,
	}

	if rule.Frequency.IsImmediate() {
		return s.dispatch(ctx, &msg)
	}

	return s.repo.SaveMessage(ctx, msg)
}

func (s *Service) NotifyTransactional(ctx context.Context, msg ports.TransactionalMessage) error {
	if len(msg.Recipients) == 0 {
		return domain.ErrNoRecipients
	}
	body := msg.Body
	if !msg.SkipSignature {
		body = domain.WithSignature(body, domain.DefaultSignature("", ""))
	}

	notification := domain.NotificationMessage{
		ID:          uuid.New(),
		TenantID:    kernel.TenantID{},
		Recipients:  msg.Recipients,
		Subject:     msg.Subject,
		Body:        body,
		Attachments: msg.Attachments,
		Status:      domain.MessageStatusPending,
	}
	return s.dispatch(ctx, &notification)
}

func (s *Service) ListSent(ctx context.Context, filter ports.SentFilter) ([]domain.NotificationMessage, error) {
	if filter.Limit <= 0 {
		filter.Limit = 100
	}
	return s.repo.ListMessages(ctx, filter)
}

func (s *Service) resolveRecipients(ctx context.Context, tenant kernel.TenantID, policy domain.RecipientPolicy) ([]string, error) {
	if len(policy.UserIDs) == 0 {
		return nil, nil
	}
	return s.resolver.ResolveUserEmails(ctx, tenant, policy.UserIDs)
}

func (s *Service) dispatch(ctx context.Context, msg *domain.NotificationMessage) error {
	msg.Attempts++
	htmlBody := email.RenderNotification(msg.Body, "Kore", "https://kore.app")
	if msg.Subject != "" && msg.RuleCode == "" {
		htmlBody = email.RenderTransactional(msg.Subject, msg.Body, "Kore", "https://kore.app")
	}
	err := s.sender.Send(ctx, ports.Email{
		To:          msg.Recipients,
		Subject:     msg.Subject,
		Body:        msg.Body,
		HTMLBody:    htmlBody,
		Attachments: msg.Attachments,
	})
	now := s.clock.Now().UTC()
	if err != nil {
		msg.Status = domain.MessageStatusPending
		if saveErr := s.repo.SaveMessage(ctx, *msg); saveErr != nil {
			return saveErr
		}
		return fmt.Errorf("send email: %w", err)
	}
	msg.Status = domain.MessageStatusSent
	msg.SentAt = &now
	return s.repo.SaveMessage(ctx, *msg)
}

var (
	_ ports.NotificationService   = (*Service)(nil)
	_ ports.NotificationPublisher = (*Service)(nil)
	_ ports.TransactionalNotifier = (*Service)(nil)
)
