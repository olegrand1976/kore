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
      <fieldset class="idp-form__section idp-form__section--provider">
        <legend class="idp-form__legend">{{ $t('idp.provider_label') }}</legend>
        <p class="settings-hint">{{ $t('idp.provider_hint') }}</p>
        <div class="idp-provider-options" role="radiogroup" :aria-label="$t('idp.provider_label')">
          <label
            v-for="option in providerOptions"
            :key="option.value"
            class="idp-provider-option"
            :class="{ 'idp-provider-option--active': providerPreset === option.value }"
          >
            <input
              v-model="providerPreset"
              class="idp-provider-option__input"
              type="radio"
              name="idp-provider"
              :value="option.value"
            />
            <span class="idp-provider-option__title">{{ option.label }}</span>
            <span class="idp-provider-option__desc">{{ option.description }}</span>
          </label>
        </div>
      </fieldset>

      <div class="settings-howto" role="note">
        <p class="settings-howto__title">{{ howtoTitle }}</p>
        <ol class="settings-howto__list settings-howto__list--ordered">
          <li v-for="(step, index) in howtoSteps" :key="index">
            <i18n-t v-if="step.key" :keypath="step.key" tag="span">
              <template v-if="step.slots?.link" #link>
                <a :href="step.slots.link.href" target="_blank" rel="noopener noreferrer">
                  {{ $t(step.slots.link.labelKey) }}
                </a>
              </template>
              <template v-if="step.slots?.type" #type>
                <strong>{{ $t(step.slots.type) }}</strong>
              </template>
            </i18n-t>
            <span v-else>{{ step.text }}</span>
          </li>
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
        <p class="settings-hint">{{ redirectUriHint }}</p>
      </div>

      <form class="idp-form" @submit.prevent="save">
        <fieldset class="idp-form__section">
          <legend class="idp-form__legend">{{ credentialsSectionTitle }}</legend>

          <AppInput
            v-if="providerPreset === 'azure'"
            id="idp-azure-tenant"
            v-model="azureTenantId"
            required
          >
            <template #label>
              <span class="settings-labelRow">
                <span>{{ $t('idp.azure_tenant_id') }}</span>
                <AppTooltip :button-label="$t('common.info')">
                  {{ $t('idp.tooltip.azure_tenant_id') }}
                </AppTooltip>
              </span>
            </template>
          </AppInput>
          <p v-if="providerPreset === 'azure'" class="settings-hint settings-hint--tight">
            {{ $t('idp.hint.azure_tenant_id') }}
          </p>

          <AppInput id="idp-client-id" v-model="form.clientId" required>
            <template #label>
              <span class="settings-labelRow">
                <span>{{ $t('idp.client_id') }}</span>
                <AppTooltip :button-label="$t('common.info')">
                  {{ clientIdTooltip }}
                </AppTooltip>
              </span>
            </template>
          </AppInput>
          <p class="settings-hint settings-hint--tight">{{ clientIdHint }}</p>

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

        <label class="idp-form__toggle idp-form__toggle--advanced">
          <input v-model="showAdvancedOidc" type="checkbox" />
          <span>{{ $t('idp.show_advanced') }}</span>
        </label>

        <details v-if="showAdvancedOidc" class="idp-form__details" open>
          <summary class="idp-form__legend">{{ $t('idp.section_advanced') }}</summary>
          <p class="settings-hint idp-form__advanced-note">{{ $t('idp.section_advanced_note') }}</p>

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

          <AppInput id="idp-issuer" v-model="form.issuer" required>
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

        <div class="idp-form__actions">
          <AppButton type="submit" variant="primary" :loading="pending">{{ $t('common.save') }}</AppButton>
          <AppButton
            variant="ghost"
            type="button"
            :disabled="!canTestConnection"
            @click="testConnection"
          >
            {{ $t('idp.test_connection') }}
          </AppButton>
        </div>
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
type ProviderPreset = 'google' | 'azure'

