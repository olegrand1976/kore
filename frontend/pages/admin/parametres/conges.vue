<template>
  <div>
    <AppCard padding="lg" class="settings-toolbar">
      <div class="settings-toolbar__row">
        <div class="settings-toolbar__field">
          <label for="societe-select">{{ $t('settings.conges.societe') }}</label>
          <select id="societe-select" v-model="selectedSocieteId" @change="loadConfigs">
            <option v-for="s in societes" :key="s.id" :value="s.id">
              {{ s.label }}
            </option>
          </select>
        </div>
        <p v-if="selectedSociete" class="settings-toolbar__pays">
          {{ $t('settings.conges.country') }} : <strong>{{ selectedSociete.pays }}</strong>
        </p>
      </div>
      <div class="settings-toolbar__actions">
        <AppButton variant="primary" size="sm" @click="openCreate">
          {{ $t('settings.conges.add') }}
        </AppButton>
        <AppButton variant="ghost" size="sm" :disabled="!selectedSocieteId || resetting" @click="showResetConfirm = true">
          {{ $t('settings.conges.reset') }}
        </AppButton>
      </div>
    </AppCard>

    <AppCard padding="lg">
      <AppTable
        :columns="columns"
        :rows="rows"
        :loading="pending"
        :empty-title="$t('settings.conges.empty')"
        row-key="id"
      >
        <template #cell-tracksBalance="{ value }">
          {{ value ? $t('common.yes') : $t('common.no') }}
        </template>
        <template #cell-active="{ value }">
          <AppBadge :variant="value ? 'success' : 'default'">
            {{ value ? $t('settings.conges.active') : $t('settings.conges.inactive') }}
          </AppBadge>
        </template>
        <template #cell-actions="{ row }">
          <div class="settings-actions">
            <AppButton variant="ghost" size="sm" @click="openEdit(row)">{{ $t('common.edit') }}</AppButton>
            <AppButton variant="ghost" size="sm" @click="remove(row)">{{ $t('common.delete') }}</AppButton>
          </div>
        </template>
      </AppTable>
    </AppCard>

    <AppCard v-if="showForm" padding="lg" class="settings-form">
      <h3 class="settings-form__title">
        {{ editingId ? $t('settings.conges.edit_title') : $t('settings.conges.add_title') }}
      </h3>
      <form class="settings-form__grid" @submit.prevent="save">
        <AppInput
          v-if="!editingId"
          id="code"
          v-model="form.code"
          :label="$t('settings.conges.col_code')"
          required
        />
        <AppInput id="label" v-model="form.label" :label="$t('settings.conges.col_label')" required />
        <label class="settings-toggle">
          <input v-model="form.tracksBalance" type="checkbox" />
          {{ $t('settings.conges.col_balance') }}
        </label>
        <label class="settings-toggle">
          <input v-model="form.active" type="checkbox" />
          {{ $t('settings.conges.col_active') }}
        </label>
        <AppInput id="sortOrder" v-model.number="form.sortOrder" type="number" :label="$t('settings.conges.col_order')" />
        <div class="settings-form__actions">
          <AppButton variant="ghost" size="sm" type="button" @click="closeForm">{{ $t('common.cancel') }}</AppButton>
          <AppButton variant="primary" size="sm" type="submit" :disabled="saving">{{ $t('common.save') }}</AppButton>
        </div>
      </form>
      <p v-if="formError" class="settings-flash settings-flash--error" role="alert">{{ formError }}</p>
    </AppCard>

    <div v-if="showResetConfirm" class="settings-overlay" role="dialog" aria-modal="true">
      <AppCard padding="lg" class="settings-dialog">
        <h3 class="settings-form__title">{{ $t('settings.conges.reset_title') }}</h3>
        <p class="settings-dialog__text">
          {{ $t('settings.conges.reset_desc', { country: selectedSociete?.pays ?? '—' }) }}
        </p>
        <div class="settings-form__actions">
          <AppButton variant="ghost" size="sm" type="button" @click="showResetConfirm = false">
            {{ $t('common.cancel') }}
          </AppButton>
          <AppButton variant="primary" size="sm" type="button" :disabled="resetting" @click="confirmReset">
            {{ $t('settings.conges.reset_confirm') }}
          </AppButton>
        </div>
        <p v-if="resetError" class="settings-flash settings-flash--error" role="alert">{{ resetError }}</p>
      </AppCard>
    </div>

    <p v-if="flash" class="settings-flash" :class="{ 'settings-flash--error': flashError }" role="status">{{ flash }}</p>
  </div>
