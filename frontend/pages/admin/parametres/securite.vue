<template>
  <div>
    <AppCard padding="lg" class="settings-toolbar">
      <div class="settings-toolbar__field">
        <label for="societe-select">{{ $t('settings.securite.societe') }}</label>
        <select id="societe-select" v-model="selectedSocieteId">
          <option v-for="s in societes" :key="s.id" :value="s.id">{{ s.raisonSociale }}</option>
        </select>
      </div>
    </AppCard>

    <AppCard v-if="selectedSocieteId" padding="lg">
      <form class="securite-form" @submit.prevent="save">
        <p class="hint">{{ $t('settings.securite.intro') }}</p>

        <label class="securite-form__check">
          <input v-model="totpDefaultEnabled" type="checkbox">
          {{ $t('settings.securite.totp_default_enabled') }}
        </label>
        <p v-if="totpDefaultEnabled" class="warn" role="note">{{ $t('settings.securite.totp_default_warning') }}</p>

        <label class="securite-form__check">
          <input v-model="totpUserConfigurable" type="checkbox" :disabled="!totpDefaultEnabled">
          {{ $t('settings.securite.totp_user_configurable') }}
        </label>
        <p v-if="!totpDefaultEnabled" class="hint">{{ $t('settings.securite.totp_user_configurable_hint') }}</p>

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
  totpDefaultEnabled?: boolean
  totpUserConfigurable?: boolean
}

const societes = ref<SocieteRow[]>([])
const selectedSocieteId = ref('')
const totpDefaultEnabled = ref(false)
const totpUserConfigurable = ref(true)
const saving = ref(false)
const message = ref('')
const isError = ref(false)

const applyRow = (row?: SocieteRow) => {
  totpDefaultEnabled.value = row?.totpDefaultEnabled ?? false
  totpUserConfigurable.value = row?.totpUserConfigurable ?? true
}

const loadSocietes = async () => {
  const res = await $fetch<{ data: SocieteRow[] }>('/api/org/societes')
  societes.value = (res.data ?? []).map((s) => ({
    id: s.id,
    raisonSociale: s.raisonSociale,
    totpDefaultEnabled: s.totpDefaultEnabled ?? false,
    totpUserConfigurable: s.totpUserConfigurable ?? true
  }))
  if (!selectedSocieteId.value && societes.value.length > 0) {
    selectedSocieteId.value = societes.value[0].id
  }
  applyRow(societes.value.find((s) => s.id === selectedSocieteId.value))
}

watch(selectedSocieteId, (id) => {
  applyRow(societes.value.find((s) => s.id === id))
})

watch(totpDefaultEnabled, (enabled) => {
  if (!enabled) {
    totpUserConfigurable.value = true
  }
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
        totpDefaultEnabled: totpDefaultEnabled.value,
        totpUserConfigurable: totpUserConfigurable.value
      }
    })
    message.value = t('settings.securite.saved')
    await loadSocietes()
  } catch {
    isError.value = true
    message.value = t('settings.securite.error_save')
  } finally {
    saving.value = false
  }
}

onMounted(loadSocietes)
</script>

<style scoped>
.settings-toolbar {
  margin-bottom: var(--kore-space-lg);
}

.settings-toolbar__field {
  display: grid;
  gap: var(--kore-space-xs);
}

.settings-toolbar__field select {
  max-width: var(--kore-form-wide-max);
  padding: 0.5rem 0.75rem;
  border: 1px solid var(--kore-border);
  border-radius: var(--kore-radius-sm);
  background: var(--kore-bg);
  color: var(--kore-text);
}

.securite-form {
  display: grid;
  gap: var(--kore-space-md);
  max-width: var(--kore-form-wide-max);
}

.securite-form__check {
  display: flex;
  align-items: flex-start;
  gap: var(--kore-space-sm);
  cursor: pointer;
}

.hint {
  margin: 0;
  color: var(--kore-text-muted);
  font-size: var(--kore-text-small);
}

.warn {
  margin: 0;
  color: var(--kore-warning, var(--kore-brand-gold));
  font-size: var(--kore-text-small);
}

.flash {
  margin: 0;
  font-size: var(--kore-text-small);
}

.flash--error {
  color: var(--kore-danger);
}
</style>
