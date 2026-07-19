package postgres

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/kore/kore/internal/modules/workflow/domain"
	"github.com/kore/kore/internal/modules/workflow/ports"
	"github.com/kore/kore/internal/platform/db"
	"github.com/kore/kore/pkg/kernel"
)

type Repository struct {
	pool *db.Pool
}

func NewRepository(pool *db.Pool) *Repository {
	return &Repository{pool: pool}
}

func (r *Repository) SaveDefinition(ctx context.Context, def domain.WorkflowDefinition) error {
	return r.pool.WithTx(ctx, func(tx pgx.Tx) error {
		_, err := tx.Exec(ctx, `
			INSERT INTO workflow.definitions (id, tenant_id, code, entity_type, version, updated_at)
			VALUES ($1, $2, $3, $4, $5, NOW())
			ON CONFLICT (tenant_id, code) DO UPDATE SET
				entity_type = EXCLUDED.entity_type,
				version = workflow.definitions.version + 1,
				updated_at = NOW()
			RETURNING id, version
		`, def.ID, def.TenantID.UUID(), def.Code, def.EntityType, def.Version)
		if err != nil {
			return err
		}
		var defID uuid.UUID
		var version int
		if err := tx.QueryRow(ctx, `
			SELECT id, version FROM workflow.definitions WHERE tenant_id = $1 AND code = $2
		`, def.TenantID.UUID(), def.Code).Scan(&defID, &version); err != nil {
			return err
		}
		if _, err := tx.Exec(ctx, `DELETE FROM workflow.states WHERE definition_id = $1`, defID); err != nil {
			return err
		}
		if _, err := tx.Exec(ctx, `DELETE FROM workflow.transitions WHERE definition_id = $1`, defID); err != nil {
			return err
		}
		for _, state := range def.States {
			onEnter, err := marshalSideEffects(state.OnEnterEffects)
			if err != nil {
				return err
			}
			if _, err := tx.Exec(ctx, `
				INSERT INTO workflow.states (id, definition_id, code, label, is_initial, is_final, on_enter_effects)
				VALUES ($1, $2, $3, $4, $5, $6, $7)
			`, uuid.New(), defID, string(state.Code), state.Label, state.IsInitial, state.IsFinal, onEnter); err != nil {
				return err
			}
		}
		for _, tr := range def.Transitions {
			var docTrigger []byte
			if tr.DocumentTrigger != nil {
				docTrigger, err = json.Marshal(tr.DocumentTrigger)
				if err != nil {
					return err
				}
			}
			onFire, err := marshalSideEffects(tr.OnFireEffects)
			if err != nil {
				return err
			}
			if _, err := tx.Exec(ctx, `
				INSERT INTO workflow.transitions (
					id, definition_id, from_state, to_state, action, guard, doc_trigger, allowed_roles, on_fire_effects
				) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
			`, uuid.New(), defID, string(tr.From), string(tr.To), string(tr.Action),
				tr.Guard, docTrigger, tr.AllowedRoles, onFire); err != nil {
				return err
			}
		}
		return nil
	})
}

func (r *Repository) GetDefinition(ctx context.Context, tenant kernel.TenantID, code string) (domain.WorkflowDefinition, error) {
	var def domain.WorkflowDefinition
	var tenantID uuid.UUID
	err := r.pool.QueryRow(ctx, `
		SELECT id, tenant_id, code, entity_type, version
		FROM workflow.definitions WHERE tenant_id = $1 AND code = $2
	`, tenant.UUID(), code).Scan(&def.ID, &tenantID, &def.Code, &def.EntityType, &def.Version)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.WorkflowDefinition{}, domain.ErrWorkflowNotFound
		}
		return domain.WorkflowDefinition{}, err
	}
	def.TenantID = kernel.NewTenantID(tenantID)

	rows, err := r.pool.Query(ctx, `
		SELECT code, label, is_initial, is_final, on_enter_effects FROM workflow.states WHERE definition_id = $1 ORDER BY code
	`, def.ID)
	if err != nil {
		return domain.WorkflowDefinition{}, err
	}
	defer rows.Close()
	for rows.Next() {
		var s domain.State
		var code string
		var onEnter []byte
		if err := rows.Scan(&code, &s.Label, &s.IsInitial, &s.IsFinal, &onEnter); err != nil {
			return domain.WorkflowDefinition{}, err
		}
		s.Code = domain.StateCode(code)
		s.OnEnterEffects, err = unmarshalSideEffects(onEnter)
		if err != nil {
			return domain.WorkflowDefinition{}, err
		}
		def.States = append(def.States, s)
	}
	if err := rows.Err(); err != nil {
		return domain.WorkflowDefinition{}, err
	}

	trows, err := r.pool.Query(ctx, `
		SELECT from_state, to_state, action, guard, doc_trigger, allowed_roles, on_fire_effects
		FROM workflow.transitions WHERE definition_id = $1 ORDER BY from_state, action
	`, def.ID)
	if err != nil {
		return domain.WorkflowDefinition{}, err
	}
	defer trows.Close()
	for trows.Next() {
		var tr domain.Transition
		var from, to, action string
		var docTrigger []byte
		var onFire []byte
		if err := trows.Scan(&from, &to, &action, &tr.Guard, &docTrigger, &tr.AllowedRoles, &onFire); err != nil {
			return domain.WorkflowDefinition{}, err
		}
		tr.From = domain.StateCode(from)
		tr.To = domain.StateCode(to)
		tr.Action = domain.ActionCode(action)
		if len(docTrigger) > 0 {
			var trigger domain.Trigger
			if err := json.Unmarshal(docTrigger, &trigger); err != nil {
				return domain.WorkflowDefinition{}, err
			}
			tr.DocumentTrigger = &trigger
		}
		tr.OnFireEffects, err = unmarshalSideEffects(onFire)
		if err != nil {
			return domain.WorkflowDefinition{}, err
		}
		def.Transitions = append(def.Transitions, tr)
	}
	return def, trows.Err()
}

