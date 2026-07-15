<script setup lang="ts">
import type { WorkflowState, WorkflowTransition } from '~/composables/useWorkflowDefinition'
import { WORKFLOW_ROLE_OPTIONS } from '~/composables/useWorkflowDefinition'

const props = defineProps<{
  states: WorkflowState[]
  transitions: WorkflowTransition[]
}>()

const emit = defineEmits<{
  'update:transitions': [transitions: WorkflowTransition[]]
}>()

const { createEmptyTransition } = useWorkflowDefinition()

const stateOptions = computed(() =>
  props.states.filter((s) => s.code.trim()).map((s) => ({
    value: s.code,
    label: s.label ? `${s.label} (${s.code})` : s.code
  }))
)

const updateTransition = (index: number, patch: Partial<WorkflowTransition>) => {
  emit(
    'update:transitions',
    props.transitions.map((tr, i) => (i === index ? { ...tr, ...patch } : tr))
  )
}

const toggleRole = (index: number, role: string, checked: boolean) => {
  const current = props.transitions[index]?.allowedRoles ?? []
  const next = checked ? [...current, role] : current.filter((r) => r !== role)
  updateTransition(index, { allowedRoles: next })
}

const addTransition = () => {
  emit('update:transitions', [...props.transitions, createEmptyTransition()])
}

const removeTransition = (index: number) => {
  emit('update:transitions', props.transitions.filter((_, i) => i !== index))
}
</script>

<template>
  <div class="wf-transition-form">
    <p class="settings-hint">{{ $t('workflows.roles_hint') }}</p>

    <div
      v-for="(tr, index) in transitions"
      :key="`transition-${index}`"
      class="wf-transition-form__row"
    >
      <div class="settings-field">
        <label :for="`wf-tr-from-${index}`">{{ $t('workflows.col_from') }}</label>
        <select
          :id="`wf-tr-from-${index}`"
          :value="tr.from"
          required
          @change="updateTransition(index, { from: ($event.target as HTMLSelectElement).value })"
        >
          <option value="" disabled>{{ $t('workflows.select_state') }}</option>
          <option v-for="opt in stateOptions" :key="opt.value" :value="opt.value">
            {{ opt.label }}
          </option>
        </select>
      </div>

      <div class="settings-field">
        <label :for="`wf-tr-to-${index}`">{{ $t('workflows.col_to') }}</label>
        <select
          :id="`wf-tr-to-${index}`"
          :value="tr.to"
          required
          @change="updateTransition(index, { to: ($event.target as HTMLSelectElement).value })"
        >
          <option value="" disabled>{{ $t('workflows.select_state') }}</option>
          <option v-for="opt in stateOptions" :key="opt.value" :value="opt.value">
            {{ opt.label }}
          </option>
        </select>
      </div>

      <AppInput
        :id="`wf-tr-action-${index}`"
        :model-value="tr.action"
        required
        @update:model-value="updateTransition(index, { action: $event })"
      >
        <template #label>{{ $t('workflows.col_action') }}</template>
      </AppInput>

      <fieldset class="wf-transition-form__roles">
        <legend>{{ $t('workflows.roles_label') }}</legend>
        <label
          v-for="role in WORKFLOW_ROLE_OPTIONS"
          :key="role"
          class="wf-transition-form__role"
        >
          <input
            type="checkbox"
            :checked="tr.allowedRoles.includes(role)"
            @change="toggleRole(index, role, ($event.target as HTMLInputElement).checked)"
          />
          <span>{{ role }}</span>
        </label>
      </fieldset>

      <AppButton variant="ghost" size="sm" type="button" @click="removeTransition(index)">
        {{ $t('workflows.remove_transition') }}
      </AppButton>
    </div>

    <AppButton variant="ghost" size="sm" type="button" @click="addTransition">
      {{ $t('workflows.add_transition') }}
    </AppButton>
  </div>
</template>

<style scoped>
.wf-transition-form {
  display: grid;
  gap: var(--kore-space-md);
}

.wf-transition-form__row {
  display: grid;
  gap: var(--kore-space-sm);
  padding: var(--kore-space-md);
  border: 1px solid var(--kore-border);
  border-radius: var(--kore-radius-md);
  background: var(--kore-bg);
}

.settings-field {
  display: flex;
  flex-direction: column;
  gap: 0.35rem;
}

.settings-field label,
.settings-field select {
  font-size: var(--kore-text-small);
}

.settings-field select {
  padding: 0.5rem 0.75rem;
  border: 1px solid var(--kore-border);
  border-radius: var(--kore-radius-sm);
  background: var(--kore-bg-elevated);
  color: var(--kore-text);
}

.settings-hint {
  margin: 0;
  color: var(--kore-text-muted);
  font-size: var(--kore-text-small);
}

.wf-transition-form__roles {
  border: none;
  padding: 0;
  margin: 0;
  display: flex;
  flex-wrap: wrap;
  gap: var(--kore-space-sm);
}

.wf-transition-form__roles legend {
  font-size: var(--kore-text-small);
  font-weight: 600;
  width: 100%;
  margin-bottom: 0.25rem;
}

.wf-transition-form__role {
  display: inline-flex;
  align-items: center;
  gap: 0.35rem;
  font-size: var(--kore-text-small);
}

@media (max-width: 768px) {
  .wf-transition-form__row :deep(.app-button) {
    width: 100%;
  }
}
</style>
