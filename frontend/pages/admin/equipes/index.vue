<template>
  <div>
    <AppPageHeader :title="$t('equipes_admin.title')" :subtitle="$t('equipes_admin.subtitle')">
      <template #actions>
        <AppButton
          variant="primary"
          size="sm"
          type="button"
          :disabled="!canCreate"
          @click="openCreate"
        >
          <AppIcon name="add" /> {{ $t('equipes_admin.create') }}
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
      {{ $t('equipes_admin.create_need_application') }}
    </p>

    <AppCard v-if="pending" padding="lg">
      <p class="muted">{{ $t('equipes_admin.loading') }}</p>
    </AppCard>

    <AppCard v-else-if="forbidden" padding="lg">
      <AppEmptyState icon="lock" :title="$t('equipes_admin.forbidden')" />
    </AppCard>

    <AppCard v-else padding="lg">
      <AppTable
        :columns="columns"
        :rows="rows"
        :empty-title="$t('equipes_admin.empty')"
        row-key="id"
      >
        <template #cell-memberCount="{ value }">
          <AppBadge :variant="Number(value) > 0 ? 'success' : 'default'">
            {{ $t('equipes_admin.members', { count: value }) }}
          </AppBadge>
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
      :aria-label="editingId ? $t('equipes_admin.edit_title') : $t('equipes_admin.create_title')"
    >
      <form class="org-admin-form" @submit.prevent="save">
        <h2 class="org-admin-form__title">
          {{ editingId ? $t('equipes_admin.edit_title') : $t('equipes_admin.create_title') }}
        </h2>

        <div v-if="!editingId" class="org-admin-form__field">
          <label for="equipe-application">{{ $t('equipes_admin.field_application') }}</label>
          <select id="equipe-application" v-model="form.applicationId" required>
            <option value="" disabled>{{ $t('equipes_admin.application_placeholder') }}</option>
            <option v-for="a in applicationOptions" :key="a.value" :value="a.value">
              {{ a.label }}
            </option>
          </select>
        </div>

        <AppInput
          id="equipe-libelle"
          v-model="form.libelle"
          :label="$t('equipes_admin.field_libelle')"
          required
        />

        <div class="org-admin-form__field">
          <label for="equipe-responsable">{{ $t('equipes_admin.field_responsable') }}</label>
          <select id="equipe-responsable" v-model="form.responsableId">
            <option value="">{{ $t('equipes_admin.responsable_none') }}</option>
            <option v-for="u in userOptions" :key="u.value" :value="u.value">
              {{ u.label }}
            </option>
          </select>
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
  listApplications,
  listEquipes,
  createEquipe,
  updateEquipe,
  orgId,
  orgLabel
} = useOrganisation()
const { list: listUsers, pickUserId, pickUserLogin, pickUserEquipeIds } = useUsers()

type EquipeRow = {
  id: string
  libelle: string
  applicationId: string
  applicationLabel: string
  responsableId: string
  responsableLabel: string
  memberCount: number
}

const pending = ref(true)
const forbidden = ref(false)
const flash = ref('')
const flashError = ref(false)
const rows = ref<EquipeRow[]>([])
const applicationOptions = ref<{ value: string; label: string }[]>([])
const userOptions = ref<{ value: string; label: string }[]>([])

const showForm = ref(false)
const editingId = ref('')
const saving = ref(false)
const formError = ref('')
const form = reactive({
  applicationId: '',
  libelle: '',
  responsableId: ''
})

const columns = computed(() => [
  { key: 'libelle', label: t('equipes_admin.col_libelle') },
  { key: 'applicationLabel', label: t('equipes_admin.col_application') },
  { key: 'responsableLabel', label: t('equipes_admin.col_responsable') },
  { key: 'memberCount', label: t('equipes_admin.col_members') },
  { key: 'actions', label: t('common.actions') }
])

const canCreate = computed(() => applicationOptions.value.length > 0)

const load = async () => {
  pending.value = true
  forbidden.value = false
  try {
    const [apps, equipes, users] = await Promise.all([
      listApplications(),
      listEquipes(),
      listUsers()
    ])
    const appMap = new Map<string, string>()
    applicationOptions.value = apps
      .filter((a) => a.active ?? a.Active ?? true)
      .map((a) => {
        const id = orgId(a)
        const label = orgLabel(a) || id
        appMap.set(id, label)
        return { value: id, label }
      })
    // Keep inactive app labels for display of existing equipes.
    for (const a of apps) {
      const id = orgId(a)
      if (!appMap.has(id)) appMap.set(id, orgLabel(a) || id)
    }

    const userMap = new Map<string, string>()
    userOptions.value = users.map((u) => {
      const id = pickUserId(u)
      const label = pickUserLogin(u) || id
      userMap.set(id, label)
      return { value: id, label }
    })

    rows.value = equipes.map((e) => {
      const applicationId = e.applicationId ?? e.ApplicationID ?? ''
      const responsableId = e.responsableId ?? e.ResponsableID ?? ''
      const id = orgId(e)
      return {
        id,
        libelle: orgLabel(e) || id,
        applicationId,
        applicationLabel: appMap.get(applicationId) || applicationId,
        responsableId,
        responsableLabel: userMap.get(responsableId) || (responsableId ? responsableId : t('common.none')),
        memberCount: users.filter((u) => pickUserEquipeIds(u).includes(id)).length
      }
    })
  } catch (err) {
    if ((err as { statusCode?: number })?.statusCode === 403) {
      forbidden.value = true
      rows.value = []
    } else {
      flash.value = extractFetchError(err, t('equipes_admin.load_error'))
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
  form.applicationId = applicationOptions.value[0]?.value ?? ''
  form.libelle = ''
  form.responsableId = ''
  formError.value = ''
  flash.value = ''
  showForm.value = true
}

const openEdit = (row: EquipeRow) => {
  editingId.value = row.id
  form.applicationId = row.applicationId
  form.libelle = row.libelle
  form.responsableId = row.responsableId
  formError.value = ''
  flash.value = ''
  showForm.value = true
}

const save = async () => {
  const libelle = form.libelle.trim()
  if (!libelle) {
    formError.value = t('equipes_admin.field_libelle_required')
    return
  }
  if (!editingId.value && !form.applicationId) {
    formError.value = t('equipes_admin.field_application_required')
    return
  }
  saving.value = true
  formError.value = ''
  try {
    if (editingId.value) {
      await updateEquipe(editingId.value, {
        libelle,
        responsableId: form.responsableId || null
      })
      flash.value = t('equipes_admin.updated')
    } else {
      await createEquipe({
        applicationId: form.applicationId,
        libelle,
        responsableId: form.responsableId || undefined
      })
      flash.value = t('equipes_admin.created')
    }
    flashError.value = false
    showForm.value = false
    await load()
  } catch (err) {
    formError.value = extractFetchError(err, t('equipes_admin.save_error'))
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
