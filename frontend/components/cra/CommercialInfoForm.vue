<template>
  <AppCard padding="lg">
    <h3 class="section-title">{{ $t('cra.commercial_title') }}</h3>
    <p class="section-hint">{{ $t('cra.commercial_hint') }}</p>
    <form class="commercial-form" @submit.prevent="$emit('submit')">
      <label v-if="missions.length" for="mission-select">{{ $t('cra.mission_select') }}</label>
      <select
        v-if="missions.length"
        id="mission-select"
        v-model="local.missionId"
        :disabled="disabled"
        @change="onMissionPick"
      >
        <option value="">{{ $t('cra.mission_manual') }}</option>
        <option v-for="mission in missions" :key="mission.id" :value="mission.id">
          {{ missionLabel(mission) }}
        </option>
      </select>

      <AppInput id="client" v-model="local.client" :label="$t('cra.client')" required :disabled="identityLocked" />
      <AppInput id="mission" v-model="local.mission" :label="$t('cra.mission')" required :disabled="identityLocked" />
      <AppInput id="description" v-model="local.description" :label="optionalLabel('cra.commercial_description')" :disabled="disabled" />
      <AppInput
        v-if="manualEntry"
        id="technologies"
        v-model="technologiesText"
        :label="optionalLabel('cra.commercial_technologies')"
        :disabled="disabled"
      />
      <AppInput id="lieu" v-model="local.lieu" :label="optionalLabel('cra.commercial_lieu')" :disabled="disabled" />
      <AppInput
        v-if="manualEntry"
        id="responsable"
        v-model="local.responsableClient"
        :label="optionalLabel('cra.commercial_responsable')"
        :disabled="disabled"
      />
      <AppButton variant="primary" size="sm" type="submit" :disabled="disabled || saving">
        {{ $t('cra.save_commercial') }}
      </AppButton>
    </form>
    <p v-if="missionLoadError" class="flash flash--error" role="alert">{{ missionLoadError }}</p>
    <p v-if="message" class="flash" :class="{ 'flash--error': isError }" role="status">{{ message }}</p>
  </AppCard>
</template>

<script setup lang="ts">
import {
  isManualCommercialEntry,
  missionCommercialPatch,
  unwrapMissionPayload
} from '~/utils/craCommercial'

export type CommercialMissionOption = {
  id: string
  clientName: string
  clientId?: string
  label?: string
}

const props = defineProps<{
  client?: string
  mission?: string
  clientId?: string
  missionId?: string
  missions?: CommercialMissionOption[]
  description?: string
  technologies?: string[]
  lieu?: string
  responsableClient?: string
  disabled?: boolean
  saving?: boolean
  message?: string
  isError?: boolean
}>()

defineEmits<{ submit: [] }>()

const { t } = useI18n()
const missions = computed(() => props.missions ?? [])

const local = reactive({
  client: props.client ?? '',
  mission: props.mission ?? '',
  clientId: props.clientId ?? '',
  missionId: props.missionId ?? '',
  description: props.description ?? '',
  technologies: [...(props.technologies ?? [])],
  lieu: props.lieu ?? '',
  responsableClient: props.responsableClient ?? ''
})

const manualEntry = computed(() => isManualCommercialEntry(local.missionId))
const identityLocked = computed(() => Boolean(props.disabled) || !manualEntry.value)
const missionLoadError = ref('')

const optionalLabel = (key: string) => `${t(key)} (${t('common.optional')})`

const technologiesText = computed({
  get: () => local.technologies.join(', '),
  set: (value: string) => {
    local.technologies = value.split(',').map((s) => s.trim()).filter(Boolean)
  }
})

const missionLabel = (mission: CommercialMissionOption) => {
  const client = mission.clientName?.trim()
  const name = mission.label?.trim() || mission.id.slice(0, 8)
  return client ? `${client} — ${name}` : name
}

const applyMissionDetail = (raw: Record<string, unknown>) => {
  const patch = missionCommercialPatch(raw)
  if (patch.client) local.client = patch.client
  if (patch.clientId) local.clientId = patch.clientId
  local.technologies = patch.technologies
  local.responsableClient = patch.responsableClient
}

let missionPickToken = 0

const onMissionPick = async () => {
  const token = ++missionPickToken
  missionLoadError.value = ''
  const picked = missions.value.find((item) => item.id === local.missionId)
  if (!picked) {
    local.clientId = ''
    return
  }
  local.clientId = picked.clientId ?? ''
  if (picked.clientName) {
    local.client = picked.clientName
  }
  if (picked.label) {
    local.mission = picked.label
  }
  local.technologies = []
  local.responsableClient = ''
  try {
    const res = await $fetch(`/api/ssii/missions/${picked.id}`)
    if (token !== missionPickToken) return
    applyMissionDetail(unwrapMissionPayload(res))
  } catch {
    if (token !== missionPickToken) return
    missionLoadError.value = t('cra.commercial_mission_load_error')
  }
}

watch(
  () => [props.client, props.mission, props.clientId, props.missionId, props.description, props.technologies, props.lieu, props.responsableClient],
  () => {
    local.client = props.client ?? ''
    local.mission = props.mission ?? ''
    local.clientId = props.clientId ?? ''
    local.missionId = props.missionId ?? ''
    local.description = props.description ?? ''
    local.technologies = [...(props.technologies ?? [])]
    local.lieu = props.lieu ?? ''
    local.responsableClient = props.responsableClient ?? ''
    missionLoadError.value = ''
  }
)

defineExpose({ local })
</script>

<style scoped>
.section-title {
  margin: 0 0 var(--kore-space-xs);
  font-size: var(--kore-text-h3);
}

.section-hint {
  margin: 0 0 var(--kore-space-lg);
  font-size: var(--kore-text-small);
  color: var(--kore-text-muted);
}

.commercial-form {
  display: flex;
  flex-direction: column;
  gap: var(--kore-space-md);
  max-width: var(--kore-form-max);
}

.commercial-form select {
  width: 100%;
  padding: var(--kore-space-sm) var(--kore-space-md);
  border: 1px solid var(--kore-border);
  border-radius: var(--kore-radius-md);
  background: var(--kore-surface);
  color: var(--kore-text);
}

.flash {
  margin-top: var(--kore-space-md);
  font-size: var(--kore-text-small);
  color: var(--kore-success);
}

.flash--error { color: var(--kore-error); }

@media (max-width: 768px) {
  .commercial-form {
    max-width: none;
  }
}
</style>
