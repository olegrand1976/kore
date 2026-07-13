<script setup lang="ts">
const props = defineProps<{
  status: string
  actions: string[]
  canValidateTma: boolean
  assigneeId?: string
  requiresChefGate?: boolean
  busy?: boolean
  users?: { id: string; label: string }[]
}>()

const emit = defineEmits<{
  validateCreation: []
  assign: [assigneeId: string]
  takeOver: []
  resolve: []
  reopen: [reason: string]
}>()

const { t } = useI18n()
const reopenReason = ref('')
const selectedAssignee = ref('')

const hasAction = (code: string) => props.actions.includes(code)

watch(
  () => props.users,
  (list) => {
    if (list?.length && !selectedAssignee.value) {
      selectedAssignee.value = list[0].id
    }
  },
  { immediate: true }
)

const showValidate = computed(
  () =>
    props.canValidateTma &&
    props.status === 'en_attente_creation' &&
    (props.requiresChefGate ?? false)
)

const showAssign = computed(() => hasAction('assign'))

const showTakeOver = computed(() => {
  if (hasAction('assign')) return false
  const open = new Set(['ouverte', 'affectee', 'rework'])
  return open.has(props.status)
})

const showResolve = computed(() => hasAction('resolve'))

const showReopen = computed(() => hasAction('reopen'))

const onAssign = () => {
  if (!selectedAssignee.value) return
  emit('assign', selectedAssignee.value)
}
</script>

<template>
  <div class="workflow-actions">
    <AppButton
      v-if="showValidate"
      variant="primary"
      size="sm"
      :disabled="busy"
      @click="emit('validateCreation')"
    >
      {{ t('tma.action_validate') }}
    </AppButton>
    <div v-if="showAssign && users?.length" class="workflow-actions__assign">
      <label class="workflow-actions__label" for="assignee-select">{{ t('tma.action_assign') }}</label>
      <select id="assignee-select" v-model="selectedAssignee" class="workflow-actions__select" :disabled="busy">
        <option v-for="u in users" :key="u.id" :value="u.id">{{ u.label }}</option>
      </select>
      <AppButton variant="secondary" size="sm" :disabled="busy || !selectedAssignee" @click="onAssign">
        {{ t('tma.action_assign_submit') }}
      </AppButton>
    </div>
    <AppButton
      v-if="showTakeOver"
      variant="secondary"
      size="sm"
      :disabled="busy"
      @click="emit('takeOver')"
    >
      {{ t('tma.action_take_over') }}
    </AppButton>
    <AppButton
      v-if="showResolve"
      variant="primary"
      size="sm"
      :disabled="busy"
      @click="emit('resolve')"
    >
      {{ t('tma.action_resolve') }}
    </AppButton>
    <div v-if="showReopen" class="workflow-actions__reopen">
      <AppInput id="reopen-reason" v-model="reopenReason" :label="t('tma.reopen_reason')" />
      <AppButton variant="secondary" size="sm" :disabled="busy || !reopenReason.trim()" @click="emit('reopen', reopenReason)">
        {{ t('tma.action_reopen') }}
      </AppButton>
    </div>
  </div>
</template>

<style scoped>
.workflow-actions {
  display: flex;
  flex-wrap: wrap;
  gap: var(--kore-space-sm);
  align-items: flex-end;
}

.workflow-actions__assign {
  display: grid;
  gap: var(--kore-space-xs);
  width: 100%;
  max-width: var(--kore-form-max);
}

.workflow-actions__label {
  font-size: var(--kore-text-small);
  color: var(--kore-text-muted);
}

.workflow-actions__select {
  width: 100%;
  padding: 0.5rem 0.75rem;
  border: 1px solid var(--kore-border);
  border-radius: var(--kore-radius-md);
  background: var(--kore-bg);
  color: var(--kore-text);
}

.workflow-actions__reopen {
  display: grid;
  gap: var(--kore-space-sm);
  width: 100%;
  max-width: var(--kore-form-max);
}
</style>
