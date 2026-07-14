<template>
  <div>
    <AppPageHeader :title="$t('users.title')" :subtitle="$t('users.subtitle')">
      <template #actions>
        <AppButton variant="primary" size="sm" @click="openCreate">
          {{ $t('users.add') }}
        </AppButton>
      </template>
    </AppPageHeader>

    <AppCard v-if="pending" padding="lg">
      <p class="muted">{{ $t('users.loading') }}</p>
    </AppCard>

    <AppCard v-else-if="forbidden" padding="lg">
      <AppEmptyState icon="lock" :title="$t('users.forbidden')" />
    </AppCard>

    <template v-else>
      <AppListToolbar
        :filters="listFilters"
        :filter-values="filterValues"
        :sort-keys="sortKeys"
        :sort-key="sortKey"
        :sort-dir="sortDir"
        :has-active-filters="hasActiveFilters"
        @update:filter="setFilter"
        @update:sort-key="setSort($event)"
        @update:sort-dir="setSortDir"
        @reset="resetFilters"
      />
      <AppCard padding="lg">
        <AppTable
          :columns="columns"
          :rows="displayRows"
          :empty-title="hasActiveFilters ? $t('common.list.no_results') : $t('users.empty')"
          row-key="id"
        >
        <template #cell-profil="{ value }">
          <AppBadge variant="default">{{ value }}</AppBadge>
        </template>
        <template #cell-active="{ value }">
          <AppBadge :variant="value ? 'success' : 'default'">
            {{ value ? $t('users.active') : $t('users.inactive') }}
          </AppBadge>
        </template>
        <template #cell-actions="{ row }">
          <div class="users-actions">
            <AppButton variant="ghost" size="sm" @click="openEdit(row)">
              {{ $t('common.edit') }}
            </AppButton>
            <AppButton
              v-if="row.active && row.id !== currentUserId"
              variant="ghost"
              size="sm"
              @click="deactivateRow(row)"
            >
              {{ $t('users.deactivate') }}
            </AppButton>
            <AppButton
              v-if="row.id !== currentUserId"
              variant="ghost"
              size="sm"
              @click="deleteRow(row)"
            >
              {{ $t('common.delete') }}
            </AppButton>
          </div>
        </template>
      </AppTable>
      </AppCard>
    </template>

    <AppCard v-if="showForm" padding="lg" class="users-form">
      <h3 class="users-form__title">
        {{ editingId ? $t('users.edit_title') : $t('users.add_title') }}
      </h3>
      <form class="users-form__grid" @submit.prevent="save">
        <AppInput
          v-if="!editingId"
          id="user-login"
          v-model="form.login"
          :label="$t('users.login')"
          placeholder="COL_dupont"
          required
        />
        <p v-if="!editingId" class="users-hint">{{ $t('users.login_hint') }}</p>
        <AppInput
          v-if="!editingId"
          id="user-password"
          v-model="form.password"
          type="password"
          :label="$t('users.password')"
          required
        />
        <AppInput
          v-else
          id="user-password-edit"
          v-model="form.password"
          type="password"
          :label="$t('users.password_optional')"
        />
        <div class="users-form__field">
          <label for="user-profile">{{ $t('users.profile') }}</label>
          <select id="user-profile" v-model="form.profil" required>
            <option v-for="p in USER_PROFILES" :key="p" :value="p">{{ p }}</option>
          </select>
        </div>
        <label v-if="editingId" class="users-toggle">
          <input v-model="form.active" type="checkbox" />
          {{ $t('users.active') }}
        </label>
        <div class="users-form__actions">
          <AppButton variant="ghost" size="sm" type="button" @click="closeForm">
            {{ $t('common.cancel') }}
          </AppButton>
          <AppButton variant="primary" size="sm" type="submit" :disabled="saving">
            {{ $t('common.save') }}
          </AppButton>
        </div>
      </form>
      <p v-if="formError" class="users-flash users-flash--error" role="alert">{{ formError }}</p>
    </AppCard>

    <p v-if="flash" class="users-flash" :class="{ 'users-flash--error': flashError }" role="status">
      {{ flash }}
    </p>
  </div>