type HowtoStep = {
  key?: string
  text?: string
  slots?: {
    link?: { href: string; labelKey: string }
    type?: string
  }
}

const GOOGLE_PRESET = {
  name: 'Google',
  issuer: 'https://accounts.google.com',
  jwksUri: 'https://www.googleapis.com/oauth2/v3/certs',
  scopes: 'openid profile email'
} as const

const AZURE_PRESET = {
  name: 'Microsoft',
  scopes: 'openid profile email'
} as const

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
const providerPreset = ref<ProviderPreset>('google')
const azureTenantId = ref('')
const showAdvancedOidc = ref(false)
const isHydrating = ref(true)
let copyTimeout: ReturnType<typeof setTimeout> | undefined

const form = reactive({
  name: GOOGLE_PRESET.name,
  issuer: GOOGLE_PRESET.issuer,
  clientId: '',
  clientSecret: '',
  jwksUri: GOOGLE_PRESET.jwksUri,
  scopes: GOOGLE_PRESET.scopes,
  defaultProfile: 'Collaborateur',
  enabled: false
})

const providerOptions = computed(() => [
  {
    value: 'google' as const,
    label: t('idp.provider_google'),
    description: t('idp.provider_google_desc')
  },
  {
    value: 'azure' as const,
    label: t('idp.provider_azure'),
    description: t('idp.provider_azure_desc')
  }
])

const howtoTitle = computed(() =>
  providerPreset.value === 'azure' ? t('idp.howto.azure_title') : t('idp.howto.google_title')
)

const howtoSteps = computed<HowtoStep[]>(() => {
  if (providerPreset.value === 'azure') {
    return [
      {
        key: 'idp.howto.azure_step_portal',
        slots: {
          link: {
            href: 'https://portal.azure.com/#view/Microsoft_AAD_RegisteredApps/ApplicationsListBlade',
            labelKey: 'idp.links.azure_portal'
          }
        }
      },
      {
        key: 'idp.howto.azure_step_create',
        slots: {
          type: 'idp.howto_type_web',
          link: {
            href: 'https://learn.microsoft.com/en-us/entra/identity-platform/quickstart-register-app',
            labelKey: 'idp.links.azure_doc'
          }
        }
      },
      { key: 'idp.howto.step_redirect' },
      { key: 'idp.howto.azure_step_copy' },
      { key: 'idp.howto.step_paste' },
      { key: 'idp.howto.step_enable' }
    ]
  }

  return [
    {
      key: 'idp.howto.google_step_console',
      slots: {
        link: {
          href: 'https://console.cloud.google.com/apis/credentials',
          labelKey: 'idp.links.google_console'
        }
      }
    },
    {
      key: 'idp.howto.google_step_create',
      slots: {
        type: 'idp.howto_type_web',
        link: {
          href: 'https://support.google.com/cloud/answer/6158849',
          labelKey: 'idp.links.google_oauth_doc'
        }
      }
    },
    { key: 'idp.howto.step_redirect' },
    { key: 'idp.howto.google_step_copy' },
    { key: 'idp.howto.step_paste' },
    { key: 'idp.howto.step_enable' }
  ]
})

const credentialsSectionTitle = computed(() =>
  providerPreset.value === 'azure' ? t('idp.section_credentials_azure') : t('idp.section_credentials_google')
)

const redirectUriHint = computed(() =>
  providerPreset.value === 'azure' ? t('idp.redirect_uri_hint_azure') : t('idp.redirect_uri_hint_google')
)

const clientIdHint = computed(() =>
  providerPreset.value === 'azure' ? t('idp.hint.client_id_azure') : t('idp.hint.client_id_google')
)

const clientIdTooltip = computed(() =>
  providerPreset.value === 'azure' ? t('idp.tooltip.client_id_azure') : t('idp.tooltip.client_id_google')
)

const canTestConnection = computed(() => form.enabled && form.clientId.trim().length > 0)

const detectProvider = (issuer: string): ProviderPreset =>
  issuer.toLowerCase().includes('microsoftonline.com') ? 'azure' : 'google'

