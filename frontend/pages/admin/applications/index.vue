<template>
  <div>
    <AppPageHeader :title="$t('applications.title')" :subtitle="$t('applications.subtitle')">
      <template #actions>
        <AppButton variant="primary" size="sm" type="button" @click="openCreate">
          <AppIcon name="add" /> {{ $t('applications.add') }}
        </AppButton>
      </template>
    </AppPageHeader>

    <p
      v-if="flash"
      class="apps-flash"
      :class="{ 'apps-flash--error': flashError }"
      role="status"
    >
      {{ flash }}
    </p>

    <AppCard v-if="pending" padding="lg">
      <p class="muted">{{ $t('applications.loading') }}</p>
    </AppCard>

    <AppCard v-else-if="forbidden" padding="lg">
      <AppEmptyState icon="lock" :title="$t('applications.forbidden')" />
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
          :empty-title="hasActiveFilters ? $t('common.list.no_results') : $t('applications.empty')"
          row-key="id"
        >
          <template #cell-active="{ value }">
            <AppBadge :variant="value ? 'success' : 'warning'">
              {{ value ? $t('applications.active') : $t('applications.inactive') }}
            </AppBadge>
          </template>
          <template #cell-actions="{ row }">
            <div class="apps-actions">
              <AppButton variant="ghost" size="sm" type="button" @click="openEdit(row)">
                {{ $t('common.edit') }}
              </AppButton>
              <AppButton
                v-if="row.active"
                variant="ghost"
                size="sm"
                type="button"
                :disabled="actionBusyId === row.id"
                @click="deactivateRow(row)"
              >
                {{ $t('applications.deactivate') }}
              </AppButton>
              <AppButton
                v-else
                variant="ghost"
                size="sm"
                type="button"
                :disabled="actionBusyId === row.id"
                @click="activateRow(row)"
              >
                {{ $t('applications.activate') }}
              </AppButton>
            </div>
          </template>
        </AppTable>
      </AppCard>
    </template>

    <AppModal
      v-model:open="showForm"
      width="lg"
      :aria-label="editingId ? $t('applications.edit_title') : $t('applications.create_title')"
    >
      <form class="apps-form" @submit.prevent="submitForm">
        <h2 class="apps-form__title">
          {{ editingId ? $t('applications.edit_title') : $t('applications.create_title') }}
        </h2>

        <label class="apps-form__field">
          <span>{{ $t('applications.field_service') }}</span>
          <select v-model="form.serviceId" class="apps-form__input" :disabled="!!editingId" required>
            <option value="" disabled>{{ $t('applications.field_service_placeholder') }}</option>
            <option v-for="s in serviceOptions" :key="s.value" :value="s.value">{{ s.label }}</option>
          </select>
        </label>

        <label class="apps-form__field">
          <span>{{ $t('applications.field_libelle') }}</span>
          <input v-model="form.libelle" class="apps-form__input" type="text" required maxlength="120" />
        </label>

        <label class="apps-form__field">
          <span>{{ $t('applications.field_proprietaire') }}</span>
          <input v-model="form.proprietaire" class="apps-form__input" type="text" maxlength="120" />
        </label>

        <label class="apps-form__field">
          <span>{{ $t('applications.field_mode') }}</span>
          <select v-model="form.modeFacturation" class="apps-form__input">
            <option value="temps_passe">{{ $t('applications.mode_temps_passe') }}</option>
            <option value="forfait">{{ $t('applications.mode_forfait') }}</option>
            <option value="non">{{ $t('applications.mode_non') }}</option>
          </select>
        </label>

        <label class="apps-form__toggle">
          <input v-model="form.uoActivee" type="checkbox" />
          <span>{{ $t('applications.field_uo') }}</span>
        </label>

        <label class="apps-form__field">
          <span>{{ $t('applications.field_chef') }}</span>
          <select v-model="form.chefUtilisateurId" class="apps-form__input">
            <option value="">{{ $t('applications.field_chef_none') }}</option>
            <option v-for="u in userOptions" :key="u.value" :value="u.value">{{ u.label }}</option>
          </select>
        </label>

        <template v-if="editingId">
          <section class="apps-form__section">
            <h3>{{ $t('applications.section_equipes') }}</h3>
            <ul v-if="editEquipes.length" class="apps-form__list">
              <li v-for="e in editEquipes" :key="e.id">{{ e.label }}</li>
            </ul>
            <p v-else class="muted">{{ $t('applications.empty_equipes') }}</p>
            <div class="apps-form__inline">
              <input
                v-model="newEquipeLibelle"
                class="apps-form__input"
                type="text"
                :placeholder="$t('applications.new_equipe_placeholder')"
              />
              <AppButton
                variant="secondary"
                size="sm"
                type="button"
                :disabled="!newEquipeLibelle.trim()"
                @click="addEquipe"
              >
                {{ $t('applications.add_equipe') }}
              </AppButton>
            </div>
          </section>

          <section class="apps-form__section">
            <h3>{{ $t('applications.section_users') }}</h3>
            <p class="muted">{{ $t('applications.users_hint') }}</p>
            <template v-if="editEquipes.length">
              <label class="apps-form__field">
                <span>{{ $t('applications.membership_equipe') }}</span>
                <select v-model="membershipEquipeId" class="apps-form__input">
                  <option v-for="e in editEquipes" :key="e.id" :value="e.id">{{ e.label }}</option>
                </select>
              </label>
              <fieldset v-if="membershipEquipeId" class="apps-form__checkgroup">
                <legend>{{ $t('applications.membership_users') }}</legend>
                <label v-for="u in userOptions" :key="u.value" class="apps-form__check">
                  <input
                    type="checkbox"
                    :checked="userInMembershipEquipe(u.value)"
                    :disabled="membershipBusyId === u.value"
                    @change="toggleUserMembership(u.value, ($event.target as HTMLInputElement).checked)"
                  />
                  <span>{{ u.label }}</span>
                </label>
              </fieldset>
            </template>
            <p v-else class="muted">{{ $t('applications.membership_need_equipe') }}</p>
            <NuxtLink class="apps-form__link" to="/admin/users">{{ $t('applications.manage_users') }}</NuxtLink>
          </section>

          <section class="apps-form__section">
            <h3>{{ $t('applications.section_budgets') }}</h3>
            <ul v-if="editBudgets.length" class="apps-form__list">
              <li v-for="b in editBudgets" :key="b.id">
                {{ b.label }}
                <AppBadge v-if="b.isDefault || b.id === form.budgetDefautId" variant="success">
                  {{ $t('applications.budget_default') }}
                </AppBadge>
              </li>
            </ul>
            <p v-else class="muted">{{ $t('applications.empty_budgets') }}</p>
            <label v-if="defaultBudgetOptions.length" class="apps-form__field">
              <span>{{ $t('applications.field_budget_defaut') }}</span>
              <select v-model="form.budgetDefautId" class="apps-form__input">
                <option value="">{{ $t('applications.field_budget_defaut_none') }}</option>
                <option v-for="b in defaultBudgetOptions" :key="b.id" :value="b.id">{{ b.label }}</option>
              </select>
              <span class="muted apps-form__hint">{{ $t('applications.field_budget_defaut_hint') }}</span>
            </label>
            <p v-else-if="editBudgets.length" class="muted apps-form__hint">
              {{ $t('applications.field_budget_defaut_hint') }}
            </p>
            <AppButton variant="secondary" size="sm" type="button" @click="goCreateBudget">
              {{ $t('applications.create_budget') }}
            </AppButton>
          </section>
        </template>

        <p v-if="formError" class="apps-form__error" role="alert">{{ formError }}</p>

        <div class="apps-form__actions">
          <AppButton variant="ghost" type="button" @click="closeForm">{{ $t('common.cancel') }}</AppButton>
          <AppButton variant="primary" type="submit" :disabled="saving">
            {{ saving ? $t('common.saving') : $t('common.save') }}
          </AppButton>
        </div>
      </form>
    </AppModal>
  </div>