</template>

<script setup lang="ts">
import { USER_PROFILES } from '~/composables/useUsers'
import { applyTextSearch, useListControls } from '~/composables/useListControls'

definePageMeta({ layout: 'default', middleware: 'admin' })

const { t } = useI18n()
const { extractFetchError } = useApiError()
const { user, fetchSession } = useAuth()
const { list, create, update, deactivate, remove, pickUserId, pickUserLogin, pickUserProfile, pickUserActive } = useUsers()

type UserRow = {
  id: string
  login: string
  profil: string
  active: boolean
}

const users = ref<UserRow[]>([])
const pending = ref(true)
const forbidden = ref(false)
const saving = ref(false)
const showForm = ref(false)
const editingId = ref('')
const formError = ref('')
const flash = ref('')
const flashError = ref(false)

const form = reactive({
  login: '',
  password: '',
  profil: USER_PROFILES[1],
  active: true
})

const currentUserId = computed(() => user.value?.userId ?? '')

const columns = computed(() => [
  { key: 'login', label: t('users.login') },
  { key: 'profil', label: t('users.profile') },
  { key: 'active', label: t('users.status') },
  { key: 'actions', label: '' }
])

const rows = computed(() => users.value)

const listFilters = computed(() => ({
  q: {
    type: 'search' as const,
    label: t('common.list.search'),
    placeholder: t('users.login'),
    match: (row: UserRow, query: string) => applyTextSearch(query, row.login)
  },
  profil: {
    type: 'select' as const,
    label: t('users.profile'),
    options: USER_PROFILES.map((p) => ({ value: p, label: p })),
    match: (row: UserRow, value: string) => row.profil === value
  },
  active: {
    type: 'select' as const,
    label: t('users.status'),
    options: [
      { value: 'true', label: t('users.active') },
      { value: 'false', label: t('users.inactive') }
    ],
    match: (row: UserRow, value: string) => String(row.active) === value
  }
}))

const sortKeys = computed(() => [
  { key: 'login', label: t('users.login'), type: 'string' as const, accessor: (row: UserRow) => row.login },
  { key: 'profil', label: t('users.profile'), type: 'string' as const, accessor: (row: UserRow) => row.profil }
])

const {
  filterValues,
  sortKey,
  sortDir,
  sortedItems,
  hasActiveFilters,
  setFilter,
  setSort,
  setSortDir,
  resetFilters
} = useListControls(rows, {
  storageKey: 'admin-users',
  defaultSort: { key: 'login', dir: 'asc' },
  filters: listFilters,
  sortKeys
})

const displayRows = computed(() => sortedItems.value)

const mapUsers = (items: Awaited<ReturnType<typeof list>>) =>
  items.map((item) => ({
    id: pickUserId(item),
    login: pickUserLogin(item),
    profil: pickUserProfile(item),
    active: pickUserActive(item)
  }))

const loadUsers = async () => {
  pending.value = true
  forbidden.value = false
  try {
    users.value = mapUsers(await list())
  } catch (err) {
    if ((err as { statusCode?: number })?.statusCode === 403) {
      forbidden.value = true
      users.value = []
    } else {
      flash.value = extractFetchError(err)
      flashError.value = true
    }
  } finally {
    pending.value = false
  }
}

onMounted(async () => {
  await fetchSession()
  await loadUsers()
})

const openCreate = () => {
  editingId.value = ''
  form.login = ''
  form.password = ''
  form.profil = USER_PROFILES[1]
  form.active = true
  formError.value = ''
  showForm.value = true
}

