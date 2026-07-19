package domain

import (
	"fmt"

	"github.com/google/uuid"
)

const MaxSideEffectsPerHook = 10

type SideEffectType string

const SideEffectTypeEmail SideEffectType = "email"

type RecipientScope string

const (
	RecipientScopeUser        RecipientScope = "user"
	RecipientScopeEquipe      RecipientScope = "equipe"
	RecipientScopeService     RecipientScope = "service"
	RecipientScopeApplication RecipientScope = "application"
	RecipientScopeAll         RecipientScope = "all"
)

type EffectRecipients struct {
	Scope         RecipientScope `json:"scope"`
	UserIDs       []uuid.UUID    `json:"userIds,omitempty"`
	EquipeID      *uuid.UUID     `json:"equipeId,omitempty"`
	ServiceID     *uuid.UUID     `json:"serviceId,omitempty"`
	ApplicationID *uuid.UUID     `json:"applicationId,omitempty"`
}

type SideEffect struct {
	Type         SideEffectType   `json:"type"`
	Recipients   EffectRecipients `json:"recipients"`
	Subject      string           `json:"subject"`
	BodyTemplate string           `json:"bodyTemplate"`
}

func (r EffectRecipients) Validate() error {
	switch r.Scope {
	case RecipientScopeAll:
		return nil
	case RecipientScopeUser:
		if len(r.UserIDs) == 0 {
			return fmt.Errorf("%w: user scope requires userIds", ErrInvalidDefinition)
		}
	case RecipientScopeEquipe:
		if r.EquipeID == nil || *r.EquipeID == uuid.Nil {
			return fmt.Errorf("%w: equipe scope requires equipeId", ErrInvalidDefinition)
		}
	case RecipientScopeService:
		if r.ServiceID == nil || *r.ServiceID == uuid.Nil {
			return fmt.Errorf("%w: service scope requires serviceId", ErrInvalidDefinition)
		}
	case RecipientScopeApplication:
		if r.ApplicationID == nil || *r.ApplicationID == uuid.Nil {
			return fmt.Errorf("%w: application scope requires applicationId", ErrInvalidDefinition)
		}
	default:
		return fmt.Errorf("%w: unknown recipient scope %q", ErrInvalidDefinition, r.Scope)
	}
	return nil
}

func (e SideEffect) Validate() error {
	switch e.Type {
	case SideEffectTypeEmail:
	default:
		return fmt.Errorf("%w: unknown side effect type %q", ErrInvalidDefinition, e.Type)
	}
	if err := e.Recipients.Validate(); err != nil {
		return err
	}
	if e.Subject == "" && e.BodyTemplate == "" {
		return fmt.Errorf("%w: side effect requires subject or bodyTemplate", ErrInvalidDefinition)
	}
	return nil
}

func ValidateSideEffects(effects []SideEffect) error {
	if len(effects) > MaxSideEffectsPerHook {
		return fmt.Errorf("%w: at most %d side effects per hook", ErrInvalidDefinition, MaxSideEffectsPerHook)
	}
	for _, e := range effects {
		if err := e.Validate(); err != nil {
			return err
		}
	}
	return nil
}

func (d WorkflowDefinition) FindState(code StateCode) (State, bool) {
	for _, s := range d.States {
		if s.Code == code {
			return s, true
		}
	}
	return State{}, false
}

func (d WorkflowDefinition) ValidateSideEffects() error {
	for _, s := range d.States {
		if err := ValidateSideEffects(s.OnEnterEffects); err != nil {
			return err
		}
	}
	for _, tr := range d.Transitions {
		if err := ValidateSideEffects(tr.OnFireEffects); err != nil {
			return err
		}
	}
	return nil
}
