<template>
  <div>
    <AppPageHeader :title="$t('sites.title')" :subtitle="$t('sites.subtitle')">
      <template #actions>
        <AppButton
          variant="primary"
          size="sm"
          type="button"
          :disabled="!canCreate"
          @click="openCreate"
        >
          <AppIcon name="add" /> {{ $t('sites.create') }}
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

    <p v-if="!pending && !forbidden && !canCreate" class="org-admin-hint" role="note">
      {{ $t('sites.create_need_societe') }}
    </p>

    <AppCard v-if="pending" padding="lg">
      <p class="muted">{{ $t('sites.loading') }}</p>
    </AppCard>

    <AppCard v-else-if="forbidden" padding="lg">
      <AppEmptyState icon="lock" :title="$t('sites.forbidden')" />
    </AppCard>

    <AppCard v-else padding="lg">
      <AppTable
        :columns="columns"
        :rows="rows"
        :empty-title="$t('sites.empty')"
        row-key="id"
      >
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
      :aria-label="editingId ? $t('sites.edit_title') : $t('sites.create_title')"
    >
      <form class="org-admin-form" @submit.prevent="save">
        <h2 class="org-admin-form__title">
          {{ editingId ? $t('sites.edit_title') : $t('sites.create_title') }}
        </h2>

        <div v-if="!editingId" class="org-admin-form__field">
          <label for="site-societe">{{ $t('sites.field_societe') }}</label>
          <select id="site-societe" v-model="form.societeId" required>
            <option value="" disabled>{{ $t('sites.societe_placeholder') }}</option>
            <option v-for="s in societeOptions" :key="s.value" :value="s.value">
              {{ s.label }}
            </option>
          </select>
        </div>

        <AppInput
          id="site-libelle"
          v-model="form.libelle"
          :label="$t('sites.field_libelle')"
          required
        />

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
  listSocietes,
  listSites,
  createSite,
  updateSite,
  orgId,
  orgLabel
} = useOrganisation()

type SiteRow = {
  id: string
  libelle: string
  societeId: string
  societeLabel: string
}

const pending = ref(true)
const forbidden = ref(false)
const flash = ref('')
const flashError = ref(false)
const rows = ref<SiteRow[]>([])
const societeOptions = ref<{ value: string; label: string }[]>([])

const showForm = ref(false)
const editingId = ref('')
const saving = ref(false)
const formError = ref('')
const form = reactive({ societeId: '', libelle: '' })

const columns = computed(() => [
  { key: 'libelle', label: t('sites.col_libelle') },
  { key: 'societeLabel', label: t('sites.col_societe') },
  { key: 'actions', label: t('common.actions') }
])

const canCreate = computed(() => societeOptions.value.length > 0)

const load = async () => {
  pending.value = true
  forbidden.value = false
  try {
    const [societes, sites] = await Promise.all([listSocietes(), listSites()])
    const map = new Map<string, string>()
    societeOptions.value = societes.map((s) => {
      const id = orgId(s)
      const label = s.raisonSociale ?? s.RaisonSociale ?? id
      map.set(id, label)
      return { value: id, label }
    })
    rows.value = sites.map((s) => {
      const societeId = s.societeId ?? s.SocieteID ?? ''
      return {
        id: orgId(s),
        libelle: orgLabel(s) || orgId(s),
        societeId,
        societeLabel: map.get(societeId) || societeId
      }
    })
  } catch (err) {
    if ((err as { statusCode?: number })?.statusCode === 403) {
      forbidden.value = true
      rows.value = []
    } else {
      flash.value = extractFetchError(err, t('sites.load_error'))
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
  form.societeId = societeOptions.value[0]?.value ?? ''
  form.libelle = ''
  formError.value = ''
  flash.value = ''
  showForm.value = true
}

const openEdit = (row: SiteRow) => {
  editingId.value = row.id
  form.societeId = row.societeId
  form.libelle = row.libelle
  formError.value = ''
  flash.value = ''
  showForm.value = true
}

const save = async () => {
  const libelle = form.libelle.trim()
  if (!libelle) {
    formError.value = t('sites.field_libelle_required')
    return
  }
  if (!editingId.value && !form.societeId) {
    formError.value = t('sites.field_societe_required')
    return
  }
  saving.value = true
  formError.value = ''
  try {
    if (editingId.value) {
      await updateSite(editingId.value, { libelle })
      flash.value = t('sites.updated')
    } else {
      await createSite({ societeId: form.societeId, libelle })
      flash.value = t('sites.created')
    }
    flashError.value = false
    showForm.value = false
    await load()
  } catch (err) {
    formError.value = extractFetchError(err, t('sites.save_error'))
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