</template>

<script setup lang="ts">
import { applyTextSearch, useListControls } from '~/composables/useListControls'
import {
  defaultBudgetsForApplication,
  isDefaultBudgetType
} from '~/composables/useApplications'
import type { OrgEquipe } from '~/composables/useOrganisation'
import type { BudgetItem } from '~/composables/useBudget'
import type { OrgUserSummary } from '~/composables/useUsers'

definePageMeta({ layout: 'default', middleware: 'admin' })

const { t } = useI18n()
const route = useRoute()
const { extractFetchError } = useApiError()
const {
  list,
  create,
  update,
  deactivate,
  activate,
  pickAppId,
  pickAppLabel,
  pickAppClient,
  pickAppActive,
  pickAppServiceId,
  pickAppMode,
  pickAppChefId,
  pickAppBudgetDefautId
} = useApplications()
const { listServices, listEquipes, createEquipe, orgId, orgLabel } = useOrganisation()
const { list: listUsers, update: updateUser, pickUserId, pickUserLogin, pickUserEquipeIds } = useUsers()
const { list: listBudgets, pickId: pickBudgetId } = useBudget()

type AppRow = {
  id: string
  libelle: string
  serviceId: string
  serviceLabel: string
  proprietaire: string
  mode: string
  modeLabel: string
  uoActivee: boolean
  active: boolean
  equipeCount: number
  chefUtilisateurId: string
  budgetDefautId: string
}

