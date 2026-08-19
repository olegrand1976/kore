<template>
  <AppCard padding="lg" class="prestation-info">
    <div class="prestation-info__header">
      <div>
        <h3 class="prestation-info__title">{{ $t('cra.prestation_title') }}</h3>
        <p class="prestation-info__hint">{{ $t('cra.prestation_hint') }}</p>
      </div>
      <AppBadge :variant="isComplete ? 'success' : 'warning'">
        {{ isComplete ? $t('cra.prestation_complete') : $t('cra.prestation_incomplete') }}
      </AppBadge>
    </div>

    <form class="prestation-info__form" @submit.prevent="$emit('submit')">
      <section v-if="missions.length || local.missionId" class="prestation-info__section">
        <h4 class="prestation-info__section-title">{{ $t('cra.prestation_section_link') }}</h4>
        <label class="visually-hidden" for="mission-select">{{ $t('cra.mission_select') }}</label>
        <select
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
        <div v-if="linkedToKnownMission" class="prestation-info__links">
          <NuxtLink :to="`/missions/${local.missionId}`" class="prestation-info__link">
            {{ $t('cra.prestation_open_mission') }}
          </NuxtLink>
          <AppBadge variant="default">{{ $t('cra.prestation_from_mission') }}</AppBadge>
        </div>
      </section>

      <section class="prestation-info__section">
        <h4 class="prestation-info__section-title">{{ $t('cra.prestation_section_context') }}</h4>
        <AppInput id="client" v-model="local.client" :label="$t('cra.client')" required :disabled="identityLocked" />
        <AppInput id="mission" v-model="local.mission" :label="$t('cra.mission')" required :disabled="identityLocked" />
        <NuxtLink
          v-if="local.clientId"
          :to="`/clients/${local.clientId}`"
          class="prestation-info__link"
        >
          {{ $t('cra.prestation_open_client') }}
        </NuxtLink>
      </section>

      <section class="prestation-info__section">
        <h4 class="prestation-info__section-title">{{ $t('cra.prestation_section_pdf') }}</h4>
        <AppInput
          id="description"
          v-model="local.description"
          :label="optionalLabel('cra.prestation_description')"
          :disabled="disabled"
        />
        <AppInput
          id="lieu"
          v-model="local.lieu"
          :label="optionalLabel('cra.prestation_lieu')"
          :disabled="disabled"
        />
        <template v-if="manualEntry">
          <AppInput
            id="technologies"
            v-model="technologiesText"
            :label="optionalLabel('cra.prestation_technologies')"
            :disabled="disabled"
          />
          <AppInput
            id="responsable"
            v-model="local.responsableClient"
            :label="optionalLabel('cra.prestation_responsable')"
            :disabled="disabled"
          />
        </template>
        <template v-else>
          <div class="prestation-info__readonly">
            <p class="prestation-info__readonly-label">{{ $t('cra.prestation_technologies') }}</p>
            <p v-if="local.technologies.length" class="prestation-info__readonly-value">
              {{ local.technologies.join(', ') }}
            </p>
            <p v-else class="prestation-info__readonly-value prestation-info__readonly-value--muted">
              {{ $t('cra.prestation_technologies_empty') }}
            </p>
          </div>
          <div class="prestation-info__readonly">
            <p class="prestation-info__readonly-label">{{ $t('cra.prestation_responsable') }}</p>
            <p
              class="prestation-info__readonly-value"
              :class="{ 'prestation-info__readonly-value--muted': !local.responsableClient.trim() }"
            >
              {{ local.responsableClient.trim() || $t('cra.prestation_responsable_empty') }}
            </p>
          </div>
        </template>
        <AppButton
          class="prestation-info__submit"
          variant="primary"
          size="sm"
          type="submit"
          :disabled="disabled || saving"
        >
          {{ $t('cra.save_prestation') }}
        </AppButton>
      </section>
    </form>

    <p
      v-if="!isComplete"
      id="cra-download-hint"
      class="prestation-info__download-hint"
    >
      {{ $t('cra.download_hint') }}
    </p>
    <p v-if="missionLoadError" class="flash flash--error" role="alert">{{ missionLoadError }}</p>
    <p v-if="message" class="flash" :class="{ 'flash--error': isError }" role="status">{{ message }}</p>
  </AppCard>
</template>

<script setup lang="ts">
import {
  isKnownMissionLink,
  missionPrestationPatch,
  prestationInfoComplete,
  unwrapMissionPayload,
  type PrestationInfoFields
} from '~/utils/craPrestation'

export type PrestationMissionOption = {
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
  missions?: PrestationMissionOption[]
  description?: string
  technologies?: string[]
  lieu?: string
  responsableClient?: string
  disabled?: boolean
  saving?: boolean
  message?: string
  isError?: boolean
}>()

const emit = defineEmits<{
  submit: []
  change: [payload: PrestationInfoFields]
}>()

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

