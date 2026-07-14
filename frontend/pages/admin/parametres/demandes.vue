<script setup lang="ts">
import type { ChannelsEnabled } from '~/composables/useRequestSettings'

definePageMeta({ layout: 'default', middleware: 'admin' })

const { t } = useI18n()
const { extractFetchError } = useApiError()
const { settings, fetchSettings, saveSettings } = useRequestSettings()

const channels = ref<ChannelsEnabled>({ tma: true, support: true, maintenance: true })
const guidesEnabled = ref(true)
const saving = ref(false)
const message = ref('')
const isError = ref(false)

onMounted(async () => {
  await fetchSettings()
  if (settings.value) {
    channels.value = { ...settings.value.channelsEnabled }
    guidesEnabled.value = settings.value.guidesEnabled
  }
})

const save = async () => {
  saving.value = true
  message.value = ''
  isError.value = false
  try {
    await saveSettings({
      channelsEnabled: { ...channels.value },
      guidesEnabled: guidesEnabled.value
    })
    message.value = t('settings.demandes.saved')
  } catch (e) {
    isError.value = true
    message.value = extractFetchError(e, t('settings.demandes.save_error'))
  } finally {
    saving.value = false
  }
}
</script>

<template>
  <div>
    <AppSectionGuide guide-key="admin.demandes" />

    <AppCard padding="lg">
      <form class="demandes-settings" @submit.prevent="save">
        <fieldset class="demandes-settings__fieldset">
          <legend>{{ t('settings.demandes.channels_title') }}</legend>
          <p class="demandes-settings__hint">{{ t('settings.demandes.channels_hint') }}</p>
          <label class="demandes-settings__check">
            <input v-model="channels.tma" type="checkbox">
            {{ t('nav.tma') }}
          </label>
          <label class="demandes-settings__check">
            <input v-model="channels.support" type="checkbox">
            {{ t('nav.support') }}
          </label>
          <label class="demandes-settings__check">
            <input v-model="channels.maintenance" type="checkbox">
            {{ t('nav.maintenance') }}
          </label>
        </fieldset>

        <label class="demandes-settings__check">
          <input v-model="guidesEnabled" type="checkbox">
          {{ t('settings.demandes.guides_enabled') }}
        </label>
        <p class="demandes-settings__hint">{{ t('settings.demandes.guides_hint') }}</p>

        <AppButton variant="primary" size="sm" type="submit" :disabled="saving">
          {{ t('common.save') }}
        </AppButton>
        <p v-if="message" class="flash" :class="{ 'flash--error': isError }" role="status">{{ message }}</p>
      </form>
    </AppCard>
  </div>
</template>

<style scoped>
.demandes-settings {
  display: flex;
  flex-direction: column;
  gap: var(--kore-space-md);
  max-width: 36rem;
}

.demandes-settings__fieldset {
  border: 1px solid var(--kore-border);
  border-radius: var(--kore-radius-md);
  padding: var(--kore-space-md);
  display: flex;
  flex-direction: column;
  gap: var(--kore-space-sm);
}

.demandes-settings__check {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  font-size: var(--kore-text-small);
}

.demandes-settings__hint {
  margin: 0;
  font-size: var(--kore-text-small);
  color: var(--kore-text-muted);
}

.flash {
  margin: 0;
  font-size: var(--kore-text-small);
  color: var(--kore-success);
}

.flash--error {
  color: var(--kore-danger);
}
</style>
