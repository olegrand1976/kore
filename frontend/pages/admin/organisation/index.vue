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
        <form @submit.prevent="save">
          <AppInput id="raison" v-model="form.raisonSociale" :label="$t('org.company_name')" />
          <AppInput id="adresse" v-model="form.adresse" :label="$t('org.address')" />
          <AppInput id="siret" v-model="form.siret" :label="$t('org.siret')" />
          <AppInput id="url" v-model="form.urlTenant" :label="$t('org.url')" />
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
  siret: '',
  urlTenant: '',
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

onMounted(async () => {
  await Promise.all([fetchBranding(), fetchSettings()])
  form.raisonSociale = branding.value.raisonSociale
  previewUrl.value = branding.value.logoUrl
  invoicingEnabled.value = settings.value?.invoicingEnabled ?? false
  try {
    const res = await apiFetch<any>('/api/org/societes')
    const first = res?.data?.[0]
    if (first) {
      form.adresse = first.adresse ?? ''
      form.siret = first.siret ?? ''
      form.urlTenant = first.urlTenant ?? ''
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
    body.append('siret', form.siret)
    body.append('urlTenant', form.urlTenant)
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

.org-form form { display: flex; flex-direction: column; gap: var(--kore-space-lg); }

.org-form__hint {
  margin: 0 0 var(--kore-space-sm);
  font-size: var(--kore-text-caption);
  color: var(--kore-text-muted);
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