const linkedToKnownMission = computed(() => isKnownMissionLink(local.missionId, missions.value))
const manualEntry = computed(() => !linkedToKnownMission.value)
const identityLocked = computed(() => Boolean(props.disabled) || linkedToKnownMission.value)
const isComplete = computed(() => prestationInfoComplete(local.client, local.mission))
const missionLoadError = ref('')

const snapshot = (): PrestationInfoFields => ({
  client: local.client,
  mission: local.mission,
  clientId: local.clientId,
  missionId: local.missionId,
  description: local.description,
  technologies: [...local.technologies],
  lieu: local.lieu,
  responsableClient: local.responsableClient
})

const optionalLabel = (key: string) => `${t(key)} (${t('common.optional')})`

const technologiesText = computed({
  get: () => local.technologies.join(', '),
  set: (value: string) => {
    local.technologies = value.split(',').map((s) => s.trim()).filter(Boolean)
  }
})

const missionLabel = (mission: PrestationMissionOption) => {
  const client = mission.clientName?.trim()
  const name = mission.label?.trim() || mission.id.slice(0, 8)
  return client ? `${client} — ${name}` : name
}

const applyMissionDetail = (raw: Record<string, unknown>) => {
  const patch = missionPrestationPatch(raw)
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
    missionLoadError.value = t('cra.prestation_mission_load_error')
  }
}

watch(
  () => [props.client, props.mission, props.clientId, props.missionId, props.description, props.technologies, props.lieu, props.responsableClient],
  () => {
    const next = {
      client: props.client ?? '',
      mission: props.mission ?? '',
      clientId: props.clientId ?? '',
      missionId: props.missionId ?? '',
      description: props.description ?? '',
      technologies: [...(props.technologies ?? [])],
      lieu: props.lieu ?? '',
      responsableClient: props.responsableClient ?? ''
    }
    if (
      local.client === next.client &&
      local.mission === next.mission &&
      local.clientId === next.clientId &&
      local.missionId === next.missionId &&
      local.description === next.description &&
      local.lieu === next.lieu &&
      local.responsableClient === next.responsableClient &&
      local.technologies.join('\0') === next.technologies.join('\0')
    ) {
      return
    }
    local.client = next.client
    local.mission = next.mission
    local.clientId = next.clientId
    local.missionId = next.missionId
    local.description = next.description
    local.technologies = next.technologies
    local.lieu = next.lieu
    local.responsableClient = next.responsableClient
    missionLoadError.value = ''
  }
)

watch(local, () => emit('change', snapshot()), { deep: true })

defineExpose({ local, isComplete })
</script>

<style scoped>
.prestation-info__header {
  display: flex;
  flex-wrap: wrap;
  align-items: flex-start;
  justify-content: space-between;
  gap: var(--kore-space-md);
  margin-bottom: var(--kore-space-lg);
}

.prestation-info__title {
  margin: 0 0 var(--kore-space-xs);
  font-size: var(--kore-text-h3);
}

.prestation-info__hint {
  margin: 0;
  font-size: var(--kore-text-small);
  color: var(--kore-text-muted);
}

.prestation-info__form {
  display: flex;
  flex-direction: column;
  gap: var(--kore-space-lg);
}

.prestation-info__section {
  display: flex;
  flex-direction: column;
  gap: var(--kore-space-md);
  padding-top: var(--kore-space-lg);
  border-top: 1px solid var(--kore-border);
}

.prestation-info__section:first-child {
  padding-top: 0;
  border-top: none;
}

.prestation-info__section-title {
  margin: 0;
  font-size: var(--kore-text-small);
  font-weight: 600;
  color: var(--kore-text-muted);
  text-transform: uppercase;
  letter-spacing: 0.04em;
}

.prestation-info__form select {
  width: 100%;
  padding: var(--kore-space-sm) var(--kore-space-md);
  border: 1px solid var(--kore-border);
  border-radius: var(--kore-radius-md);
  background: var(--kore-surface);
  color: var(--kore-text);
}

.prestation-info__links {
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  gap: var(--kore-space-sm);
}

.prestation-info__link {
  font-size: var(--kore-text-small);
  color: var(--kore-link);
  text-decoration: underline;
}

.prestation-info__download-hint {
  margin: var(--kore-space-md) 0 0;
  font-size: var(--kore-text-caption);
  color: var(--kore-text-muted);
}

.prestation-info__readonly-label {
  margin: 0;
  font-size: var(--kore-text-caption);
  color: var(--kore-text-muted);
}

.prestation-info__readonly-value {
  margin: var(--kore-space-xs) 0 0;
  font-size: var(--kore-text-small);
}

.prestation-info__readonly-value--muted {
  color: var(--kore-text-muted);
}

.visually-hidden {
  position: absolute;
  width: 1px;
  height: 1px;
  padding: 0;
  margin: -1px;
  overflow: hidden;
  clip: rect(0, 0, 0, 0);
  white-space: nowrap;
  border: 0;
}

.flash {
  margin-top: var(--kore-space-md);
  font-size: var(--kore-text-small);
  color: var(--kore-success);
}

.flash--error { color: var(--kore-error); }

@media (max-width: 768px) {
  .prestation-info__submit {
    width: 100%;
  }
}
</style>
