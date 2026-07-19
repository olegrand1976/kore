<script setup lang="ts">
import type { WorkflowPresetCode, WorkflowState, WorkflowTransition } from '~/composables/useWorkflowDefinition'
import { getStateMeta, stateReferencedByTransition } from '~/composables/useWorkflowDefinition'

const props = defineProps<{
  states: WorkflowState[]
  transitions: WorkflowTransition[]
  presetCode?: WorkflowPresetCode
  guided?: boolean
  lockCodes?: boolean
}>()

const emit = defineEmits<{
  'update:states': [states: WorkflowState[]]
}>()

const { t } = useI18n()
const { createEmptyState } = useWorkflowDefinition()

const isGuided = computed(() => props.guided === true)

const updateState = (index: number, patch: Partial<WorkflowState>) => {
  const next = props.states.map((s, i) => {
    if (i !== index) return s
    const merged = { ...s, ...patch }
    if (patch.onEnterEffects === undefined && !merged.onEnterEffects) {
      merged.onEnterEffects = []
    }
    return merged
  })
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

const stateHint = (code: string) => {
  if (!props.presetCode) return ''
  const meta = getStateMeta(props.presetCode, code)
  return meta ? t(meta.hintKey) : ''
}
</script>

<template>
  <div class="wf-state-form">
    <div
      v-for="(state, index) in states"
      :key="`state-${index}-${state.code}`"
      class="wf-state-form__row"
    >
      <template v-if="isGuided">
        <div class="wf-state-form__readonly">
          <span class="wf-state-form__readonly-label">
            {{ $t('workflows.col_code') }}
            <AppTooltip :button-label="$t('common.info')">
              {{ stateHint(state.code) }}
            </AppTooltip>
          </span>
          <code class="wf-state-form__code">{{ state.code }}</code>
          <div class="wf-state-form__badges">
            <AppBadge v-if="state.isInitial" variant="gold">{{ $t('workflows.initial') }}</AppBadge>
            <AppBadge v-if="state.isFinal" variant="success">{{ $t('workflows.final') }}</AppBadge>
          </div>
        </div>

        <AppInput
          :id="`wf-state-label-${index}`"
          :model-value="state.label"
          required
          @update:model-value="updateState(index, { label: $event })"
        >
          <template #label>{{ $t('workflows.col_label') }}</template>
        </AppInput>
        <p v-if="stateHint(state.code)" class="settings-hint">{{ stateHint(state.code) }}</p>

        <div class="wf-state-form__effects">
          <h4 class="wf-state-form__effects-title">{{ $t('workflows.effects.on_enter_title') }}</h4>
          <WorkflowSideEffectsEditor
            :id-prefix="`state-${index}`"
            :model-value="state.onEnterEffects ?? []"
            @update:model-value="updateState(index, { onEnterEffects: $event })"
          />
        </div>
      </template>

      <template v-else>
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

        <div class="wf-state-form__effects">
          <h4 class="wf-state-form__effects-title">{{ $t('workflows.effects.on_enter_title') }}</h4>
          <WorkflowSideEffectsEditor
            :id-prefix="`state-${index}`"
            :model-value="state.onEnterEffects ?? []"
            @update:model-value="updateState(index, { onEnterEffects: $event })"
          />
        </div>
      </template>
    </div>

    <AppButton v-if="!isGuided" variant="ghost" size="sm" type="button" @click="addState">
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

.wf-state-form__readonly {
  display: grid;
  gap: 0.35rem;
}

.wf-state-form__readonly-label {
  display: inline-flex;
  align-items: center;
  gap: 0.35rem;
  font-size: var(--kore-text-small);
  font-weight: 600;
}

.wf-state-form__code {
  font-size: var(--kore-text-small);
  color: var(--kore-text-muted);
}

.wf-state-form__badges {
  display: flex;
  flex-wrap: wrap;
  gap: var(--kore-space-xs);
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

.settings-hint {
  margin: 0;
  color: var(--kore-text-muted);
  font-size: var(--kore-text-small);
}

.wf-state-form__effects {
  display: grid;
  gap: var(--kore-space-sm);
  margin-top: var(--kore-space-sm);
}

.wf-state-form__effects-title {
  margin: 0;
  font-size: var(--kore-text-small);
  font-weight: 600;
}

@media (max-width: 768px) {
  .wf-state-form__row :deep(.app-button) {
    width: 100%;
  }
}
</style>
