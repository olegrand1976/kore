<template>
  <div class="idp-page">
    <AppPageHeader :title="$t('idp.title')" :subtitle="$t('idp.subtitle')">
      <template #actions>
        <AppButton
          v-if="guideRef?.dismissed"
          variant="ghost"
          size="sm"
          type="button"
          @click="guideRef?.showAgain()"
        >
          {{ $t('guides.show') }}
        </AppButton>
      </template>
    </AppPageHeader>

    <AppSectionGuide ref="guideRef" guide-key="admin.identity" />

    <AppCard padding="lg">
      <div class="settings-howto" role="note">
        <p class="settings-howto__title">{{ $t('idp.howto.title') }}</p>
        <ol class="settings-howto__list settings-howto__list--ordered">
          <li>
            <i18n-t keypath="idp.howto.step_console" tag="span">
              <template #link>
                <a
                  href="https://console.cloud.google.com/apis/credentials"
                  target="_blank"
                  rel="noopener noreferrer"
                >{{ $t('idp.links.google_console') }}</a>
              </template>
            </i18n-t>
          </li>
          <li>
            <i18n-t keypath="idp.howto.step_create" tag="span">
              <template #type>
                <strong>{{ $t('idp.howto_type_web') }}</strong>
              </template>
              <template #link>
                <a
                  href="https://support.google.com/cloud/answer/6158849"
                  target="_blank"
                  rel="noopener noreferrer"
                >{{ $t('idp.links.google_oauth_doc') }}</a>
              </template>
            </i18n-t>
          </li>
          <li>{{ $t('idp.howto.step_redirect') }}</li>
          <li>{{ $t('idp.howto.step_copy') }}</li>
          <li>{{ $t('idp.howto.step_paste') }}</li>
          <li>{{ $t('idp.howto.step_enable') }}</li>
        </ol>
      </div>

      <div class="idp-redirect">
        <label class="idp-redirect__label" for="idp-redirect-uri">{{ $t('idp.redirect_uri') }}</label>
        <div class="idp-redirect__row">
          <code id="idp-redirect-uri" class="idp-redirect__value">{{ redirectUri }}</code>
          <AppButton variant="ghost" size="sm" type="button" @click="copyRedirectUri">
            {{ copied ? $t('idp.copied') : $t('idp.copy') }}
          </AppButton>
        </div>
        <p v-if="copyError" class="settings-hint settings-hint--error" role="alert">{{ copyError }}</p>
        <p class="settings-hint">{{ $t('idp.redirect_uri_hint') }}</p>
      </div>

      <form class="idp-form" @submit.prevent="save">
        <fieldset class="idp-form__section">
          <legend class="idp-form__legend">{{ $t('idp.section_credentials') }}</legend>

          <AppInput id="idp-client-id" v-model="form.clientId" required>
            <template #label>
              <span class="settings-labelRow">
                <span>{{ $t('idp.client_id') }}</span>
                <AppTooltip :button-label="$t('common.info')">
                  {{ $t('idp.tooltip.client_id') }}
                </AppTooltip>
              </span>
            </template>
          </AppInput>
          <p class="settings-hint settings-hint--tight">{{ $t('idp.hint.client_id') }}</p>

          <AppInput id="idp-client-secret" v-model="form.clientSecret" type="password">
            <template #label>
              <span class="settings-labelRow">
                <span>{{ $t('idp.client_secret') }}</span>
                <AppTooltip :button-label="$t('common.info')">
                  {{ $t('idp.tooltip.client_secret') }}
                </AppTooltip>
              </span>
            </template>
          </AppInput>
          <p class="settings-hint settings-hint--tight">
            {{ $t('idp.hint.client_secret') }} {{ $t('idp.client_secret_keep') }}
          </p>
        </fieldset>

        <details class="idp-form__details">
          <summary class="idp-form__legend">{{ $t('idp.section_advanced') }}</summary>
          <p class="settings-hint idp-form__advanced-note">
            {{ $t('idp.section_advanced_note') }}
            <a
              href="https://developers.google.com/identity/openid-connect/openid-connect"
              target="_blank"
              rel="noopener noreferrer"
            >{{ $t('idp.links.google_oidc_doc') }}</a>
          </p>

          <AppInput id="idp-name" v-model="form.name" required>
            <template #label>
              <span class="settings-labelRow">
                <span>{{ $t('idp.name') }}</span>
                <AppTooltip :button-label="$t('common.info')">
                  {{ $t('idp.tooltip.name') }}
                </AppTooltip>
              </span>
            </template>
          </AppInput>

          <AppInput
            id="idp-issuer"
            v-model="form.issuer"
            placeholder="https://accounts.google.com"
            required
          >
            <template #label>
              <span class="settings-labelRow">
                <span>{{ $t('idp.issuer') }}</span>
                <AppTooltip :button-label="$t('common.info')">
                  {{ $t('idp.tooltip.issuer') }}
                </AppTooltip>
              </span>
            </template>
          </AppInput>

          <AppInput id="idp-jwks" v-model="form.jwksUri">
            <template #label>
              <span class="settings-labelRow">
                <span>{{ $t('idp.jwks_uri') }}</span>
                <AppTooltip :button-label="$t('common.info')">
                  {{ $t('idp.tooltip.jwks_uri') }}
                </AppTooltip>
              </span>
            </template>
          </AppInput>

          <AppInput id="idp-scopes" v-model="form.scopes">
            <template #label>
              <span class="settings-labelRow">
                <span>{{ $t('idp.scopes') }}</span>
                <AppTooltip :button-label="$t('common.info')">
                  {{ $t('idp.tooltip.scopes') }}
                </AppTooltip>
              </span>
            </template>
          </AppInput>
        </details>

        <fieldset class="idp-form__section">
          <legend class="idp-form__legend">{{ $t('idp.section_activation') }}</legend>

          <div class="settings-field">
            <label for="idp-default-profile" class="settings-labelRow">
              <span>{{ $t('idp.default_profile') }}</span>
              <AppTooltip :button-label="$t('common.info')">
                {{ $t('idp.tooltip.default_profile') }}
              </AppTooltip>
            </label>
            <select id="idp-default-profile" v-model="form.defaultProfile" required>
              <option value="Collaborateur">{{ $t('idp.profile_collaborateur') }}</option>
              <option value="Administrateur">{{ $t('idp.profile_administrateur') }}</option>
            </select>
            <p class="settings-hint">{{ $t('idp.hint.default_profile') }}</p>
          </div>

          <label class="idp-form__toggle">
            <input v-model="form.enabled" type="checkbox" />
            <span class="settings-labelRow">
              <span>{{ $t('idp.enabled') }}</span>
              <AppTooltip :button-label="$t('common.info')">
                {{ $t('idp.tooltip.enabled') }}
              </AppTooltip>
            </span>
          </label>
          <p class="settings-hint settings-hint--tight">{{ $t('idp.hint.enabled') }}</p>
        </fieldset>

        <AppButton type="submit" variant="primary" :loading="pending">{{ $t('common.save') }}</AppButton>
        <p
          v-if="message"
          class="settings-flash"
          :class="{ 'settings-flash--error': isError }"
          :role="isError ? 'alert' : 'status'"
        >
          {{ message }}
        </p>
      </form>
    </AppCard>
  </div>
