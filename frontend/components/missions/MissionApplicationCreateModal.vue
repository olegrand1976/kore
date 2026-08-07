<script setup lang="ts">
const props = defineProps<{
  open: boolean
  defaultLibelle?: string
  sites: { id: string; label: string }[]
}>()

const emit = defineEmits<{
  'update:open': [value: boolean]
  created: [applicationId: string]
}>()

const { t } = useI18n()
const { createApplication } = useOrganisation()
const { extractFetchError } = useApiError()

const creating = ref(false)
const error = ref('')
const form = reactive({
  libelle: '',
  siteId: ''
})

watch(
  () => props.open,
  (open) => {
    if (!open) return
    form.libelle = (props.defaultLibelle ?? '').trim()
    form.siteId = props.sites[0]?.id ?? ''
    error.value = ''
  }
)

const close = () => emit('update:open', false)

const submit = async () => {
  if (!form.libelle.trim() || !form.siteId) {
    error.value = t('missions.app_create_required')
    return
  }
  creating.value = true
  error.value = ''
  try {
    const res = await createApplication({
      libelle: form.libelle.trim(),
      siteIds: [form.siteId]
    })
    const created = (res as { data?: { id?: string; ID?: string } })?.data
    const createdId = created?.id ?? created?.ID ?? ''
    if (createdId) emit('created', createdId)
    close()
  } catch (err) {
    error.value = extractFetchError(err, t('missions.app_create_error'))
  } finally {
    creating.value = false
  }
}
</script>

<template>
  <AppModal
    :open="open"
    width="sm"
    :aria-label="$t('missions.app_add_inline')"
    @update:open="emit('update:open', $event)"
  >
    <form class="mission-app-create" @submit.prevent="submit">
      <h2 class="mission-app-create__title">{{ $t('missions.app_add_inline') }}</h2>
      <AppInput
        id="mission-app-create-libelle"
        v-model="form.libelle"
        :label="$t('missions.app_libelle')"
        required
      />
      <div class="mission-app-create__field">
        <label for="mission-app-create-site">{{ $t('missions.app_site') }}</label>
        <select id="mission-app-create-site" v-model="form.siteId" class="mission-app-create__select" required>
          <option value="">{{ $t('missions.app_site_placeholder') }}</option>
          <option v-for="s in sites" :key="s.id" :value="s.id">{{ s.label }}</option>
        </select>
      </div>
      <p class="mission-app-create__hint">{{ $t('missions.app_create_hint') }}</p>
      <p v-if="error" class="flash flash--error" role="alert">{{ error }}</p>
      <div class="mission-app-create__actions">
        <AppButton variant="ghost" type="button" @click="close">
          {{ $t('common.cancel') }}
        </AppButton>
        <AppButton variant="primary" type="submit" :loading="creating">
          {{ $t('common.save') }}
        </AppButton>
      </div>
    </form>
  </AppModal>
</template>

<style scoped>
.mission-app-create {
  display: grid;
  gap: var(--kore-space-md);
}

.mission-app-create__title {
  margin: 0;
  font-size: var(--kore-text-h3);
}

.mission-app-create__field {
  display: flex;
  flex-direction: column;
  gap: var(--kore-space-xs);
}

.mission-app-create__field label {
  font-size: var(--kore-text-small);
  font-weight: 500;
  color: var(--kore-text-muted);
}

.mission-app-create__select {
  padding: 0.625rem 0.75rem;
  border: 1px solid var(--kore-border);
  border-radius: var(--kore-radius-md);
  background: var(--kore-bg);
  color: var(--kore-text);
  font-size: var(--kore-text-small);
}

.mission-app-create__hint {
  margin: 0;
  font-size: var(--kore-text-caption);
  color: var(--kore-text-muted);
}

.mission-app-create__actions {
  display: flex;
  flex-wrap: wrap;
  gap: var(--kore-space-sm);
  justify-content: flex-end;
}

.flash--error {
  margin: 0;
  color: var(--kore-error);
  font-size: var(--kore-text-small);
}

@media (max-width: 768px) {
  .mission-app-create__actions {
    flex-direction: column-reverse;
  }

  .mission-app-create__actions :deep(.app-button) {
    width: 100%;
  }
}
</style>
