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
          <AppButton variant="ghost" size="sm" @click="navigateTo(`/maintenance/${row.id}`)">
            {{ $t('maintenance.open') }}
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
const { extractFetchError } = useApiError()
const { list, create, pickId, pickSubject, pickState } = useMaintenance()
const { uploadAll } = useRequestAttachments()

const pending = ref(true)
const busy = ref(false)
const showForm = ref(false)
const errorMsg = ref('')
const requests = ref<Awaited<ReturnType<typeof list>>>([])

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

await load()
</script>

<style scoped>
.mb { margin-bottom: var(--kore-space-lg); }
.muted { color: var(--kore-text-muted); }
.flash { margin-top: var(--kore-space-md); font-size: var(--kore-text-small); }
.flash--error { color: var(--kore-status-danger); }
</style>
