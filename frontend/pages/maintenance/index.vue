<template>
  <div>
    <AppPageHeader :title="$t('maintenance.title')">
      <template #actions>
        <AppButton variant="primary" size="sm" @click="showForm = !showForm">
          {{ $t('maintenance.new') }}
        </AppButton>
      </template>
    </AppPageHeader>

    <AppCard v-if="showForm" padding="lg" class="mb">
      <ServiceRequestForm :busy="busy" @submit="onCreate" />
    </AppCard>

    <AppCard v-if="pending" padding="lg"><p class="muted">{{ $t('maintenance.loading') }}</p></AppCard>
    <AppCard v-else-if="!rows.length" padding="lg">
      <AppEmptyState icon="build" :title="$t('maintenance.empty')" />
    </AppCard>
    <AppCard v-else padding="none">
      <AppTable :columns="columns" :rows="rows" row-key="id">
        <template #cell-state="{ value }">
          <AppBadge variant="neutral">{{ maintenanceStateLabel(String(value)) }}</AppBadge>
        </template>
        <template #cell-actions="{ row }">
          <div v-if="row.state === 'created'" class="maintenance-actions">
            <select v-model="assignTargets[row.id]" class="maintenance-actions__select">
              <option value="">{{ $t('requests.assign_to') }}</option>
              <option v-for="u in users" :key="pickUserId(u)" :value="pickUserId(u)">
                {{ pickUserLogin(u) }}
              </option>
            </select>
            <AppButton
              variant="ghost"
              size="sm"
              :disabled="busy || !assignTargets[row.id]"
              @click="onAssign(row.id, assignTargets[row.id])"
            >
              {{ $t('maintenance.assign_self') }}
            </AppButton>
          </div>
          <AppButton
            v-if="row.state === 'assigned'"
            variant="ghost"
            size="sm"
            :disabled="busy"
            @click="onProgress(row.id)"
          >
            {{ $t('maintenance.start') }}
          </AppButton>
          <AppButton
            v-if="row.state === 'in_progress'"
            variant="primary"
            size="sm"
            :disabled="busy"
            @click="onComplete(row.id)"
          >
            {{ $t('maintenance.complete') }}
          </AppButton>
        </template>
      </AppTable>
    </AppCard>
    <p v-if="errorMsg" class="flash flash--error" role="alert">{{ errorMsg }}</p>
  </div>
</template>

<script setup lang="ts">
import type { ServiceRequestPayload } from '~/components/requests/ServiceRequestForm.vue'
import { REQUEST_RESOURCE, useRequestAttachments } from '~/composables/useRequestAttachments'

definePageMeta({ layout: 'default' })

const { t } = useI18n()
const { user } = useAuth()
const { extractFetchError } = useApiError()
const { list, create, assign, progress, complete, pickId, pickSubject, pickState } = useMaintenance()
const { uploadAll } = useRequestAttachments()
const { list: listUsers, pickUserId, pickUserLogin } = useUsers()

const pending = ref(true)
const busy = ref(false)
const showForm = ref(false)
const errorMsg = ref('')
const requests = ref<Awaited<ReturnType<typeof list>>>([])
const users = ref<Awaited<ReturnType<typeof listUsers>>>([])
const assignTargets = reactive<Record<string, string>>({})

const userId = computed(() => user.value?.userId ?? user.value?.id ?? '')

const columns = computed(() => [
  { key: 'subject', label: t('maintenance.col_subject') },
  { key: 'state', label: t('maintenance.col_state') },
  { key: 'actions', label: '' }
])

const rows = computed(() =>
  requests.value.map((wr) => ({
    id: pickId(wr),
    subject: pickSubject(wr),
    state: pickState(wr)
  }))
)

const maintenanceStateLabel = (state: string) => t(`maintenance.state_${state}`, state)

const load = async () => {
  pending.value = true
  errorMsg.value = ''
  try {
    const [workRequests, userList] = await Promise.all([list(), listUsers()])
    requests.value = workRequests
    users.value = userList
    if (userId.value) {
      for (const row of rows.value) {
        if (!assignTargets[row.id]) assignTargets[row.id] = userId.value
      }
    }
  } catch (e) {
    errorMsg.value = extractFetchError(e)
  } finally {
    pending.value = false
  }
}

const onCreate = async (payload: ServiceRequestPayload) => {
  busy.value = true
  try {
    const created = await create(payload)
    const id = pickId(created)
    if (id && payload.files.length) {
      await uploadAll(REQUEST_RESOURCE.maintenance, id, payload.files)
    }
    showForm.value = false
    await load()
  } catch (e) {
    errorMsg.value = extractFetchError(e)
  } finally {
    busy.value = false
  }
}

const onAssign = async (id: string, assigneeId: string) => {
  if (!assigneeId) return
  busy.value = true
  try {
    await assign(id, assigneeId)
    await load()
  } catch (e) {
    errorMsg.value = extractFetchError(e)
  } finally {
    busy.value = false
  }
}

const onProgress = async (id: string) => {
  busy.value = true
  try {
    await progress(id, 1)
    await load()
  } catch (e) {
    errorMsg.value = extractFetchError(e)
  } finally {
    busy.value = false
  }
}

const onComplete = async (id: string) => {
  busy.value = true
  try {
    await complete(id)
    await load()
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
.muted { color: var(--kore-text-muted); }
.flash { margin-top: var(--kore-space-md); font-size: var(--kore-text-small); }
.flash--error { color: var(--kore-status-danger); }
.maintenance-actions {
  display: flex;
  flex-wrap: wrap;
  gap: var(--kore-space-sm);
  align-items: center;
}
.maintenance-actions__select {
  min-width: 10rem;
  padding: 0.5rem 0.75rem;
  border: 1px solid var(--kore-border);
  border-radius: var(--kore-radius-md);
  background: var(--kore-bg-elevated);
  color: var(--kore-text);
}
@media (max-width: 768px) {
  .maintenance-actions {
    flex-direction: column;
    align-items: stretch;
  }
  .maintenance-actions__select { width: 100%; }
}
</style>
