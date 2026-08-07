<template>
  <div>
    <AppPageHeader :title="$t('org.title')" :subtitle="$t('org.subtitle')" />

    <nav class="org-tabs" role="tablist" :aria-label="$t('org.tabs_label')">
      <button
        type="button"
        role="tab"
        class="org-tab"
        :class="{ 'org-tab--active': tab === 'identite' }"
        :aria-selected="tab === 'identite'"
        @click="tab = 'identite'"
      >
        {{ $t('org.tab_identite') }}
      </button>
      <button
        type="button"
        role="tab"
        class="org-tab"
        :class="{ 'org-tab--active': tab === 'modules' }"
        :aria-selected="tab === 'modules'"
        @click="tab = 'modules'"
      >
        {{ $t('org.tab_modules') }}
      </button>
      <button
        type="button"
        role="tab"
        class="org-tab"
        :class="{ 'org-tab--active': tab === 'structure' }"
        :aria-selected="tab === 'structure'"
        @click="tab = 'structure'"
      >
        {{ $t('org.tab_structure') }}
      </button>
    </nav>

    <OrgTree v-if="tab === 'structure'" />

    <AppCard v-else-if="tab === 'modules'" padding="lg" class="org-modules">
      <form class="org-modules__form" @submit.prevent="saveModules">
        <fieldset class="org-modules__fieldset">
          <legend>{{ $t('org.modules_title') }}</legend>
          <p class="org-form__hint">{{ $t('org.modules_hint') }}</p>
          <label class="org-modules__check">
            <input v-model="invoicingEnabled" type="checkbox">
            {{ $t('org.invoicing_enabled') }}
          </label>
          <p class="org-form__hint">{{ $t('org.invoicing_enabled_hint') }}</p>
        </fieldset>
        <AppButton variant="primary" type="submit" :disabled="savingModules">
          {{ $t('org.save') }}
        </AppButton>
        <p
          v-if="modulesMessage"
          class="org-form__msg"
          :class="{ 'org-form__msg--error': modulesError }"
          role="status"
        >
          {{ modulesMessage }}
        </p>
      </form>
    </AppCard>

    <div v-else class="split-layout">
      <AppCard padding="lg" class="org-form">
        <form class="org-form__fields" @submit.prevent="save">
          <AppInput id="raison" v-model="form.raisonSociale" :label="$t('org.company_name')" />
          <AppInput id="adresse" v-model="form.adresse" :label="$t('org.address_street')" />
          <div class="org-form__row">
            <AppInput id="adresse-numero" v-model="form.adresseNumero" :label="$t('org.address_number')" />
            <AppInput id="adresse-boite" v-model="form.adresseBoite" :label="addressBoxLabel" />
          </div>
          <div class="org-form__row">
            <AppInput id="code-postal" v-model="form.codePostal" :label="$t('org.postal_code')" />
            <AppInput id="ville" v-model="form.ville" :label="$t('org.city')" />
          </div>
          <div class="org-form__field">
            <label for="pays" class="org-form__label">{{ $t('org.country') }}</label>
            <select id="pays" v-model="form.pays" class="org-form__select">
              <option value="FR">{{ $t('org.country_fr') }}</option>
              <option value="BE">{{ $t('org.country_be') }}</option>
              <option value="MG">{{ $t('org.country_mg') }}</option>
              <option value="MA">{{ $t('org.country_ma') }}</option>
              <option value="TN">{{ $t('org.country_tn') }}</option>
              <option value="CA">{{ $t('org.country_ca') }}</option>
            </select>
          </div>
          <div>
            <AppInput
              id="siret"
              v-model="form.siret"
              :label="registryLabel"
              :placeholder="registryPlaceholder"
            />
            <p class="org-form__hint">{{ registryHint }}</p>
          </div>
          <div class="org-form__logo">
            <label for="logo-upload">{{ $t('org.logo') }}</label>
            <p class="org-form__hint">{{ $t('org.logo_hint') }}</p>
            <label class="org-form__upload" for="logo-upload">
              <AppIcon name="upload" />
              <span>{{ form.logoFile?.name || $t('org.choose_file') }}</span>
              <input id="logo-upload" type="file" accept="image/png,image/svg+xml,image/jpeg,image/webp" hidden @change="onFileChange" />
            </label>
          </div>
          <AppButton variant="primary" type="submit" :disabled="saving">{{ $t('org.save') }}</AppButton>
        </form>
        <p v-if="message" class="org-form__msg" :class="{ 'org-form__msg--error': isError }" role="status">{{ message }}</p>
      </AppCard>

      <AppCard padding="lg" class="org-preview">
        <h3>{{ $t('org.preview_title') }}</h3>
        <p class="org-preview__hint">{{ $t('org.preview_hint') }}</p>
        <div class="org-preview__frame">
          <TenantLogo :logo-url="previewUrl" :alt="form.raisonSociale || 'Société'" size="lg" />
          <p class="org-preview__name">{{ form.raisonSociale || '—' }}</p>
        </div>
      </AppCard>
    </div>
  </div>
</template>

