<script setup lang="ts">
import type { WorkflowState, WorkflowTransition } from '~/composables/useWorkflowDefinition'
import { stateReferencedByTransition } from '~/composables/useWorkflowDefinition'

const props = defineProps<{
  states: WorkflowState[]
  transitions: WorkflowTransition[]
  lockCodes?: boolean
}>()

const emit = defineEmits<{
  'update:states': [states: WorkflowState[]]
}>()

const { createEmptyState } = useWorkflowDefinition()

const updateState = (index: number, patch: Partial<WorkflowState>) => {
  const next = props.states.map((s, i) => (i === index ? { ...s, ...patch } : s))
  if (patch.isInitial) {
    emit('update:states', next.map((s, i) => ({ ...s, isInitial: i === index })))
    return
  }
  emit('update:states', next)
}

const addState = () => {
  emit('update:states', [...props.states, createEmptyState()])
}

const removeState = (index: number) => {
  const code = props.states[index]?.code
  if (code && stateReferencedByTransition(code, props.transitions)) return
  emit('update:states', props.states.filter((_, i) => i !== index))
}

const canRemove = (state: WorkflowState) =>
  !stateReferencedByTransition(state.code, props.transitions)
</script>

<template>
  <div class="wf-state-form">
    <div
      v-for="(state, index) in states"
      :key="`state-${index}-${state.code}`"
      class="wf-state-form__row"
    >
      <AppInput
        :id="`wf-state-code-${index}`"
        :model-value="state.code"
        :disabled="lockCodes"
        required
        @update:model-value="updateState(index, { code: $event })"
      >
        <template #label>{{ $t('workflows.col_code') }}</template>
      </AppInput>

      <AppInput
        :id="`wf-state-label-${index}`"
        :model-value="state.label"
        required
        @update:model-value="updateState(index, { label: $event })"
      >
        <template #label>{{ $t('workflows.col_label') }}</template>
      </AppInput>

      <div class="wf-state-form__flags">
        <label class="wf-state-form__flag">
          <input
            type="radio"
            name="wf-initial-state"
            :checked="state.isInitial"
            @change="updateState(index, { isInitial: true })"
          />
          <span>{{ $t('workflows.initial') }}</span>
        </label>
        <label class="wf-state-form__flag">
          <input
            type="checkbox"
            :checked="state.isFinal"
            @change="updateState(index, { isFinal: ($event.target as HTMLInputElement).checked })"
          />
          <span>{{ $t('workflows.final') }}</span>
        </label>
      </div>

      <AppButton
        variant="ghost"
        size="sm"
        type="button"
        :disabled="!canRemove(state)"
        :title="canRemove(state) ? '' : $t('workflows.cannot_remove_state')"
        @click="removeState(index)"
      >
        {{ $t('workflows.remove_state') }}
      </AppButton>
    </div>

    <AppButton variant="ghost" size="sm" type="button" @click="addState">
      {{ $t('workflows.add_state') }}
    </AppButton>
  </div>
</template>

<style scoped>
.wf-state-form {
  display: grid;
  gap: var(--kore-space-md);
}

.wf-state-form__row {
  display: grid;
  gap: var(--kore-space-sm);
  padding: var(--kore-space-md);
  border: 1px solid var(--kore-border);
  border-radius: var(--kore-radius-md);
  background: var(--kore-bg);
}

.wf-state-form__flags {
  display: flex;
  flex-wrap: wrap;
  gap: var(--kore-space-md);
}

.wf-state-form__flag {
  display: inline-flex;
  align-items: center;
  gap: 0.35rem;
  font-size: var(--kore-text-small);
}

@media (max-width: 768px) {
  .wf-state-form__row :deep(.app-button) {
    width: 100%;
  }
}
</style>
