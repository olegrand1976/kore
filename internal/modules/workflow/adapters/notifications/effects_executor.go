package notifications

import (
	"context"
	"fmt"

	notifdomain "github.com/kore/kore/internal/modules/notifications/domain"
	notifports "github.com/kore/kore/internal/modules/notifications/ports"
	"github.com/kore/kore/internal/modules/workflow/domain"
	"github.com/kore/kore/internal/modules/workflow/ports"
	"github.com/kore/kore/pkg/kernel"
)

type EffectsExecutor struct {
	resolver notifports.RecipientResolver
	notifier notifports.TransactionalNotifier
}

func NewEffectsExecutor(
	resolver notifports.RecipientResolver,
	notifier notifports.TransactionalNotifier,
) ports.SideEffectExecutor {
	return &EffectsExecutor{resolver: resolver, notifier: notifier}
}

func (e *EffectsExecutor) Execute(ctx context.Context, effects []domain.SideEffect, effectCtx ports.SideEffectContext) error {
	if e == nil || len(effects) == 0 {
		return nil
	}
	vars := sideEffectVars(effectCtx)
	for _, effect := range effects {
		if err := e.executeOne(ctx, effect, effectCtx, vars); err != nil {
			return err
		}
	}
	return nil
}

func (e *EffectsExecutor) executeOne(
	ctx context.Context,
	effect domain.SideEffect,
	effectCtx ports.SideEffectContext,
	vars map[string]string,
) error {
	switch effect.Type {
	case domain.SideEffectTypeEmail:
	default:
		return fmt.Errorf("unsupported side effect type %q", effect.Type)
	}
	if e.resolver == nil || e.notifier == nil {
		return nil
	}
	recipients, err := e.resolveEmails(ctx, effectCtx.TenantID, effect.Recipients)
	if err != nil {
		return err
	}
	if len(recipients) == 0 {
		return notifdomain.ErrNoRecipients
	}
	subject := notifdomain.ApplyTemplate(effect.Subject, vars)
	if subject == "" {
		subject = fmt.Sprintf("Workflow %s", effectCtx.DefinitionCode)
	}
	body := notifdomain.ApplyTemplate(effect.BodyTemplate, vars)
	body = notifdomain.WithSignature(body, notifdomain.DefaultSignature("", ""))
	return e.notifier.NotifyTransactional(ctx, notifports.TransactionalMessage{
		Subject:    subject,
		Body:       body,
		Recipients: recipients,
	})
}

func (e *EffectsExecutor) resolveEmails(ctx context.Context, tenant kernel.TenantID, recipients domain.EffectRecipients) ([]string, error) {
	switch recipients.Scope {
	case domain.RecipientScopeAll:
		return e.resolver.ResolveTenantUserEmails(ctx, tenant)
	case domain.RecipientScopeUser:
		return e.resolver.ResolveUserEmails(ctx, tenant, recipients.UserIDs)
	case domain.RecipientScopeEquipe:
		return e.resolver.ResolveEquipeUserEmails(ctx, tenant, *recipients.EquipeID)
	case domain.RecipientScopeApplication:
		return e.resolver.ResolveApplicationUserEmails(ctx, tenant, *recipients.ApplicationID)
	case domain.RecipientScopeService:
		return e.resolver.ResolveServiceUserEmails(ctx, tenant, *recipients.ServiceID)
	default:
		return nil, fmt.Errorf("unknown recipient scope %q", recipients.Scope)
	}
}

func sideEffectVars(effectCtx ports.SideEffectContext) map[string]string {
	action := string(effectCtx.Action)
	if action == "" {
		action = "-"
	}
	return map[string]string{
		"entityId":       effectCtx.EntityID,
		"definitionCode": effectCtx.DefinitionCode,
		"fromState":      string(effectCtx.FromState),
		"toState":        string(effectCtx.ToState),
		"action":         action,
		"actorId":        effectCtx.ActorID.String(),
		"instanceId":     effectCtx.InstanceID.String(),
	}
}