const rows = ref<AppRow[]>([])
const pending = ref(true)
const forbidden = ref(false)
const saving = ref(false)
const showForm = ref(false)
const editingId = ref('')
const formError = ref('')
const flash = ref('')
const flashError = ref(false)
const actionBusyId = ref('')
const membershipBusyId = ref('')
const membershipEquipeId = ref('')
const newEquipeLibelle = ref('')

const serviceOptions = ref<{ value: string; label: string }[]>([])
const equipes = ref<OrgEquipe[]>([])
const users = ref<OrgUserSummary[]>([])
const budgets = ref<BudgetItem[]>([])

const form = reactive({
  serviceId: '',
  libelle: '',
  proprietaire: '',
  modeFacturation: 'temps_passe',
  uoActivee: false,
  chefUtilisateurId: '',
  budgetDefautId: ''
})

const modeLabel = (mode: string) => {
  if (mode === 'non') return t('applications.mode_non')
  if (mode === 'forfait') return t('applications.mode_forfait')
  return t('applications.mode_temps_passe')
}

const columns = computed(() => [
  { key: 'libelle', label: t('applications.col_libelle') },
  { key: 'serviceLabel', label: t('applications.col_service') },
  { key: 'proprietaire', label: t('applications.col_proprietaire') },
  { key: 'modeLabel', label: t('applications.col_mode') },
  { key: 'equipeCount', label: t('applications.col_equipes') },
  { key: 'active', label: t('applications.col_status') },
  { key: 'actions', label: t('common.actions') }
])

const listFilters = computed(() => ({
  q: {
    type: 'search' as const,
    label: t('common.list.search'),
    placeholder: t('applications.search_placeholder'),
    match: (item: AppRow, query: string) =>
      applyTextSearch(query, item.libelle, item.proprietaire, item.serviceLabel)
  },
  active: {
    type: 'select' as const,
    label: t('applications.col_status'),
    options: [
      { value: 'true', label: t('applications.active') },
      { value: 'false', label: t('applications.inactive') }
    ],
    match: (item: AppRow, value: string) => {
      if (!value) return true
      return String(item.active) === value
    }
  },
  serviceId: {
    type: 'select' as const,
    label: t('applications.col_service'),
    options: serviceOptions.value,
    match: (item: AppRow, value: string) => !value || item.serviceId === value
  }
}))