const openEdit = (row: UserRow) => {
  editingId.value = row.id
  form.login = row.login
  form.password = ''
  form.profil = row.profil
  form.active = row.active
  formError.value = ''
  showForm.value = true
}

const closeForm = () => {
  showForm.value = false
  editingId.value = ''
}

const save = async () => {
  saving.value = true
  formError.value = ''
  try {
    if (editingId.value) {
      const body: { profil: string; active: boolean; password?: string } = {
        profil: form.profil,
        active: form.active
      }
      if (form.password) body.password = form.password
      await update(editingId.value, body)
      flash.value = t('users.saved')
      flashError.value = false
    } else {
      await create({
        login: form.login.trim(),
        password: form.password,
        profil: form.profil
      })
      flash.value = t('users.created')
      flashError.value = false
    }
    closeForm()
    await loadUsers()
  } catch (err) {
    formError.value = mapUserError(err)
  } finally {
    saving.value = false
  }
}

const deactivateRow = async (row: UserRow) => {
  if (!confirm(t('users.deactivate_confirm', { login: row.login }))) return
  try {
    await deactivate(row.id)
    flash.value = t('users.deactivated')
    flashError.value = false
    await loadUsers()
  } catch (err) {
    flash.value = mapUserError(err)
    flashError.value = true
  }
}

const deleteRow = async (row: UserRow) => {
  if (!confirm(t('users.delete_confirm', { login: row.login }))) return
  try {
    await remove(row.id)
    flash.value = t('users.deleted')
    flashError.value = false
    await loadUsers()
  } catch (err) {
    flash.value = mapUserError(err)
    flashError.value = true
  }
}

function mapUserError(err: unknown) {
  const message = extractFetchError(err, t('users.error_generic'))
  if (message.includes('login already exists')) return t('users.error_login_exists')
  if (message.includes('invalid login format')) return t('users.error_login_format')
  if (message.includes('seat limit reached')) return t('users.error_seat_limit')
  if (message.includes('cannot modify own account')) return t('users.error_self')
  return message
}
</script>

<style scoped>
.muted { color: var(--kore-text-muted); }

.users-actions {
  display: flex;
  flex-wrap: wrap;
  gap: var(--kore-space-xs);
}

.users-form {
  margin-top: var(--kore-space-lg);
}

.users-form__title {
  margin: 0 0 var(--kore-space-md);
  font-size: var(--kore-text-h3);
}

.users-form__grid {
  display: grid;
  gap: var(--kore-space-md);
  max-width: var(--kore-form-max);
}

.users-form__field {
  display: flex;
  flex-direction: column;
  gap: var(--kore-space-xs);
}

.users-form__field label {
  font-size: var(--kore-text-small);
  color: var(--kore-text-muted);
  font-weight: 500;
}

.users-form__field select {
  background: var(--kore-bg-elevated);
  border: 1px solid var(--kore-border);
  border-radius: var(--kore-radius-md);
  color: var(--kore-text);
  font-family: var(--kore-font);
  font-size: var(--kore-text-body);
  padding: 0.75rem 1rem;
}

.users-form__actions {
  display: flex;
  flex-wrap: wrap;
  gap: var(--kore-space-sm);
}

.users-toggle {
  display: flex;
  align-items: center;
  gap: var(--kore-space-sm);
  font-size: var(--kore-text-small);
  color: var(--kore-text);
}

.users-hint {
  margin: calc(-1 * var(--kore-space-sm)) 0 0;
  font-size: var(--kore-text-caption);
  color: var(--kore-text-muted);
}

.users-flash {
  margin-top: var(--kore-space-md);
  font-size: var(--kore-text-small);
  color: var(--kore-success);
}

.users-flash--error {
  color: var(--kore-error);
}

@media (max-width: 768px) {
  .users-form__actions :deep(.app-btn),
  .users-actions :deep(.app-btn) {
    flex: 1 1 calc(50% - var(--kore-space-sm));
  }
}
</style>
