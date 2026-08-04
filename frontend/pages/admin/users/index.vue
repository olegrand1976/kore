<template>
  <div>
    <AppPageHeader :title="$t('users.title')" :subtitle="$t('users.subtitle')">
      <template #actions>
        <AppButton variant="primary" size="sm" @click="openCreate">
          {{ $t('users.add') }}
        </AppButton>
      </template>
    </AppPageHeader>

    <p
      v-if="flash"
      class="users-flash"
      :class="{ 'users-flash--error': flashError }"
      role="status"
    >
      {{ flash }}
    </p>

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

    <AppModal
      :open="showForm"
      width="lg"
      title-id="users-form-title"
      :aria-label="editingId ? $t('users.edit_title') : $t('users.add_title')"
      :close-label="$t('common.cancel')"
      @update:open="(open) => { if (!open) closeForm() }"
    >
      <div class="users-form">
        <h3 id="users-form-title" class="users-form__title">
          {{ editingId ? $t('users.edit_title') : $t('users.add_title') }}
        </h3>

        <div v-if="!editingId" class="users-howto" role="note">
          <p class="users-howto__title">{{ $t('users.howto.title') }}</p>
          <dl class="users-howto__list">
            <div>
              <dt>{{ $t('users.howto.login_title') }}</dt>
              <dd>{{ $t('users.howto.login_body') }}</dd>
            </div>
            <div>
              <dt>{{ $t('users.howto.profile_title') }}</dt>
              <dd>{{ $t('users.howto.profile_body') }}</dd>
            </div>
            <div>
              <dt>{{ $t('users.howto.password_title') }}</dt>
              <dd>{{ $t('users.password_hint') }}</dd>
            </div>
          </dl>
        </div>

        <p v-if="formError" class="users-alert" role="alert">{{ formError }}</p>

        <form class="users-form__grid" novalidate @submit.prevent="save">
          <AppInput
            v-if="!editingId"
            id="user-login"
            v-model="form.login"
            :label="$t('users.login')"
            placeholder="COL_olivier"
            :error="fieldErrors.login"
            required
          />
          <p v-if="!editingId" class="users-hint">{{ $t('users.login_hint') }}</p>
          <AppInput
            v-if="!editingId"
            id="user-password"
            v-model="form.password"
            type="password"
            :label="$t('users.password')"
            :error="fieldErrors.password"
            required
          />
          <p v-if="!editingId" class="users-hint">{{ $t('users.password_hint') }}</p>
          <template v-else>
            <AppInput
              id="user-password-edit"
              v-model="form.password"
              type="password"
              :label="$t('users.password_optional')"
              :error="fieldErrors.password"
            />
            <p v-if="form.password" class="users-hint">{{ $t('users.password_hint') }}</p>
          </template>
          <div class="users-form__field">
            <label for="user-profile">{{ $t('users.profile') }}</label>
            <select id="user-profile" v-model="form.profil" required>
              <option v-for="p in USER_PROFILES" :key="p" :value="p">{{ p }}</option>
            </select>
          </div>
          <div class="users-form__field">
            <label for="user-equipe">{{ $t('users.equipe') }}</label>
            <select id="user-equipe" v-model="form.equipeId">
              <option value="">{{ $t('users.equipe_none') }}</option>
              <option v-for="e in equipeOptions" :key="e.value" :value="e.value">
                {{ e.label }}
              </option>
            </select>
            <p class="users-hint">
              {{ equipeOptions.length ? $t('users.equipe_hint') : $t('users.equipe_empty_hint') }}
            </p>
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
      </div>
    </AppModal>
  </div>
</template>

<script setup lang="ts">
import { USER_PROFILES } from '~/composables/useUsers'
import type { EquipeOption } from '~/composables/useOrganisation'
import { applyTextSearch, useListControls } from '~/composables/useListControls'

definePageMeta({ layout: 'default', middleware: 'admin' })

const { t } = useI18n()
const { extractFetchError } = useApiError()
const { user, fetchSession } = useAuth()
const { list, create, update, deactivate, remove, pickUserId, pickUserLogin, pickUserProfile, pickUserActive, pickUserEquipeId } = useUsers()
const { listEquipes, listApplications, buildEquipeOptions } = useOrganisation()