const sortKeys = computed(() => [
  {
    key: 'libelle',
    label: t('applications.col_libelle'),
    type: 'string' as const,
    accessor: (item: AppRow) => item.libelle
  },
  {
    key: 'serviceLabel',
    label: t('applications.col_service'),
    type: 'string' as const,
    accessor: (item: AppRow) => item.serviceLabel
  },
  {
    key: 'equipeCount',
    label: t('applications.col_equipes'),
    type: 'number' as const,
    accessor: (item: AppRow) => item.equipeCount
  }
])

const {
  filterValues,
  sortKey,
  sortDir,
  hasActiveFilters,
  setFilter,
  setSort,
  setSortDir,
  resetFilters,
  sortedItems
} = useListControls(rows, {
  storageKey: 'admin-applications',
  filters: listFilters,
  sortKeys,
  defaultSort: { key: 'libelle', dir: 'asc' }
})

const displayRows = computed(() => sortedItems.value)

const userOptions = computed(() =>
  users.value.map((u) => ({ value: pickUserId(u), label: pickUserLogin(u) }))
)

const editEquipes = computed(() => {
  if (!editingId.value) return []
  return equipes.value
    .filter((e) => (e.applicationId ?? e.ApplicationID) === editingId.value)
    .map((e) => ({ id: orgId(e), label: orgLabel(e) }))
})

watch(
  editEquipes,
  (list) => {
    if (!list.length) {
      membershipEquipeId.value = ''
      return
    }
    if (!list.some((e) => e.id === membershipEquipeId.value)) {
      membershipEquipeId.value = list[0].id
    }
  },
  { immediate: true }
)

const userInMembershipEquipe = (userId: string) => {
  const user = users.value.find((u) => pickUserId(u) === userId)
  if (!user || !membershipEquipeId.value) return false
  return pickUserEquipeIds(user).includes(membershipEquipeId.value)
}

const toggleUserMembership = async (userId: string, checked: boolean) => {
  const equipeId = membershipEquipeId.value
  if (!equipeId) return
  const user = users.value.find((u) => pickUserId(u) === userId)
  if (!user) return
  const current = pickUserEquipeIds(user)
  const next = checked
    ? Array.from(new Set([...current, equipeId]))
    : current.filter((id) => id !== equipeId)
  membershipBusyId.value = userId
  try {
    await updateUser(userId, { equipeIds: next })
    users.value = await listUsers()
    flash.value = t('applications.membership_updated')
    flashError.value = false
  } catch (err) {
    formError.value = extractFetchError(err)
  } finally {
    membershipBusyId.value = ''
  }
}

const editBudgets = computed(() => {
  if (!editingId.value) return []
  return budgets.value
    .filter((b) => (b.applicationId ?? b.ApplicationID) === editingId.value)
    .map((b) => {
      const type = String(b.type ?? b.Type ?? '').toLowerCase()
      return {
        id: pickBudgetId(b),
        label: type || pickBudgetId(b),
        isDefault: isDefaultBudgetType(type)
      }
    })
})

const defaultBudgetOptions = computed(() => {
  if (!editingId.value) return []
  return defaultBudgetsForApplication(budgets.value, editingId.value).map((b) => {
    const type = String(b.type ?? b.Type ?? '').toLowerCase()
    return {
      id: pickBudgetId(b),
      label: type || pickBudgetId(b)
    }
  })
})

const suggestDefaultBudgetIfEmpty = () => {
  if (!editingId.value || form.budgetDefautId) return
  const first = defaultBudgetOptions.value[0]
  if (first) form.budgetDefautId = first.id
}

