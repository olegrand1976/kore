<template>
  <div>
    <AppPageHeader :title="$t('org.title')" :subtitle="$t('org.subtitle')" />

    <div class="split-layout">
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

const { t } = useI18n()
const { branding, fetchBranding } = useTenantBranding()

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

onMounted(async () => {
  await fetchBranding()
  form.raisonSociale = branding.value.raisonSociale
  previewUrl.value = branding.value.logoUrl
  try {
    const res = await $fetch<any>('/api/org/societes')
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
    await $fetch(`/api/org/societes/${branding.value.societeId}/branding`, { method: 'PUT', body })
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
</script>

<style scoped>
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
