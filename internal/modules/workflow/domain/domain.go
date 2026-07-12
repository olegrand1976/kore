package domain

import (
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/kore/kore/internal/platform/authx"
	"github.com/kore/kore/pkg/kernel"
)

var (
	ErrTransitionNotAllowed = errors.New("transition not allowed")
	ErrGuardFailed          = errors.New("guard failed")
	ErrActionNotPermitted   = errors.New("action not permitted")
	ErrWorkflowNotFound     = errors.New("workflow not found")
	ErrInvalidDefinition    = errors.New("invalid workflow definition")
	ErrInstanceNotFound     = errors.New("workflow instance not found")
)

type StateCode string
type ActionCode string

type Trigger struct {
	Document string
	Action   string
}

type State struct {
	Code      StateCode
	Label     string
	IsInitial bool
	IsFinal   bool
}

type Transition struct {
	From            StateCode
	To              StateCode
	Action          ActionCode
	Guard           string
	DocumentTrigger *Trigger
	AllowedRoles    []string
}

type WorkflowDefinition struct {
	ID          uuid.UUID
	TenantID    kernel.TenantID
	Code        string
	EntityType  string
	Version     int
	States      []State
	Transitions []Transition
}

func (d WorkflowDefinition) Validate() error {
	if d.Code == "" || d.EntityType == "" {
		return fmt.Errorf("%w: code and entity type required", ErrInvalidDefinition)
	}
	if len(d.States) == 0 {
		return fmt.Errorf("%w: at least one state required", ErrInvalidDefinition)
	}
	initialCount := 0
	finalCount := 0
	for _, s := range d.States {
		if s.IsInitial {
			initialCount++
		}
		if s.IsFinal {
			finalCount++
		}
	}
	if initialCount != 1 {
		return fmt.Errorf("%w: exactly one initial state required", ErrInvalidDefinition)
	}
	if finalCount == 0 {
		return fmt.Errorf("%w: at least one final state required", ErrInvalidDefinition)
	}
	return nil
}

func (d WorkflowDefinition) InitialState() (StateCode, error) {
	for _, s := range d.States {
		if s.IsInitial {
			return s.Code, nil
		}
	}
	return "", ErrInvalidDefinition
}

func (d WorkflowDefinition) FindTransition(from StateCode, action ActionCode) (Transition, bool) {
	for _, t := range d.Transitions {
		if t.From == from && t.Action == action {
			return t, true
		}
	}
	return Transition{}, false
}

func (d WorkflowDefinition) AvailableTransitions(from StateCode) []Transition {
	var out []Transition
	for _, t := range d.Transitions {
		if t.From == from {
			out = append(out, t)
		}
	}
	return out
}

type InstanceID = uuid.UUID

type WorkflowInstance struct {
	ID             InstanceID
	TenantID       kernel.TenantID
	DefinitionCode string
	EntityID       string
	CurrentState   StateCode
}

type TransitionLog struct {
	ID         uuid.UUID
	TenantID   kernel.TenantID
	InstanceID InstanceID
	FromState  StateCode
	ToState    StateCode
	Action     ActionCode
	ActorID    uuid.UUID
	OccurredAt string
}

type TransitionOccurred struct {
	TenantID       kernel.TenantID
	InstanceID     InstanceID
	DefinitionCode string
	EntityID       string
	FromState      StateCode
	ToState        StateCode
	Action         ActionCode
	ActorID        uuid.UUID
}

func TransitionAllowed(t Transition, actor authx.Identity) bool {
	if len(t.AllowedRoles) == 0 {
		return true
	}
	for _, role := range t.AllowedRoles {
		if string(actor.Profile) == role {
			return true
		}
		for _, r := range actor.Roles {
			if r == role {
				return true
			}
		}
	}
	return false
}
