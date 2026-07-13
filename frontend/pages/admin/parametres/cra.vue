<template>
  <div>
    <AppCard padding="lg" class="settings-toolbar">
      <div class="settings-toolbar__field">
        <label for="societe-select">{{ $t('settings.cra.societe') }}</label>
        <select id="societe-select" v-model="selectedSocieteId">
          <option v-for="s in societes" :key="s.id" :value="s.id">{{ s.raisonSociale }}</option>
        </select>
      </div>
    </AppCard>

    <AppCard v-if="selectedSocieteId" padding="lg">
      <form class="cra-settings-form" @submit.prevent="save">
        <label for="week-start-day">{{ $t('settings.cra.week_start_day') }}</label>
        <select id="week-start-day" v-model.number="weekStartDay">
          <option v-for="opt in weekDayOptions" :key="opt.value" :value="opt.value">{{ opt.label }}</option>
        </select>

        <label for="day-capacity">{{ $t('settings.cra.day_capacity') }}</label>
        <input id="day-capacity" v-model.number="dayCapacityMinutes" type="number" min="60" max="1440" step="30">

        <label for="submit-policy">{{ $t('settings.cra.week_submit_policy') }}</label>
        <select id="submit-policy" v-model="weekSubmitPolicy">
          <option value="warn">{{ $t('settings.cra.submit_policy_warn') }}</option>
          <option value="block">{{ $t('settings.cra.submit_policy_block') }}</option>
          <option value="none">{{ $t('settings.cra.submit_policy_none') }}</option>
        </select>

        <label class="cra-settings-form__check">
          <input v-model="craMailAuto" type="checkbox">
          {{ $t('settings.cra.cra_mail_auto') }}
        </label>

        <p class="hint">{{ $t('settings.cra.week_start_hint') }}</p>
        <AppButton variant="primary" size="sm" type="submit" :disabled="saving">
          {{ $t('common.save') }}
        </AppButton>
        <p v-if="message" class="flash" :class="{ 'flash--error': isError }" role="status">{{ message }}</p>
      </form>
    </AppCard>
  </div>
</template>

<script setup lang="ts">
definePageMeta({ layout: 'default', middleware: 'admin' })

const { t } = useI18n()

type SocieteRow = {
  id: string
  raisonSociale: string
  weekStartDay?: number
  dayCapacityMinutes?: number
  craMailAuto?: boolean
  weekSubmitPolicy?: string
}

const societes = ref<SocieteRow[]>([])
const selectedSocieteId = ref('')
const weekStartDay = ref(1)
const dayCapacityMinutes = ref(480)
const craMailAuto = ref(false)
const weekSubmitPolicy = ref('warn')
const saving = ref(false)
const message = ref('')
const isError = ref(false)

const weekDayOptions = computed(() => [
  { value: 1, label: t('settings.cra.weekday_mon') },
  { value: 2, label: t('settings.cra.weekday_tue') },
  { value: 3, label: t('settings.cra.weekday_wed') },
  { value: 4, label: t('settings.cra.weekday_thu') },
  { value: 5, label: t('settings.cra.weekday_fri') },
  { value: 6, label: t('settings.cra.weekday_sat') },
  { value: 0, label: t('settings.cra.weekday_sun') }
])

const applyRow = (row?: SocieteRow) => {
  weekStartDay.value = row?.weekStartDay ?? 1
  dayCapacityMinutes.value = row?.dayCapacityMinutes ?? 480
  craMailAuto.value = row?.craMailAuto ?? false
  weekSubmitPolicy.value = row?.weekSubmitPolicy ?? 'warn'
}

const loadSocietes = async () => {
  const res = await $fetch<{ data: SocieteRow[] }>('/api/org/societes')
  societes.value = (res.data ?? []).map((s) => ({
    id: s.id,
    raisonSociale: s.raisonSociale,
    weekStartDay: s.weekStartDay ?? 1,
    dayCapacityMinutes: s.dayCapacityMinutes ?? 480,
    craMailAuto: s.craMailAuto ?? false,
    weekSubmitPolicy: s.weekSubmitPolicy ?? 'warn'
  }))
  if (!selectedSocieteId.value && societes.value.length > 0) {
    selectedSocieteId.value = societes.value[0].id
  }
  applyRow(societes.value.find((s) => s.id === selectedSocieteId.value))
}

watch(selectedSocieteId, (id) => {
  applyRow(societes.value.find((s) => s.id === id))
})

const save = async () => {
  if (!selectedSocieteId.value) return
  saving.value = true
  message.value = ''
  isError.value = false
  try {
    await $fetch(`/api/org/societes/${selectedSocieteId.value}/settings`, {
      method: 'PUT',
      body: {
        weekStartDay: weekStartDay.value,
        dayCapacityMinutes: dayCapacityMinutes.value,
        craMailAuto: craMailAuto.value,
        weekSubmitPolicy: weekSubmitPolicy.value
      }
    })
    message.value = t('settings.cra.saved')
    await loadSocietes()
  } catch {
    message.value = t('settings.cra.save_error')
    isError.value = true
  } finally {
    saving.value = false
  }
}

await loadSocietes()
</script>

<style scoped>
.settings-toolbar {
  margin-bottom: var(--kore-space-lg);
}

.settings-toolbar__field {
  display: grid;
  gap: var(--kore-space-xs);
  max-width: var(--kore-form-max);
}

.cra-settings-form {
  display: grid;
  gap: var(--kore-space-md);
  max-width: var(--kore-form-max);
}

.cra-settings-form select,
.cra-settings-form input[type='number'] {
  padding: 0.625rem 0.75rem;
  border: 1px solid var(--kore-border);
  border-radius: var(--kore-radius-md);
  background: var(--kore-bg);
  color: var(--kore-text);
}

.cra-settings-form__check {
  display: flex;
  align-items: center;
  gap: var(--kore-space-sm);
}

.hint {
  margin: 0;
  font-size: var(--kore-text-small);
  color: var(--kore-text-muted);
}

.flash {
  font-size: var(--kore-text-small);
  color: var(--kore-success);
}

.flash--error {
  color: var(--kore-error);
}
</style>