func (r *Repository) SaveInstance(ctx context.Context, inst domain.WorkflowInstance) error {
	_, err := r.pool.Exec(ctx, `
		INSERT INTO workflow.instances (id, tenant_id, definition_code, entity_id, current_state, updated_at)
		VALUES ($1, $2, $3, $4, $5, NOW())
		ON CONFLICT (id) DO UPDATE SET
			current_state = EXCLUDED.current_state,
			updated_at = NOW()
	`, inst.ID, inst.TenantID.UUID(), inst.DefinitionCode, inst.EntityID, string(inst.CurrentState))
	return err
}

func (r *Repository) GetInstance(ctx context.Context, tenant kernel.TenantID, id domain.InstanceID) (domain.WorkflowInstance, error) {
	var inst domain.WorkflowInstance
	var tenantID uuid.UUID
	var state string
	err := r.pool.QueryRow(ctx, `
		SELECT id, tenant_id, definition_code, entity_id, current_state
		FROM workflow.instances WHERE tenant_id = $1 AND id = $2
	`, tenant.UUID(), id).Scan(&inst.ID, &tenantID, &inst.DefinitionCode, &inst.EntityID, &state)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.WorkflowInstance{}, domain.ErrInstanceNotFound
		}
		return domain.WorkflowInstance{}, err
	}
	inst.TenantID = kernel.NewTenantID(tenantID)
	inst.CurrentState = domain.StateCode(state)
	return inst, nil
}

func (r *Repository) AppendLog(ctx context.Context, log domain.TransitionLog) error {
	_, err := r.pool.Exec(ctx, `
		INSERT INTO workflow.transition_logs (
			id, tenant_id, instance_id, from_state, to_state, action, actor_id, occurred_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, NOW())
	`, log.ID, log.TenantID.UUID(), log.InstanceID, string(log.FromState), string(log.ToState),
		string(log.Action), log.ActorID)
	return err
}

func (r *Repository) ListLogs(ctx context.Context, tenant kernel.TenantID, instanceID domain.InstanceID) ([]domain.TransitionLog, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT id, tenant_id, instance_id, from_state, to_state, action, actor_id, occurred_at
		FROM workflow.transition_logs
		WHERE tenant_id = $1 AND instance_id = $2
		ORDER BY occurred_at ASC
	`, tenant.UUID(), instanceID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []domain.TransitionLog
	for rows.Next() {
		var log domain.TransitionLog
		var tenantID uuid.UUID
		var from, to, action string
		var occurredAt any
		if err := rows.Scan(&log.ID, &tenantID, &log.InstanceID, &from, &to, &action, &log.ActorID, &occurredAt); err != nil {
			return nil, err
		}
		log.TenantID = kernel.NewTenantID(tenantID)
		log.FromState = domain.StateCode(from)
		log.ToState = domain.StateCode(to)
		log.Action = domain.ActionCode(action)
		log.OccurredAt = fmt.Sprint(occurredAt)
		out = append(out, log)
	}
	return out, rows.Err()
}

func marshalSideEffects(effects []domain.SideEffect) ([]byte, error) {
	if len(effects) == 0 {
		return []byte("[]"), nil
	}
	return json.Marshal(effects)
}

func unmarshalSideEffects(raw []byte) ([]domain.SideEffect, error) {
	if len(raw) == 0 {
		return nil, nil
	}
	var effects []domain.SideEffect
	if err := json.Unmarshal(raw, &effects); err != nil {
		return nil, err
	}
	return effects, nil
}

var _ ports.WorkflowRepository = (*Repository)(nil)
