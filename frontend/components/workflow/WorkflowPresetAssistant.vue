<script setup lang="ts">
import type { WorkflowDefinition, WorkflowPresetCode } from '~/composables/useWorkflowDefinition'
import { getPresetMeta } from '~/composables/useWorkflowDefinition'

const props = defineProps<{
  presetCode: WorkflowPresetCode
  definition: WorkflowDefinition
  showRestore: boolean
}>()

const emit = defineEmits<{
  restore: []
}>()

const { t } = useI18n()
const meta = computed(() => getPresetMeta(props.presetCode))

const stateLabel = (code: string) => {
  const state = props.definition.states.find((s) => s.code === code)
  return state?.label || code
}

const flowSteps = computed(() =>
  meta.value.transitions.map((tr) => ({
    from: stateLabel(tr.from),
    to: stateLabel(tr.to),
    actionLabel: t(tr.labelKey),
    actionCode: tr.action,
    hint: t(tr.hintKey),
    screen: t(tr.screenKey)
  }))
)
</script>

<template>
  <div class="wf-assistant">
    <p class="wf-assistant__notice" role="note">{{ $t('workflows.assistant.locked_notice') }}</p>

    <p class="wf-assistant__summary">{{ $t(meta.summaryKey) }}</p>

    <div class="wf-assistant__section">
      <h3 class="wf-assistant__heading">{{ $t('workflows.assistant.flow_title') }}</h3>
      <ol class="wf-assistant__flow">
        <li v-for="(step, index) in flowSteps" :key="`${step.actionCode}-${index}`" class="wf-assistant__flow-item">
          <span class="wf-assistant__flow-path">
            <strong>{{ step.from }}</strong>
            <span aria-hidden="true">→</span>
            <AppBadge variant="default">{{ step.actionLabel }}</AppBadge>
            <span class="wf-assistant__flow-code">({{ step.actionCode }})</span>
            <span aria-hidden="true">→</span>
            <strong>{{ step.to }}</strong>
          </span>
          <span class="wf-assistant__flow-hint">{{ step.hint }}</span>
          <span class="wf-assistant__flow-screen">{{ step.screen }}</span>
        </li>
      </ol>
    </div>

    <div class="wf-assistant__section">
      <h3 class="wf-assistant__heading">{{ $t('workflows.assistant.customize_title') }}</h3>
      <ul class="wf-assistant__customize">
        <li>{{ $t('workflows.assistant.customize_labels') }}</li>
        <li>{{ $t('workflows.assistant.customize_roles') }}</li>
        <li>{{ $t('workflows.assistant.customize_effects') }}</li>
      </ul>
    </div>

    <AppButton
      v-if="showRestore"
      variant="ghost"
      size="sm"
      type="button"
      @click="emit('restore')"
    >
      {{ $t('workflows.assistant.restore') }}
    </AppButton>
  </div>
</template>

<style scoped>
.wf-assistant {
  display: grid;
  gap: var(--kore-space-md);
}

.wf-assistant__notice {
  margin: 0;
  padding: var(--kore-space-md);
  border: 1px solid var(--kore-border);
  border-radius: var(--kore-radius-md);
  background: var(--kore-bg-elevated);
  color: var(--kore-text-muted);
  font-size: var(--kore-text-small);
}

.wf-assistant__summary {
  margin: 0;
  font-size: var(--kore-text-small);
  color: var(--kore-text);
}

.wf-assistant__section {
  display: grid;
  gap: var(--kore-space-sm);
}

.wf-assistant__heading {
  margin: 0;
  font-size: var(--kore-text-small);
  font-weight: 600;
}

.wf-assistant__flow {
  margin: 0;
  padding-left: 1.25rem;
  display: grid;
  gap: var(--kore-space-md);
}

.wf-assistant__flow-item {
  display: grid;
  gap: 0.25rem;
  font-size: var(--kore-text-small);
}

.wf-assistant__flow-path {
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  gap: 0.35rem;
}

.wf-assistant__flow-code {
  color: var(--kore-text-muted);
  font-family: var(--kore-font-mono, monospace);
  font-size: 0.85em;
}

.wf-assistant__flow-hint,
.wf-assistant__flow-screen {
  color: var(--kore-text-muted);
}

.wf-assistant__customize {
  margin: 0;
  padding-left: 1.25rem;
  color: var(--kore-text-muted);
  font-size: var(--kore-text-small);
  display: grid;
  gap: 0.25rem;
}

@media (max-width: 768px) {
  .wf-assistant :deep(.app-button) {
    width: 100%;
  }
}
</style>
