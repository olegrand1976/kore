<script setup lang="ts">
import type { WorkflowState, WorkflowTransition } from '~/composables/useWorkflowDefinition'

const props = defineProps<{
  states: WorkflowState[]
  transitions: WorkflowTransition[]
}>()

const { t } = useI18n()

const stateLabel = (code: string) => {
  const state = props.states.find((s) => s.code === code)
  return state?.label || code
}
</script>

<template>
  <div class="wf-diagram" role="img" :aria-label="$t('workflows.diagram_title')">
    <div class="wf-diagram__states">
      <div
        v-for="state in states"
        :key="state.code"
        class="wf-diagram__node"
        :class="{
          'wf-diagram__node--initial': state.isInitial,
          'wf-diagram__node--final': state.isFinal
        }"
      >
        <span class="wf-diagram__node-label">{{ state.label || state.code }}</span>
        <span class="wf-diagram__node-code">{{ state.code }}</span>
        <div class="wf-diagram__badges">
          <AppBadge v-if="state.isInitial" variant="gold">{{ $t('workflows.initial') }}</AppBadge>
          <AppBadge v-if="state.isFinal" variant="success">{{ $t('workflows.final') }}</AppBadge>
        </div>
      </div>
    </div>

    <ul v-if="transitions.length" class="wf-diagram__flows">
      <li v-for="(tr, index) in transitions" :key="`${tr.from}-${tr.action}-${tr.to}-${index}`" class="wf-diagram__flow">
        <span class="wf-diagram__flow-from">{{ stateLabel(tr.from) }}</span>
        <span class="wf-diagram__flow-arrow" aria-hidden="true">→</span>
        <AppBadge variant="default">{{ tr.action }}</AppBadge>
        <span class="wf-diagram__flow-arrow" aria-hidden="true">→</span>
        <span class="wf-diagram__flow-to">{{ stateLabel(tr.to) }}</span>
        <span v-if="tr.allowedRoles.length" class="wf-diagram__flow-roles">
          ({{ tr.allowedRoles.join(', ') }})
        </span>
        <span v-else class="wf-diagram__flow-roles">{{ t('workflows.roles_all') }}</span>
      </li>
    </ul>
    <p v-else class="wf-diagram__empty">{{ $t('workflows.diagram_empty') }}</p>
  </div>
</template>

<style scoped>
.wf-diagram {
  display: grid;
  gap: var(--kore-space-md);
}

.wf-diagram__states {
  display: flex;
  flex-wrap: wrap;
  gap: var(--kore-space-sm);
}

.wf-diagram__node {
  display: grid;
  gap: 0.2rem;
  min-width: 8rem;
  padding: var(--kore-space-sm) var(--kore-space-md);
  border: 1px solid var(--kore-border);
  border-radius: var(--kore-radius-md);
  background: var(--kore-bg-elevated);
}

.wf-diagram__node--initial {
  border-color: var(--kore-primary);
}

.wf-diagram__node--final {
  border-style: dashed;
}

.wf-diagram__node-label {
  font-size: var(--kore-text-small);
  font-weight: 600;
}

.wf-diagram__node-code {
  font-size: var(--kore-text-small);
  color: var(--kore-text-muted);
}

.wf-diagram__badges {
  display: flex;
  flex-wrap: wrap;
  gap: 0.25rem;
}

.wf-diagram__flows {
  list-style: none;
  margin: 0;
  padding: 0;
  display: grid;
  gap: var(--kore-space-sm);
}

.wf-diagram__flow {
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  gap: 0.35rem;
  font-size: var(--kore-text-small);
  padding: var(--kore-space-sm);
  border: 1px solid var(--kore-border);
  border-radius: var(--kore-radius-sm);
  background: var(--kore-bg);
}

.wf-diagram__flow-from,
.wf-diagram__flow-to {
  font-weight: 500;
}

.wf-diagram__flow-arrow {
  color: var(--kore-text-muted);
}

.wf-diagram__flow-roles {
  color: var(--kore-text-muted);
  font-size: 0.85em;
}

.wf-diagram__empty {
  margin: 0;
  color: var(--kore-text-muted);
  font-size: var(--kore-text-small);
}

@media (max-width: 768px) {
  .wf-diagram__states {
    flex-direction: column;
  }

  .wf-diagram__node {
    width: 100%;
  }

  .wf-diagram__flow {
    flex-direction: column;
    align-items: flex-start;
  }
}
</style>