</template>

<script setup lang="ts">
import type { LeaveTypeConfig } from '~/composables/useLeave'
import { pickLeaveTypeCode, pickLeaveTypeLabel, pickSortOrder, pickTracksBalance } from '~/composables/useLeave'

definePageMeta({ layout: 'default', middleware: 'admin' })

const { t } = useI18n()
const { extractFetchError } = useApiError()
const { fetchForSociete } = useLeaveTypeConfigs()

type SocieteRow = { id: string; label: string; pays: string }

const societes = ref<SocieteRow[]>([])
const selectedSocieteId = ref('')
const configs = ref<LeaveTypeConfig[]>([])
const pending = ref(false)
const saving = ref(false)
const resetting = ref(false)
const showForm = ref(false)
const showResetConfirm = ref(false)
const editingId = ref('')
const formError = ref('')
const resetError = ref('')
const flash = ref('')
const flashError = ref(false)

const form = reactive({
  code: '',
  label: '',
  tracksBalance: true,
  active: true,
  sortOrder: 0
})

const selectedSociete = computed(() => societes.value.find((s) => s.id === selectedSocieteId.value))

const columns = computed(() => [
  { key: 'code', label: t('settings.conges.col_code') },
  { key: 'label', label: t('settings.conges.col_label') },
  { key: 'tracksBalance', label: t('settings.conges.col_balance') },
  { key: 'active', label: t('settings.conges.col_active') },
  { key: 'sortOrder', label: t('settings.conges.col_order') },
  { key: 'actions', label: '' }
])

const rows = computed(() =>
  configs.value.map((item) => ({
    id: item.id ?? item.ID ?? pickLeaveTypeCode(item),
    code: pickLeaveTypeCode(item),
    label: pickLeaveTypeLabel(item),
    tracksBalance: pickTracksBalance(item),
    active: item.active ?? item.Active ?? true,
    sortOrder: pickSortOrder(item)
  }))
)

const loadSocietes = async () => {
  const res = await $fetch<{ data?: Array<{ id?: string; ID?: string; raisonSociale?: string; RaisonSociale?: string; pays?: string; Pays?: string }> }>(
    '/api/org/societes'
  )
  societes.value = (res?.data ?? []).map((item) => {
    const id = item.id ?? item.ID ?? ''
    const name = item.raisonSociale ?? item.RaisonSociale ?? '—'
    const pays = item.pays ?? item.Pays ?? 'FR'
    return { id, label: `${name} (${pays})`, pays }
  })
  if (!selectedSocieteId.value && societes.value.length > 0) {
    selectedSocieteId.value = societes.value[0].id
  }
}

const loadConfigs = async () => {
  if (!selectedSocieteId.value) return
  pending.value = true
  try {
    configs.value = await fetchForSociete(selectedSocieteId.value)
  } catch (err) {
    flash.value = extractFetchError(err)
    flashError.value = true
  } finally {
    pending.value = false
  }
}

onMounted(async () => {
  await loadSocietes()
  await loadConfigs()
})

const openCreate = () => {
  editingId.value = ''
  form.code = ''
  form.label = ''
  form.tracksBalance = true
  form.active = true
  form.sortOrder = configs.value.length + 1
  formError.value = ''
  showForm.value = true
}

const openEdit = (row: { id: string; code: string; label: string; tracksBalance: boolean; active: boolean; sortOrder: number }) => {
  editingId.value = row.id
  form.code = row.code
  form.label = row.label
  form.tracksBalance = row.tracksBalance
  form.active = row.active
  form.sortOrder = row.sortOrder
  formError.value = ''
  showForm.value = true
}

const closeForm = () => {
  showForm.value = false
  editingId.value = ''
}

