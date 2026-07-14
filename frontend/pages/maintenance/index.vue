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
      <form class="maintenance-form" @submit.prevent="onCreate">
        <AppInput v-model="form.applicationId" :label="$t('maintenance.application_id')" required />
        <AppInput v-model="form.subject" :label="$t('maintenance.col_subject')" required />
        <AppButton variant="primary" size="sm" type="submit" :disabled="busy">
          {{ $t('maintenance.create') }}
        </AppButton>
      </form>
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
          <AppButton
            v-if="row.state === 'created'"
            variant="ghost"
            size="sm"
            :disabled="busy"
            @click="onAssign(row.id)"
          >
            {{ $t('maintenance.assign_self') }}
          </AppButton>
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
definePageMeta({ layout: 'default' })

const { t } = useI18n()
const { user } = useAuth()
const { extractFetchError } = useApiError()
const { list, create, assign, progress, complete, pickId, pickSubject, pickState } = useMaintenance()

const pending = ref(true)
const busy = ref(false)
const showForm = ref(false)
const errorMsg = ref('')
const requests = ref<Awaited<ReturnType<typeof list>>>([])
const form = reactive({ applicationId: '', subject: '' })

const userId = computed(() => user.value?.id ?? '')

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
    requests.value = await list()
  } catch (e) {
    errorMsg.value = extractFetchError(e)
  } finally {
    pending.value = false
  }
}

const onCreate = async () => {
  busy.value = true
  try {
    await create(form)
    showForm.value = false
    form.subject = ''
    await load()
  } catch (e) {
    errorMsg.value = extractFetchError(e)
  } finally {
    busy.value = false
  }
}

const onAssign = async (id: string) => {
  if (!userId.value) return
  busy.value = true
  try {
    await assign(id, userId.value)
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
.maintenance-form {
  display: flex;
  flex-direction: column;
  gap: var(--kore-space-md);
  max-width: var(--kore-form-max);
}
.muted { color: var(--kore-text-muted); }
.flash { margin-top: var(--kore-space-md); font-size: var(--kore-text-small); }
.flash--error { color: var(--kore-status-danger); }
@media (max-width: 768px) {
  .maintenance-form :deep(.app-button) { width: 100%; }
}
</style>