</template>

<script setup lang="ts">
definePageMeta({ layout: 'default', middleware: 'admin' })

const { t } = useI18n()
const pending = ref(false)
const message = ref('')
const isError = ref(false)
const idpId = ref(crypto.randomUUID())
const guideRef = ref<{ showAgain: () => void; dismissed: boolean } | null>(null)
const redirectUri = ref('')
const copied = ref(false)
const copyError = ref('')
let copyTimeout: ReturnType<typeof setTimeout> | undefined

const form = reactive({
  name: 'Google',
  issuer: 'https://accounts.google.com',
  clientId: '',
  clientSecret: '',
  jwksUri: 'https://www.googleapis.com/oauth2/v3/certs',
  scopes: 'openid profile email',
  defaultProfile: 'Collaborateur',
  enabled: false
})

onMounted(() => {
  redirectUri.value = `${window.location.origin}/login`
})

onBeforeUnmount(() => {
  if (copyTimeout) clearTimeout(copyTimeout)
})

const copyRedirectUri = async () => {
  if (!redirectUri.value) return
  copyError.value = ''
  try {
    await navigator.clipboard.writeText(redirectUri.value)
    copied.value = true
    if (copyTimeout) clearTimeout(copyTimeout)
    copyTimeout = setTimeout(() => { copied.value = false }, 2000)
  } catch {
    copyError.value = t('idp.copy_failed')
  }
}

const { data } = await useFetch<Array<Record<string, unknown>>>('/api/admin/identity-providers')
watch(data, (items) => {
  const first = items?.[0]
  if (!first) return
  idpId.value = String(first.id)
  form.name = String(first.name ?? form.name)
  form.issuer = String(first.issuer ?? '')
  form.clientId = String(first.clientId ?? '')
  form.jwksUri = String(first.jwksUri ?? '')
  form.scopes = String(first.scopes ?? form.scopes)
  form.defaultProfile = String(first.defaultProfile ?? form.defaultProfile)
  form.enabled = Boolean(first.enabled)
}, { immediate: true })