const loadAll = async () => {
  pending.value = true
  forbidden.value = false
  try {
    const [svcList, appList, equipeList, userList, budgetList] = await Promise.all([
      listServices(),
      list({ active: 'all' }),
      listEquipes(),
      listUsers(),
      listBudgets()
    ])
    serviceOptions.value = svcList.map((s) => ({
      value: orgId(s),
      label: orgLabel(s) || orgId(s)
    }))
    equipes.value = equipeList
    users.value = userList
    budgets.value = budgetList
    const serviceMap = new Map(serviceOptions.value.map((s) => [s.value, s.label]))
    rows.value = appList.map((app) => {
      const id = pickAppId(app)
      const serviceId = pickAppServiceId(app)
      const mode = pickAppMode(app)
      return {
        id,
        libelle: pickAppLabel(app),
        serviceId,
        serviceLabel: serviceMap.get(serviceId) ?? serviceId,
        proprietaire: pickAppClient(app),
        mode,
        modeLabel: modeLabel(mode),
        uoActivee: app.uoActivee ?? app.UOActivee ?? false,
        active: pickAppActive(app),
        equipeCount: equipeList.filter((e) => (e.applicationId ?? e.ApplicationID) === id).length,
        chefUtilisateurId: pickAppChefId(app),
        budgetDefautId: pickAppBudgetDefautId(app)
      }
    })
    suggestDefaultBudgetIfEmpty()
  } catch (err) {
    if ((err as { statusCode?: number })?.statusCode === 403) {
      forbidden.value = true
      rows.value = []
    } else {
      flash.value = extractFetchError(err)
      flashError.value = true
    }
  } finally {
    pending.value = false
  }
}

const openCreate = () => {
  editingId.value = ''
  form.serviceId = serviceOptions.value[0]?.value ?? ''
  form.libelle = ''
  form.proprietaire = ''
  form.modeFacturation = 'temps_passe'
  form.uoActivee = false
  form.chefUtilisateurId = ''
  form.budgetDefautId = ''
  formError.value = ''
  newEquipeLibelle.value = ''
  showForm.value = true
}

const openEdit = (row: AppRow) => {
  editingId.value = row.id
  form.serviceId = row.serviceId
  form.libelle = row.libelle
  form.proprietaire = row.proprietaire
  form.modeFacturation = row.mode || 'temps_passe'
  form.uoActivee = row.uoActivee
  form.chefUtilisateurId = row.chefUtilisateurId
  form.budgetDefautId = row.budgetDefautId
  formError.value = ''
  newEquipeLibelle.value = ''
  showForm.value = true
  suggestDefaultBudgetIfEmpty()
}

const closeForm = () => {
  showForm.value = false
  editingId.value = ''
  formError.value = ''
}

const submitForm = async () => {
  formError.value = ''
  if (!form.libelle.trim() || (!editingId.value && !form.serviceId)) {
    formError.value = t('applications.validation_required')
    return
  }
  saving.value = true
  try {
    const chef = form.chefUtilisateurId || null
    const budgetDefaut = form.budgetDefautId || null
    if (editingId.value) {
      await update(editingId.value, {
        libelle: form.libelle.trim(),
        proprietaire: form.proprietaire.trim(),
        modeFacturation: form.modeFacturation,
        uoActivee: form.uoActivee,
        chefUtilisateurId: chef,
        budgetDefautId: budgetDefaut
      })
      flash.value = t('applications.updated')
    } else {
      await create({
        serviceId: form.serviceId,
        libelle: form.libelle.trim(),
        proprietaire: form.proprietaire.trim(),
        modeFacturation: form.modeFacturation,
        uoActivee: form.uoActivee,
        chefUtilisateurId: chef || undefined
      })
      flash.value = t('applications.created')
    }
    flashError.value = false
    closeForm()
    await loadAll()
  } catch (err) {
    formError.value = extractFetchError(err)
  } finally {
    saving.value = false
  }
}

const addEquipe = async () => {
  if (!editingId.value || !newEquipeLibelle.value.trim()) return
  try {
    await createEquipe({
      applicationId: editingId.value,
      libelle: newEquipeLibelle.value.trim()
    })
    newEquipeLibelle.value = ''
    flash.value = t('applications.equipe_created')
    flashError.value = false
    await loadAll()
  } catch (err) {
    formError.value = extractFetchError(err)
  }
}