const save = async () => {
  if (!selectedSocieteId.value) return
  saving.value = true
  formError.value = ''
  try {
    if (editingId.value) {
      await $fetch(`/api/conges/leave-type-configs/${editingId.value}`, {
        method: 'PUT',
        body: {
          label: form.label,
          tracksBalance: form.tracksBalance,
          active: form.active,
          sortOrder: form.sortOrder
        }
      })
      flash.value = t('settings.conges.saved')
      flashError.value = false
    } else {
      await $fetch('/api/conges/leave-type-configs', {
        method: 'POST',
        body: {
          societeId: selectedSocieteId.value,
          code: form.code,
          label: form.label,
          tracksBalance: form.tracksBalance,
          active: form.active,
          sortOrder: form.sortOrder
        }
      })
      flash.value = t('settings.conges.created')
      flashError.value = false
    }
    closeForm()
    await loadConfigs()
  } catch (err) {
    formError.value = extractFetchError(err)
  } finally {
    saving.value = false
  }
}

const remove = async (row: { id: string; code: string }) => {
  if (!confirm(t('settings.conges.delete_confirm', { code: row.code }))) return
  try {
    await $fetch(`/api/conges/leave-type-configs/${row.id}`, { method: 'DELETE' })
    flash.value = t('settings.conges.deleted')
    flashError.value = false
    await loadConfigs()
  } catch (err) {
    flash.value = extractFetchError(err)
    flashError.value = true
  }
}

const confirmReset = async () => {
  if (!selectedSocieteId.value) return
  resetting.value = true
  resetError.value = ''
  try {
    await $fetch('/api/conges/leave-type-configs/reset', {
      method: 'POST',
      body: { societeId: selectedSocieteId.value, confirm: true }
    })
    showResetConfirm.value = false
    flash.value = t('settings.conges.reset_done')
    flashError.value = false
    await loadConfigs()
  } catch (err) {
    resetError.value = extractFetchError(err)
  } finally {
    resetting.value = false
  }
}
</script>

<style scoped>
.settings-toolbar {
  margin-bottom: var(--kore-space-lg);
}

.settings-toolbar__row {
  display: flex;
  flex-wrap: wrap;
  gap: var(--kore-space-md);
  align-items: flex-end;
  margin-bottom: var(--kore-space-md);
}

.settings-toolbar__field {
  display: flex;
  flex-direction: column;
  gap: var(--kore-space-xs);
  min-width: min(100%, 280px);
}

.settings-toolbar__field label {
  font-size: var(--kore-text-small);
  color: var(--kore-text-muted);
  font-weight: 500;
}

.settings-toolbar__field select {
  background: var(--kore-bg-elevated);
  border: 1px solid var(--kore-border);
  border-radius: var(--kore-radius-md);
  color: var(--kore-text);
  font-family: var(--kore-font);
  font-size: var(--kore-text-body);
  padding: 0.75rem 1rem;
}

.settings-toolbar__pays {
  margin: 0;
  font-size: var(--kore-text-small);
  color: var(--kore-text-muted);
}

.settings-toolbar__actions {
  display: flex;
  flex-wrap: wrap;
  gap: var(--kore-space-sm);
}

.settings-actions {
  display: flex;
  flex-wrap: wrap;
  gap: var(--kore-space-xs);
}

.settings-form {
  margin-top: var(--kore-space-lg);
}

.settings-form__title {
  margin: 0 0 var(--kore-space-md);
  font-size: var(--kore-text-h3);
}

.settings-form__grid {
  display: grid;
  gap: var(--kore-space-md);
  max-width: 420px;
}

.settings-form__actions {
  display: flex;
  flex-wrap: wrap;
  gap: var(--kore-space-sm);
}

.settings-toggle {
  display: flex;
  align-items: center;
  gap: var(--kore-space-sm);
  font-size: var(--kore-text-small);
  color: var(--kore-text);
}

.settings-overlay {
  position: fixed;
  inset: 0;
  z-index: 40;
  display: flex;
  align-items: center;
  justify-content: center;
  padding: var(--kore-space-md);
  background: rgba(0, 0, 0, 0.45);
}

.settings-dialog {
  width: min(100%, 480px);
}

.settings-dialog__text {
  margin: 0 0 var(--kore-space-lg);
  font-size: var(--kore-text-small);
  color: var(--kore-text-muted);
  line-height: 1.5;
}

.settings-flash {
  margin-top: var(--kore-space-md);
  font-size: var(--kore-text-small);
  color: var(--kore-success);
}

.settings-flash--error {
  color: var(--kore-error);
}

@media (max-width: 768px) {
  .settings-toolbar__actions :deep(.app-btn),
  .settings-form__actions :deep(.app-btn) {
    flex: 1 1 calc(50% - var(--kore-space-sm));
  }
}
</style>