const save = async () => {
  pending.value = true
  message.value = ''
  isError.value = false
  try {
    await $fetch(`/api/admin/identity-providers/${idpId.value}`, {
      method: 'PUT',
      body: { ...form }
    })
    message.value = t('idp.saved')
  } catch {
    isError.value = true
    message.value = t('common.error')
  } finally {
    pending.value = false
  }
}
</script>

<style scoped>
.idp-form {
  display: grid;
  gap: var(--kore-space-md);
  max-width: var(--kore-form-wide-max);
}

.idp-form__section {
  display: grid;
  gap: var(--kore-space-md);
  border: none;
  padding: 0;
  margin: 0;
}

.idp-form__legend {
  font-size: var(--kore-text-small);
  font-weight: 600;
  color: var(--kore-text);
  margin-bottom: var(--kore-space-xs);
}

.idp-form__details {
  display: grid;
  gap: var(--kore-space-md);
}

.idp-form__details > summary {
  cursor: pointer;
  list-style: none;
}

.idp-form__details > summary::-webkit-details-marker {
  display: none;
}

.idp-form__details > summary::before {
  content: '▸ ';
  display: inline-block;
  transition: transform 0.15s ease;
}

.idp-form__details[open] > summary::before {
  transform: rotate(90deg);
}

.idp-form__advanced-note {
  margin: 0;
}

.idp-form__toggle {
  display: flex;
  align-items: center;
  gap: var(--kore-space-sm);
}

.settings-flash {
  margin-top: var(--kore-space-sm);
  padding: 0.75rem 1rem;
  border-radius: var(--kore-radius-sm);
  background: var(--kore-bg-elevated);
  font-size: var(--kore-text-small);
}

.settings-flash--error {
  color: var(--kore-status-danger);
  border: 1px solid var(--kore-status-danger);
}

.idp-redirect {
  margin-bottom: var(--kore-space-lg);
  max-width: var(--kore-form-wide-max);
}

.idp-redirect__label {
  display: block;
  font-size: var(--kore-text-small);
  font-weight: 600;
  margin-bottom: var(--kore-space-xs);
}

.idp-redirect__row {
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  gap: var(--kore-space-sm);
}

.idp-redirect__value {
  flex: 1 1 12rem;
  padding: 0.5rem 0.75rem;
  border: 1px solid var(--kore-border);
  border-radius: var(--kore-radius-sm);
  background: var(--kore-bg);
  color: var(--kore-text);
  font-size: var(--kore-text-small);
  overflow-wrap: anywhere;
}

.settings-howto {
  margin: 0 0 var(--kore-space-lg);
  padding: var(--kore-space-md);
  border: 1px solid var(--kore-border);
  border-radius: var(--kore-radius-md);
  background: var(--kore-bg-elevated);
  max-width: var(--kore-form-wide-max);
}

.settings-howto__title {
  margin: 0 0 var(--kore-space-sm);
  font-size: var(--kore-text-small);
  font-weight: 600;
}

.settings-howto__list {
  margin: 0;
  padding-left: 1.25rem;
  color: var(--kore-text-muted);
  font-size: var(--kore-text-small);
}

.settings-howto__list--ordered {
  display: flex;
  flex-direction: column;
  gap: 0.35rem;
}

.settings-howto__list a {
  color: var(--kore-primary);
}

.settings-labelRow {
  display: inline-flex;
  align-items: center;
  gap: 0.35rem;
}

.settings-labelRow :deep(.app-tooltip__button) {
  width: 1.75rem;
  height: 1.75rem;
}

.settings-hint {
  margin: 0;
  color: var(--kore-text-muted);
  font-size: var(--kore-text-small);
  line-height: 1.35;
}

.settings-hint--tight {
  margin-top: calc(var(--kore-space-md) * -1);
}

.settings-hint--error {
  color: var(--kore-status-danger);
}

.settings-field {
  display: flex;
  flex-direction: column;
  gap: 0.35rem;
}

.settings-field label,
.settings-field select {
  font-size: var(--kore-text-small);
}

.settings-field select {
  padding: 0.5rem 0.75rem;
  border: 1px solid var(--kore-border);
  border-radius: var(--kore-radius-sm);
  background: var(--kore-bg);
  color: var(--kore-text);
}

@media (max-width: 768px) {
  .idp-redirect__row {
    flex-direction: column;
    align-items: stretch;
  }

  .idp-redirect__value {
    flex: none;
    width: 100%;
  }
}
</style>