type UserRow = {
  id: string
  login: string
  profil: string
  active: boolean
  equipeId: string
  equipeLabel: string
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
const fieldErrors = reactive({ login: '', password: '' })

const LOGIN_PATTERN = /^[A-Z]{3}_[a-z0-9_]+$/
const PASSWORD_PATTERN = /^(?=.*[a-z])(?=.*[A-Z])(?=.*\d).{8,}$/

const form = reactive({
  login: '',
  password: '',
  profil: USER_PROFILES[1],
  active: true,
  equipeId: ''
})

// L'équipe porte le rattachement du collaborateur à son application de travail.
const equipeOptions = ref<EquipeOption[]>([])

const equipeLabelById = (id: string) =>
  equipeOptions.value.find((e) => e.value === id)?.label ?? ''

const loadEquipes = async () => {
  try {
    const [eq, apps] = await Promise.all([listEquipes(), listApplications()])
    equipeOptions.value = buildEquipeOptions(eq, apps)
  } catch {
    // Rattachement facultatif : une liste vide n'empêche pas de créer un compte.
    equipeOptions.value = []
  }
}

const clearFieldErrors = () => {
  fieldErrors.login = ''
  fieldErrors.password = ''
}

const validatePassword = (password: string, required: boolean): string => {
  if (!password) {
    return required ? t('users.error_password_required') : ''
  }
  if (!PASSWORD_PATTERN.test(password)) {
    return t('users.error_password_rules')
  }
  return ''
}

const validateForm = (): boolean => {
  clearFieldErrors()
  formError.value = ''

  if (!editingId.value) {
    const login = form.login.trim()
    if (!login) {
      fieldErrors.login = t('users.error_login_required')
    } else if (!LOGIN_PATTERN.test(login)) {
      fieldErrors.login = t('users.error_login_format')
    }
    fieldErrors.password = validatePassword(form.password, true)
  } else if (form.password) {
    fieldErrors.password = validatePassword(form.password, false)
  }

  return !fieldErrors.login && !fieldErrors.password
}

const currentUserId = computed(() => user.value?.userId ?? '')

const columns = computed(() => [
  { key: 'login', label: t('users.login') },
  { key: 'profil', label: t('users.profile') },
  { key: 'equipeLabel', label: t('users.equipe') },
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
  equipe: {
    type: 'select' as const,
    label: t('users.equipe'),
    options: equipeOptions.value,
    match: (row: UserRow, value: string) => row.equipeId === value
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
  items.map((item) => {
    const equipeId = pickUserEquipeId(item)
    return {
      id: pickUserId(item),
      login: pickUserLogin(item),
      profil: pickUserProfile(item),
      active: pickUserActive(item),
      equipeId,
      equipeLabel: equipeId ? equipeLabelById(equipeId) || '—' : '—'
    }
  })

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
  // Les équipes d'abord : loadUsers résout le libellé d'équipe de chaque ligne.
  await loadEquipes()
  await loadUsers()
})

const openCreate = () => {
  editingId.value = ''
  form.login = ''
  form.password = ''
  form.profil = USER_PROFILES[1]
  form.active = true
  form.equipeId = ''
  formError.value = ''
  clearFieldErrors()
  showForm.value = true
}

const openEdit = (row: UserRow) => {
  editingId.value = row.id
  form.login = row.login
  form.password = ''
  form.profil = row.profil
  form.active = row.active
  form.equipeId = row.equipeId
  formError.value = ''
  clearFieldErrors()
  showForm.value = true
}

const closeForm = () => {
  showForm.value = false
  editingId.value = ''
  formError.value = ''
  clearFieldErrors()
}

const save = async () => {
  if (!validateForm()) return
  saving.value = true
  formError.value = ''
  try {
    if (editingId.value) {
      // equipeId toujours envoyé en édition : la chaîne vide détache le collaborateur.
      const body: { profil: string; active: boolean; password?: string; equipeId: string } = {
        profil: form.profil,
        active: form.active,
        equipeId: form.equipeId
      }
      if (form.password) body.password = form.password
      await update(editingId.value, body)
      flash.value = t('users.saved')
      flashError.value = false
    } else {
      await create({
        login: form.login.trim(),
        password: form.password,
        profil: form.profil,
        ...(form.equipeId ? { equipeId: form.equipeId } : {})
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
  const message = String(extractFetchError(err, t('users.error_generic')))
  if (message.includes('login already exists')) return t('users.error_login_exists')
  if (message.includes('invalid login format')) return t('users.error_login_format')
  if (message.includes('weak password')) return t('users.error_password_rules')
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

.users-form__title {
  margin: 0 0 var(--kore-space-md);
  font-size: var(--kore-text-h3);
}

.users-howto {
  margin: 0 0 var(--kore-space-md);
  padding: var(--kore-space-md);
  border: 1px solid var(--kore-border);
  border-radius: var(--kore-radius-md);
  background: var(--kore-bg-elevated);
}

.users-howto__title {
  margin: 0 0 var(--kore-space-sm);
  font-size: var(--kore-text-small);
  font-weight: 600;
}

.users-howto__list {
  margin: 0;
  display: grid;
  gap: var(--kore-space-sm);
}

.users-howto__list dt {
  margin: 0;
  font-size: var(--kore-text-caption);
  font-weight: 600;
  color: var(--kore-text);
}

.users-howto__list dd {
  margin: var(--kore-space-xs) 0 0;
  font-size: var(--kore-text-caption);
  color: var(--kore-text-muted);
  line-height: 1.4;
}

.users-alert {
  margin: 0 0 var(--kore-space-md);
  padding: var(--kore-space-sm) var(--kore-space-md);
  border: 1px solid var(--kore-error);
  border-radius: var(--kore-radius-md);
  background: color-mix(in srgb, var(--kore-error) 12%, transparent);
  color: var(--kore-error);
  font-size: var(--kore-text-small);
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
  margin: 0 0 var(--kore-space-md);
  padding: var(--kore-space-sm) var(--kore-space-md);
  border: 1px solid var(--kore-success);
  border-radius: var(--kore-radius-md);
  background: color-mix(in srgb, var(--kore-success) 12%, transparent);
  color: var(--kore-success);
  font-size: var(--kore-text-small);
}

.users-flash--error {
  border-color: var(--kore-error);
  background: color-mix(in srgb, var(--kore-error) 12%, transparent);
  color: var(--kore-error);
}

@media (max-width: 768px) {
  .users-form__actions :deep(.app-btn),
  .users-actions :deep(.app-btn) {
    flex: 1 1 calc(50% - var(--kore-space-sm));
  }
}
</style>