<script setup lang="ts">
definePageMeta({ layout: 'default', middleware: 'admin' })

const { apiFetch } = useApiFetch()
const { t } = useI18n()
const { extractFetchError } = useApiError()
const { branding, fetchBranding } = useTenantBranding()
const { settings, fetchSettings, saveSettings } = useRequestSettings()

type CountryCode = 'FR' | 'BE' | 'MG' | 'MA' | 'TN' | 'CA'

const route = useRoute()
const tab = ref<'identite' | 'structure' | 'modules'>(
  route.query.tab === 'structure'
    ? 'structure'
    : route.query.tab === 'modules'
      ? 'modules'
      : 'identite'
)

const form = reactive({
  raisonSociale: '',
  adresse: '',
  adresseNumero: '',
  adresseBoite: '',
  codePostal: '',
  ville: '',
  pays: 'FR' as CountryCode,
  siret: '',
  logoFile: null as File | null
})
const previewUrl = ref<string | null>(null)
const saving = ref(false)
const message = ref('')
const isError = ref(false)

const invoicingEnabled = ref(false)
const savingModules = ref(false)
const modulesMessage = ref('')
const modulesError = ref(false)

const normalizeCountry = (value: unknown): CountryCode => {
  const code = typeof value === 'string' ? value.trim().toUpperCase() : ''
  switch (code) {
    case 'BE':
      return 'BE'
    case 'FR':
      return 'FR'
    case 'MG':
    case 'MD': // legacy typo (Moldova ISO); treat as Madagascar
      return 'MG'
    case 'MA':
      return 'MA'
    case 'TN':
      return 'TN'
    case 'CA':
      return 'CA'
    default:
      return 'FR'
  }
}

const registryLabel = computed(() => {
  switch (form.pays) {
    case 'BE':
      return t('org.bce')
    case 'FR':
      return t('org.siret')
    case 'MG':
      return t('org.nif_stat')
    case 'MA':
      return t('org.ice')
    case 'TN':
      return t('org.matricule_fiscal')
    case 'CA':
      return t('org.ne')
    default: {
      const _exhaustive: never = form.pays
      return _exhaustive
    }
  }
})

const registryPlaceholder = computed(() => {
  switch (form.pays) {
    case 'BE':
      return '0123456789'
    case 'FR':
      return '12345678901234'
    case 'MG':
      return '1234567890'
    case 'MA':
      return '123456789012345'
    case 'TN':
      return '1234567/A/A/A/000'
    case 'CA':
      return '123456789'
    default: {
      const _exhaustive: never = form.pays
      return _exhaustive
    }
  }
})

const registryHint = computed(() => {
  switch (form.pays) {
    case 'BE':
      return t('org.registry_hint_be')
    case 'FR':
      return t('org.registry_hint_fr')
    case 'MG':
      return t('org.registry_hint_mg')
    case 'MA':
      return t('org.registry_hint_ma')
    case 'TN':
      return t('org.registry_hint_tn')
    case 'CA':
      return t('org.registry_hint_ca')
    default: {
      const _exhaustive: never = form.pays
      return _exhaustive
    }
  }
})

const addressBoxLabel = computed(() => {
  switch (form.pays) {
    case 'BE':
    case 'FR':
      return t('org.address_box')
    case 'CA':
      return t('org.address_box_ca')
    case 'MG':
    case 'MA':
    case 'TN':
      return t('org.address_box_other')
    default: {
      const _exhaustive: never = form.pays
      return _exhaustive
    }
  }
})

onMounted(async () => {
  await Promise.all([fetchBranding(), fetchSettings()])
  form.raisonSociale = branding.value.raisonSociale
  previewUrl.value = branding.value.logoUrl
  invoicingEnabled.value = settings.value?.invoicingEnabled ?? false
  try {
    const res = await apiFetch<{
      data?: Array<{
        adresse?: string
        adresseNumero?: string
        adresseBoite?: string
        codePostal?: string
        ville?: string
        pays?: string
        siret?: string
      }>
    }>('/api/org/societes')
    const first = res?.data?.[0]
    if (first) {
      form.adresse = first.adresse ?? ''
      form.adresseNumero = first.adresseNumero ?? ''
      form.adresseBoite = first.adresseBoite ?? ''
      form.codePostal = first.codePostal ?? ''
      form.ville = first.ville ?? ''
      form.pays = normalizeCountry(first.pays)
      form.siret = first.siret ?? ''
    }
  } catch {
    // ignore
  }
})

const onFileChange = (e: Event) => {
  const input = e.target as HTMLInputElement
  const file = input.files?.[0]
  if (!file) return
  form.logoFile = file
  previewUrl.value = URL.createObjectURL(file)
}