const goCreateBudget = () => {
  if (!editingId.value) return
  navigateTo(`/budget?create=1&applicationId=${editingId.value}`)
}

const deactivateRow = async (row: AppRow) => {
  if (!confirm(t('applications.deactivate_confirm', { name: row.libelle }))) return
  actionBusyId.value = row.id
  try {
    await deactivate(row.id)
    flash.value = t('applications.deactivated')
    flashError.value = false
    await loadAll()
  } catch (err) {
    flash.value = extractFetchError(err)
    flashError.value = true
  } finally {
    actionBusyId.value = ''
  }
}

const activateRow = async (row: AppRow) => {
  actionBusyId.value = row.id
  try {
    await activate(row.id)
    flash.value = t('applications.activated')
    flashError.value = false
    await loadAll()
  } catch (err) {
    flash.value = extractFetchError(err)
    flashError.value = true
  } finally {
    actionBusyId.value = ''
  }
}

onMounted(async () => {
  await loadAll()
  const id = typeof route.query.id === 'string' ? route.query.id : ''
  if (id) {
    const row = rows.value.find((a) => a.id === id)
    if (row) openEdit(row)
  }
})
</script>

<style scoped>
.apps-flash {
  margin: 0 0 var(--kore-space-3);
  padding: var(--kore-space-2) var(--kore-space-3);
  border-radius: var(--kore-radius-md);
  background: color-mix(in srgb, var(--kore-success) 18%, transparent);
  color: var(--kore-text);
}
.apps-flash--error {
  background: color-mix(in srgb, var(--kore-danger) 18%, transparent);
}
.apps-actions {
  display: flex;
  flex-wrap: wrap;
  gap: var(--kore-space-1);
}
.apps-form {
  display: flex;
  flex-direction: column;
  gap: var(--kore-space-3);
  max-width: var(--kore-form-wide-max);
}
.apps-form__title {
  margin: 0;
  font-size: var(--kore-font-size-lg);
}
.apps-form__field {
  display: flex;
  flex-direction: column;
  gap: var(--kore-space-1);
}
.apps-form__input {
  width: 100%;
  padding: var(--kore-space-2);
  border: 1px solid var(--kore-border);
  border-radius: var(--kore-radius-md);
  background: var(--kore-surface);
  color: var(--kore-text);
}
.apps-form__toggle {
  display: flex;
  align-items: center;
  gap: var(--kore-space-2);
}
.apps-form__section {
  display: flex;
  flex-direction: column;
  gap: var(--kore-space-2);
  padding-top: var(--kore-space-2);
  border-top: 1px solid var(--kore-border);
}
.apps-form__section h3 {
  margin: 0;
  font-size: var(--kore-font-size-md);
}
.apps-form__list {
  margin: 0;
  padding-left: 1.2rem;
}
.apps-form__inline {
  display: flex;
  flex-wrap: wrap;
  gap: var(--kore-space-2);
}
.apps-form__checkgroup {
  display: flex;
  flex-direction: column;
  gap: var(--kore-space-1);
  margin: 0;
  padding: 0;
  border: none;
}
.apps-form__check {
  display: flex;
  align-items: center;
  gap: var(--kore-space-2);
}
.apps-form__link {
  color: var(--kore-accent);
  text-decoration: underline;
}
.apps-form__error {
  color: var(--kore-danger);
  margin: 0;
}
.apps-form__actions {
  display: flex;
  flex-wrap: wrap;
  gap: var(--kore-space-2);
  justify-content: flex-end;
}
.muted {
  color: var(--kore-text-muted);
  margin: 0;
}
@media (max-width: 768px) {
  .apps-form__actions {
    flex-direction: column;
  }
  .apps-form__actions :deep(.app-button) {
    width: 100%;
  }
}
</style>
