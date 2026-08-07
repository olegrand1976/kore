<template>
  <div>
    <AppPageHeader :title="$t('services_admin.title')" :subtitle="$t('services_admin.subtitle')">
      <template #actions>
        <AppButton
          variant="primary"
          size="sm"
          type="button"
          :disabled="!canCreate"
          @click="openCreate"
        >
          <AppIcon name="add" /> {{ $t('services_admin.create') }}
        </AppButton>
      </template>
    </AppPageHeader>

    <p
      v-if="flash"
      class="org-admin-flash"
      :class="{ 'org-admin-flash--error': flashError }"
      role="status"
    >
      {{ flash }}
    </p>

    <p v-if="!pending && !forbidden && createBlockedHint" class="org-admin-hint" role="note">
      {{ createBlockedHint }}
    </p>

    <AppCard v-if="pending" padding="lg">
      <p class="muted">{{ $t('services_admin.loading') }}</p>
    </AppCard>

    <AppCard v-else-if="forbidden" padding="lg">
      <AppEmptyState icon="lock" :title="$t('services_admin.forbidden')" />
    </AppCard>

    <AppCard v-else padding="lg">
      <AppTable
        :columns="columns"
        :rows="rows"
        :empty-title="$t('services_admin.empty')"
        row-key="id"
      >
        <template #cell-type="{ value }">
          <AppBadge variant="default">{{ typeLabel(String(value)) }}</AppBadge>
        </template>
        <template #cell-actions="{ row }">
          <AppButton variant="ghost" size="sm" type="button" @click="openEdit(row)">
            {{ $t('common.edit') }}
          </AppButton>
        </template>
      </AppTable>
    </AppCard>

    <AppModal
      v-model:open="showForm"
      width="sm"
      :aria-label="editingId ? $t('services_admin.edit_title') : $t('services_admin.create_title')"
    >
      <form class="org-admin-form" @submit.prevent="save">
        <h2 class="org-admin-form__title">
          {{ editingId ? $t('services_admin.edit_title') : $t('services_admin.create_title') }}
        </h2>

        <p v-if="editingId" class="org-admin-form__hint">{{ $t('services_admin.edit_libelle_only') }}</p>

        <div v-if="!editingId" class="org-admin-form__field">
          <label for="service-site">{{ $t('services_admin.field_site') }}</label>
          <select id="service-site" v-model="form.siteId" required>
            <option value="" disabled>{{ $t('services_admin.site_placeholder') }}</option>
            <option v-for="s in siteOptions" :key="s.value" :value="s.value">
              {{ s.label }}
            </option>
          </select>
        </div>

        <AppInput
          id="service-libelle"
          v-model="form.libelle"
          :label="$t('services_admin.field_libelle')"
          required
        />

        <div v-if="!editingId" class="org-admin-form__field">
          <label for="service-type">{{ $t('services_admin.field_type') }}</label>
          <select id="service-type" v-model="form.type">
            <option value="interne">{{ $t('services_admin.type_interne') }}</option>
            <option value="externe">{{ $t('services_admin.type_externe') }}</option>
          </select>
        </div>

        <div v-if="!editingId" class="org-admin-form__field">
          <label for="service-responsable">{{ $t('services_admin.field_responsable') }}</label>
          <select id="service-responsable" v-model="form.responsableId" required>
            <option value="" disabled>{{ $t('services_admin.responsable_placeholder') }}</option>
            <option v-for="u in userOptions" :key="u.value" :value="u.value">
              {{ u.label }}
            </option>
          </select>
          <p class="org-admin-form__hint">{{ $t('services_admin.responsable_hint') }}</p>
        </div>

        <p v-if="formError" class="org-admin-form__error" role="alert">{{ formError }}</p>

        <div class="org-admin-form__actions">
          <AppButton variant="ghost" type="button" @click="showForm = false">
            {{ $t('common.cancel') }}
          </AppButton>
          <AppButton variant="primary" type="submit" :disabled="saving">
            {{ $t('common.save') }}
          </AppButton>
        </div>
      </form>
    </AppModal>
  </div>
</template>

<script setup lang="ts">
definePageMeta({ layout: 'default', middleware: 'admin' })

const { t } = useI18n()
const { extractFetchError } = useApiError()
const {
  listSites,
  listServices,
  createService,
  updateService,
  orgId,
  orgLabel
} = useOrganisation()
const { list: listUsers, pickUserId, pickUserLogin } = useUsers()

type ServiceRow = {
  id: string
  libelle: string
  type: string
  siteId: string
  siteLabel: string
  responsableId: string
  responsableLabel: string
}

const pending = ref(true)
const forbidden = ref(false)
const flash = ref('')
const flashError = ref(false)
const rows = ref<ServiceRow[]>([])
const siteOptions = ref<{ value: string; label: string }[]>([])
const userOptions = ref<{ value: string; label: string }[]>([])

const showForm = ref(false)
const editingId = ref('')
const saving = ref(false)
const formError = ref('')
const form = reactive({
  siteId: '',
  libelle: '',
  type: 'interne',
  responsableId: ''
})

const columns = computed(() => [
  { key: 'libelle', label: t('services_admin.col_libelle') },
  { key: 'type', label: t('services_admin.col_type') },
  { key: 'siteLabel', label: t('services_admin.col_site') },
  { key: 'responsableLabel', label: t('services_admin.col_responsable') },
  { key: 'actions', label: t('common.actions') }
])

