<template>
  <AppModal v-model:open="open" width="md" :title-id="titleId" :aria-label="$t('cra.add_activity')">
    <h3 :id="titleId" class="modal-title">{{ $t('cra.add_activity') }}</h3>
    <form class="add-form" @submit.prevent="confirm">
      <label class="add-form__label" for="activity-type">{{ $t('cra.activity_type') }}</label>
      <select id="activity-type" v-model="selectedType" class="add-form__select">
        <option v-for="opt in typeOptions" :key="opt.value" :value="opt.value">{{ opt.label }}</option>
      </select>
      <template v-if="selectedType === 'mission'">
        <label class="add-form__label" for="activity-mission">{{ $t('cra.mission') }}</label>
        <select id="activity-mission" v-model="selectedMissionId" class="add-form__select" required>
          <option v-for="m in missions" :key="m.id" :value="m.id">{{ missionLabel(m) }}</option>
        </select>
      </template>
      <div class="add-form__actions">
        <AppButton variant="ghost" type="button" @click="open = false">{{ $t('common.cancel') }}</AppButton>
        <AppButton variant="primary" type="submit">{{ $t('cra.add_activity') }}</AppButton>
      </div>
    </form>
  </AppModal>
</template>

<script setup lang="ts">
import type { MissionSummary } from '~/composables/useCraSourceLabels'

const props = withDefaults(defineProps<{
  missions: MissionSummary[]
  taskTypes?: string[]
}>(), {
  taskTypes: () => ['manual', 'interne', 'formation', 'mission']
})

const emit = defineEmits<{
  add: [payload: { sourceType: string; sourceId: string }]
}>()

const { t } = useI18n()
const open = defineModel<boolean>('open', { default: false })
const selectedType = ref('manual')
const selectedMissionId = ref('')
const titleId = 'cra-add-activity-title'

const missionLabel = (m: MissionSummary) => m.clientName || m.id.slice(0, 8)

const typeLabels: Record<string, string> = {
  manual: 'cra.source_manual',
  interne: 'cra.source_internal',
  formation: 'cra.source_training',
  mission: 'cra.source_mission'
}

const typeOptions = computed(() => {
  const enabled = props.taskTypes.length > 0 ? props.taskTypes : ['manual', 'interne', 'formation', 'mission']
  return enabled
    .filter((code) => code !== 'mission' || props.missions.length > 0)
    .map((code) => ({
      value: code,
      label: t(typeLabels[code] ?? code, code)
    }))
})

const confirm = () => {
  if (selectedType.value === 'mission') {
    if (!selectedMissionId.value) return
    emit('add', { sourceType: 'mission', sourceId: selectedMissionId.value })
  } else {
    emit('add', { sourceType: selectedType.value, sourceId: 'default' })
  }
  open.value = false
}

watch(open, (isOpen) => {
  if (!isOpen) return
  const first = typeOptions.value[0]?.value ?? 'manual'
  selectedType.value = first
  selectedMissionId.value = props.missions[0]?.id ?? ''
})
</script>

<style scoped>
.modal-title {
  margin: 0 0 var(--kore-space-md);
  font-size: var(--kore-text-h3);
}

.add-form {
  display: grid;
  gap: var(--kore-space-md);
}

.add-form__label {
  font-size: var(--kore-text-small);
  color: var(--kore-text-muted);
}

.add-form__select {
  width: 100%;
  padding: 0.625rem 0.75rem;
  border: 1px solid var(--kore-border);
  border-radius: var(--kore-radius-md);
  background: var(--kore-bg);
  color: var(--kore-text);
}

.add-form__actions {
  display: flex;
  flex-wrap: wrap;
  gap: var(--kore-space-sm);
  justify-content: flex-end;
}
</style>
