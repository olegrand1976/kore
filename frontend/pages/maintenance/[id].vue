<template>
  <div>
    <AppPageHeader :title="pageTitle">
      <template #actions>
        <AppButton variant="ghost" size="sm" @click="navigateTo('/maintenance')">
          {{ $t('maintenance.back') }}
        </AppButton>
      </template>
    </AppPageHeader>

    <p v-if="errorMsg" class="flash flash--error" role="alert">{{ errorMsg }}</p>
    <AppCard v-if="pending" padding="lg"><p class="muted">{{ $t('maintenance.loading') }}</p></AppCard>

    <template v-else-if="workRequest">
      <AppCard padding="lg" class="mb">
        <dl class="meta">
          <div>
            <dt>{{ $t('maintenance.col_state') }}</dt>
            <dd><AppBadge variant="neutral">{{ stateLabel }}</AppBadge></dd>
          </div>
          <div>
            <dt>{{ $t('requests.col_application') }}</dt>
            <dd>{{ applicationLabel }}</dd>
          </div>
          <div>
            <dt>{{ $t('requests.col_priority') }}</dt>
            <dd>{{ priorityLabel }}</dd>
          </div>
          <div>
            <dt>{{ $t('requests.col_due_at') }}</dt>
            <dd>{{ dueAtLabel }}</dd>
          </div>
          <div v-if="description">
            <dt>{{ $t('requests.form_description') }}</dt>
            <dd>{{ description }}</dd>
          </div>
        </dl>

        <div v-if="canWrite && state === 'created'" class="assign-block">
          <label for="maintenance-assignee" class="assign-block__label">{{ $t('requests.assign_to') }}</label>
          <div class="assign-block__row">
            <select id="maintenance-assignee" v-model="assigneeId" class="assign-block__select">
              <option value="">{{ $t('requests.assign_to') }}</option>
              <option v-for="u in users" :key="pickUserId(u)" :value="pickUserId(u)">
                {{ pickUserLogin(u) }}
              </option>
            </select>
            <AppButton variant="primary" size="sm" :disabled="busy || !assigneeId" @click="onAssign">
              {{ $t('maintenance.assign_action') }}
            </AppButton>
          </div>
        </div>

        <div class="actions">
          <AppButton
            v-if="canWrite && state === 'assigned'"
            variant="ghost"
            size="sm"
            :disabled="busy"
            @click="onProgress"
          >
            {{ $t('maintenance.start') }}
          </AppButton>
          <AppButton
            v-if="canWrite && state === 'in_progress'"
            variant="primary"
            size="sm"
            :disabled="busy"
            @click="onComplete"
          >
            {{ $t('maintenance.complete') }}
          </AppButton>
        </div>
      </AppCard>

      <RequestAttachmentsPanel
        resource="maintenance"
        :resource-id="id"
        :can-upload="canWrite"
      />
    </template>
  </div>
</template>

<script setup lang="ts">
definePageMeta({ layout: 'default' })

const route = useRoute()
const { t } = useI18n()
const { user } = useAuth()
const { can } = usePermissions()
const { extractFetchError } = useApiError()
const {
  get,
  assign,
  progress,
  complete,
  pickSubject,
  pickState,
  pickDescription,
  pickPriority,
  pickDueAt,
  pickApplicationId
} = useMaintenance()
const { list: listUsers, pickUserId, pickUserLogin } = useUsers()
const { list: listApps, pickAppLabel, appById } = useApplications()

const id = computed(() => String(route.params.id))
const pending = ref(true)
const busy = ref(false)
const errorMsg = ref('')
const workRequest = ref<Awaited<ReturnType<typeof get>> | null>(null)
const users = ref<Awaited<ReturnType<typeof listUsers>>>([])
const apps = ref<Awaited<ReturnType<typeof listApps>>>([])
const assigneeId = ref('')

const canWrite = computed(() => can('maintenance', 'E'))
const userId = computed(() => user.value?.userId ?? user.value?.id ?? '')

const pageTitle = computed(() => pickSubject(workRequest.value ?? {}) || t('maintenance.title'))
const state = computed(() => pickState(workRequest.value ?? {}))
const stateLabel = computed(() => t(`maintenance.state_${state.value}`, state.value))
const description = computed(() => pickDescription(workRequest.value ?? {}))
const priorityLabel = computed(() => {
  const p = pickPriority(workRequest.value ?? {})
  return t(`requests.priority_${p}` as const, p)
})
const dueAtLabel = computed(() => {
  const raw = pickDueAt(workRequest.value ?? {})
  if (!raw) return '—'
  const d = new Date(raw)
  return Number.isNaN(d.getTime()) ? '—' : d.toLocaleString()
})
const applicationLabel = computed(() => {
  const appId = pickApplicationId(workRequest.value ?? {})
  if (!appId) return '—'
  return pickAppLabel(appById(apps.value).get(appId)) || appId
})

const load = async () => {
  pending.value = true
  errorMsg.value = ''
  try {
    const [loaded, userList, appList] = await Promise.all([get(id.value), listUsers(), listApps()])
    workRequest.value = loaded
    users.value = userList
    apps.value = appList
    assigneeId.value = userId.value
  } catch (e) {
    errorMsg.value = extractFetchError(e)
  } finally {
    pending.value = false
  }
}

const onAssign = async () => {
  if (!assigneeId.value) return
  busy.value = true
  try {
    workRequest.value = await assign(id.value, assigneeId.value)
  } catch (e) {
    errorMsg.value = extractFetchError(e)
  } finally {
    busy.value = false
  }
}

const onProgress = async () => {
  busy.value = true
  try {
    workRequest.value = await progress(id.value, 1)
  } catch (e) {
    errorMsg.value = extractFetchError(e)
  } finally {
    busy.value = false
  }
}

const onComplete = async () => {
  busy.value = true
  try {
    workRequest.value = await complete(id.value)
  } catch (e) {
    errorMsg.value = extractFetchError(e)
  } finally {
    busy.value = false
  }
}

await load()
</script>

<style scoped>
.mb { margin-bottom: var(--kore-space-lg); }
.meta { display: grid; gap: var(--kore-space-md); margin: 0 0 var(--kore-space-lg); }
.meta dt { font-size: var(--kore-text-small); color: var(--kore-text-muted); }
.meta dd { margin: 0.25rem 0 0; }
.assign-block { display: grid; gap: var(--kore-space-sm); margin-bottom: var(--kore-space-lg); }
.assign-block__label { font-size: var(--kore-text-small); color: var(--kore-text-muted); font-weight: 500; }
.assign-block__row { display: flex; flex-wrap: wrap; gap: var(--kore-space-sm); align-items: center; }
.assign-block__select {
  min-width: 12rem;
  padding: 0.5rem 0.75rem;
  border: 1px solid var(--kore-border);
  border-radius: var(--kore-radius-md);
  background: var(--kore-bg-elevated);
  color: var(--kore-text);
}
.actions { display: flex; flex-wrap: wrap; gap: var(--kore-space-sm); }
.muted { color: var(--kore-text-muted); }
.flash { margin-bottom: var(--kore-space-md); font-size: var(--kore-text-small); }
.flash--error { color: var(--kore-status-danger); }
@media (max-width: 768px) {
  .actions :deep(.app-button),
  .assign-block__row :deep(.app-button) { width: 100%; }
  .assign-block__select { width: 100%; }
  .assign-block__row { flex-direction: column; align-items: stretch; }
}
</style>