const save = async () => {
  if (!branding.value.societeId) {
    message.value = t('org.no_company')
    isError.value = true
    return
  }
  saving.value = true
  message.value = ''
  isError.value = false
  try {
    const body = new FormData()
    body.append('raisonSociale', form.raisonSociale)
    body.append('adresse', form.adresse)
    body.append('adresseNumero', form.adresseNumero)
    body.append('adresseBoite', form.adresseBoite)
    body.append('codePostal', form.codePostal)
    body.append('ville', form.ville)
    body.append('pays', form.pays)
    body.append('siret', form.siret)
    if (form.logoFile) body.append('logo', form.logoFile)
    await apiFetch(`/api/org/societes/${branding.value.societeId}/branding`, { method: 'PUT', body })
    await fetchBranding()
    previewUrl.value = branding.value.logoUrl
    message.value = t('org.saved')
  } catch {
    message.value = t('org.error')
    isError.value = true
  } finally {
    saving.value = false
  }
}

const saveModules = async () => {
  savingModules.value = true
  modulesMessage.value = ''
  modulesError.value = false
  try {
    const channels = settings.value?.channelsEnabled ?? {
      tma: true,
      support: false,
      maintenance: false
    }
    await saveSettings({
      channelsEnabled: { ...channels },
      guidesEnabled: settings.value?.guidesEnabled ?? true,
      invoicingEnabled: invoicingEnabled.value
    })
    modulesMessage.value = t('org.modules_saved')
  } catch (e) {
    modulesError.value = true
    modulesMessage.value = extractFetchError(e, t('org.modules_save_error'))
  } finally {
    savingModules.value = false
  }
}
</script>

<style scoped>
.org-tabs {
  display: flex;
  gap: var(--kore-space-xs);
  margin-bottom: var(--kore-space-lg);
  border-bottom: 1px solid var(--kore-border);
}

.org-tab {
  padding: var(--kore-space-sm) var(--kore-space-md);
  border: 0;
  border-bottom: 2px solid transparent;
  background: none;
  color: var(--kore-text-muted);
  font-family: var(--kore-font);
  font-size: var(--kore-text-small);
  font-weight: 500;
  cursor: pointer;
}

.org-tab:hover { color: var(--kore-text); }

.org-tab--active {
  color: var(--kore-text);
  border-bottom-color: var(--kore-brand-gold);
}

@media (max-width: 768px) {
  .org-tabs { overflow-x: auto; }
  .org-tab { white-space: nowrap; }
}

.org-form__fields {
  display: flex;
  flex-direction: column;
  gap: var(--kore-space-lg);
  max-width: var(--kore-form-wide-max);
}

.org-form__row {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: var(--kore-space-md);
}

.org-form__field {
  display: flex;
  flex-direction: column;
  gap: var(--kore-space-xs);
}

.org-form__label {
  font-size: var(--kore-text-small);
  color: var(--kore-text-muted);
  font-weight: 500;
}

.org-form__select {
  padding: 0.75rem 1rem;
  font-family: var(--kore-font);
  font-size: var(--kore-text-body);
  color: var(--kore-text);
  background: var(--kore-bg-elevated);
  border: 1px solid var(--kore-border);
  border-radius: var(--kore-radius-md);
}

.org-form__hint {
  margin: var(--kore-space-xs) 0 0;
  font-size: var(--kore-text-caption);
  color: var(--kore-text-muted);
}

@media (max-width: 768px) {
  .org-form__row { grid-template-columns: 1fr; }
  .org-form__fields :deep(.app-button) { width: 100%; }
}

.org-form__upload {
  display: flex;
  align-items: center;
  gap: var(--kore-space-sm);
  padding: var(--kore-space-md);
  border: 1px dashed var(--kore-border);
  border-radius: var(--kore-radius-md);
  cursor: pointer;
  font-size: var(--kore-text-small);
  color: var(--kore-text-muted);
  transition: border-color 0.15s, color 0.15s;
}

.org-form__upload:hover {
  border-color: var(--kore-brand-gold);
  color: var(--kore-brand-gold);
}

.org-form__msg {
  margin: var(--kore-space-md) 0 0;
  font-size: var(--kore-text-small);
  color: var(--kore-success);
}

.org-form__msg--error { color: var(--kore-error); }

.org-modules__form {
  display: flex;
  flex-direction: column;
  gap: var(--kore-space-md);
  max-width: var(--kore-form-max);
}

.org-modules__fieldset {
  border: 1px solid var(--kore-border);
  border-radius: var(--kore-radius-md);
  padding: var(--kore-space-md);
  display: flex;
  flex-direction: column;
  gap: var(--kore-space-sm);
}

.org-modules__check {
  display: flex;
  align-items: center;
  gap: var(--kore-space-sm);
  font-size: var(--kore-text-small);
  color: var(--kore-text);
  cursor: pointer;
}

.org-preview h3 {
  margin: 0 0 var(--kore-space-xs);
  font-size: var(--kore-text-h3);
}

.org-preview__hint {
  margin: 0 0 var(--kore-space-lg);
  font-size: var(--kore-text-caption);
  color: var(--kore-text-muted);
}

.org-preview__frame {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: var(--kore-space-md);
  padding: var(--kore-space-xl);
  border-radius: var(--kore-radius-md);
  background: var(--kore-bg-subtle);
  border: 1px solid var(--kore-border);
}

.org-preview__name {
  margin: 0;
  font-weight: 600;
  font-size: var(--kore-text-small);
  color: var(--kore-text);
}
</style>