const canCreate = computed(
  () => siteOptions.value.length > 0 && userOptions.value.length > 0
)

const createBlockedHint = computed(() => {
  if (canCreate.value) return ''
  if (!siteOptions.value.length) return t('services_admin.create_need_site')
  if (!userOptions.value.length) return t('services_admin.create_need_user')
  return ''
})

const typeLabel = (value: string) => {
  switch (value) {
    case 'externe':
      return t('services_admin.type_externe')
    case 'interne':
      return t('services_admin.type_interne')
    default:
      return value || t('services_admin.type_interne')
  }
}

const load = async () => {
  pending.value = true
  forbidden.value = false
  try {
    const [sites, services, users] = await Promise.all([
      listSites(),
      listServices(),
      listUsers()
    ])
    const siteMap = new Map<string, string>()
    siteOptions.value = sites.map((s) => {
      const id = orgId(s)
      const label = orgLabel(s) || id
      siteMap.set(id, label)
      return { value: id, label }
    })
    const userMap = new Map<string, string>()
    userOptions.value = users.map((u) => {
      const id = pickUserId(u)
      const label = pickUserLogin(u) || id
      userMap.set(id, label)
      return { value: id, label }
    })

    rows.value = services.map((s) => {
      const siteId = s.siteId ?? s.SiteID ?? ''
      const responsableId = s.responsableId ?? s.ResponsableID ?? ''
      return {
        id: orgId(s),
        libelle: orgLabel(s) || s.type || orgId(s),
        type: s.type || 'interne',
        siteId,
        siteLabel: s.siteLabel || siteMap.get(siteId) || siteId,
        responsableId,
        responsableLabel: userMap.get(responsableId) || (responsableId || t('common.none'))
      }
    })
  } catch (err) {
    if ((err as { statusCode?: number })?.statusCode === 403) {
      forbidden.value = true
      rows.value = []
    } else {
      flash.value = extractFetchError(err, t('services_admin.load_error'))
      flashError.value = true
    }
  } finally {
    pending.value = false
  }
}

onMounted(load)

const openCreate = () => {
  if (!canCreate.value) return
  editingId.value = ''
  form.siteId = siteOptions.value[0]?.value ?? ''
  form.libelle = ''
  form.type = 'interne'
  form.responsableId = userOptions.value[0]?.value ?? ''
  formError.value = ''
  flash.value = ''
  showForm.value = true
}

const openEdit = (row: ServiceRow) => {
  editingId.value = row.id
  form.siteId = row.siteId
  form.libelle = row.libelle
  form.type = row.type
  form.responsableId = row.responsableId
  formError.value = ''
  flash.value = ''
  showForm.value = true
}

const save = async () => {
  const libelle = form.libelle.trim()
  if (!libelle) {
    formError.value = t('services_admin.field_libelle_required')
    return
  }
  if (!editingId.value) {
    if (!form.siteId) {
      formError.value = t('services_admin.field_site_required')
      return
    }
    if (!form.responsableId) {
      formError.value = t('services_admin.field_responsable_required')
      return
    }
  }
  saving.value = true
  formError.value = ''
  try {
    if (editingId.value) {
      await updateService(editingId.value, { libelle })
      flash.value = t('services_admin.updated')
    } else {
      await createService({
        siteId: form.siteId,
        libelle,
        type: form.type,
        responsableId: form.responsableId
      })
      flash.value = t('services_admin.created')
    }
    flashError.value = false
    showForm.value = false
    await load()
  } catch (err) {
    formError.value = extractFetchError(err, t('services_admin.save_error'))
  } finally {
    saving.value = false
  }
}
</script>

<style scoped>
.org-admin-flash {
  margin: 0 0 var(--kore-space-md);
  font-size: var(--kore-text-small);
  color: var(--kore-success);
}
.org-admin-flash--error {
  color: var(--kore-error);
}
.org-admin-hint {
  margin: 0 0 var(--kore-space-md);
  font-size: var(--kore-text-small);
  color: var(--kore-text-muted);
}
.org-admin-form {
  display: flex;
  flex-direction: column;
  gap: var(--kore-space-lg);
}
.org-admin-form__title {
  margin: 0;
  font-size: var(--kore-text-h3);
}
.org-admin-form__field {
  display: flex;
  flex-direction: column;
  gap: var(--kore-space-xs);
}
.org-admin-form__field label {
  font-size: var(--kore-text-small);
  font-weight: 500;
  color: var(--kore-text-muted);
}
.org-admin-form__field select {
  padding: 0.75rem 1rem;
  font-family: var(--kore-font);
  font-size: var(--kore-text-body);
  color: var(--kore-text);
  background: var(--kore-bg-elevated);
  border: 1px solid var(--kore-border);
  border-radius: var(--kore-radius-md);
}
.org-admin-form__hint {
  margin: 0;
  font-size: var(--kore-text-caption);
  color: var(--kore-text-muted);
}
.org-admin-form__error {
  margin: 0;
  font-size: var(--kore-text-small);
  color: var(--kore-error);
}
.org-admin-form__actions {
  display: flex;
  justify-content: flex-end;
  gap: var(--kore-space-sm);
}
@media (max-width: 768px) {
  .org-admin-form__actions {
    flex-direction: column-reverse;
  }
  .org-admin-form__actions :deep(.app-button) {
    width: 100%;
  }
}
</style>