const extractAzureTenantId = (issuer: string): string => {
  const match = issuer.match(/login\.microsoftonline\.com\/([^/]+)/i)
  return match?.[1] ?? ''
}

const buildAzureIssuer = (tenantId: string) => {
  const tid = tenantId.trim()
  if (!tid) return ''
  return `https://login.microsoftonline.com/${tid}/v2.0`
}

const buildAzureJwks = (tenantId: string) => {
  const tid = tenantId.trim()
  if (!tid) return ''
  return `https://login.microsoftonline.com/${tid}/discovery/v2.0/keys`
}

const applyProviderPreset = (preset: ProviderPreset, preserveCredentials = false) => {
  if (preset === 'azure') {
    form.name = AZURE_PRESET.name
    form.scopes = AZURE_PRESET.scopes
    if (!preserveCredentials) {
      azureTenantId.value = ''
      form.issuer = ''
      form.jwksUri = ''
    }
    return
  }

  form.name = GOOGLE_PRESET.name
  form.issuer = GOOGLE_PRESET.issuer
  form.jwksUri = GOOGLE_PRESET.jwksUri
  form.scopes = GOOGLE_PRESET.scopes
  azureTenantId.value = ''
}

watch(providerPreset, (preset) => {
  if (isHydrating.value) return
  applyProviderPreset(preset)
  showAdvancedOidc.value = false
})

watch(azureTenantId, (tenantId) => {
  if (providerPreset.value !== 'azure') return
  form.issuer = buildAzureIssuer(tenantId)
  form.jwksUri = buildAzureJwks(tenantId)
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
  if (!first) {
    isHydrating.value = false
    return
  }

  idpId.value = String(first.id)
  form.name = String(first.name ?? form.name)
  form.issuer = String(first.issuer ?? '')
  form.clientId = String(first.clientId ?? '')
  form.jwksUri = String(first.jwksUri ?? '')
  form.scopes = String(first.scopes ?? form.scopes)
  form.defaultProfile = String(first.defaultProfile ?? form.defaultProfile)
  form.enabled = Boolean(first.enabled)

  const preset = detectProvider(form.issuer)
  providerPreset.value = preset
  if (preset === 'azure') {
    azureTenantId.value = extractAzureTenantId(form.issuer)
  }

  isHydrating.value = false
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

const testConnection = async () => {
  await navigateTo('/login', { open: { target: '_blank' } })
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

.idp-form__section--provider {
  margin-bottom: var(--kore-space-lg);
}

.idp-form__legend {
  font-size: var(--kore-text-small);
  font-weight: 600;
  color: var(--kore-text);
  margin-bottom: var(--kore-space-xs);
}

.idp-provider-options {
  display: grid;
  gap: var(--kore-space-sm);
}

.idp-provider-option {
  display: grid;
  gap: 0.2rem;
  padding: var(--kore-space-md);
  border: 1px solid var(--kore-border);
  border-radius: var(--kore-radius-md);
  background: var(--kore-bg);
  cursor: pointer;
}

.idp-provider-option--active {
  border-color: var(--kore-primary);
  background: var(--kore-bg-elevated);
}

.idp-provider-option__input {
  position: absolute;
  opacity: 0;
  pointer-events: none;
}

.idp-provider-option__title {
  font-size: var(--kore-text-small);
  font-weight: 600;
}

.idp-provider-option__desc {
  font-size: var(--kore-text-small);
  color: var(--kore-text-muted);
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

.idp-form__toggle--advanced {
  font-size: var(--kore-text-small);
  color: var(--kore-text-muted);
}

.idp-form__actions {
  display: flex;
  flex-wrap: wrap;
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
  .idp-redirect__row,
  .idp-form__actions {
    flex-direction: column;
    align-items: stretch;
  }

  .idp-redirect__value {
    flex: none;
    width: 100%;
  }

  .idp-form__actions :deep(.app-button) {
    width: 100%;
  }
}
</style>
