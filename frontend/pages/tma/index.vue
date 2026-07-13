<template>
  <div>
    <AppPageHeader :title="$t('tma.title')">
      <template #actions>
        <AppButton variant="ghost" size="sm" @click="navigateTo('/tma/gantt')">
          {{ $t('tma.gantt') }}
        </AppButton>
        <AppButton variant="ghost" size="sm" @click="exportXml">
          {{ $t('tma.export') }}
        </AppButton>
        <AppButton variant="primary" size="sm" @click="showForm = !showForm">
          {{ $t('tma.new') }}
        </AppButton>
      </template>
    </AppPageHeader>

    <AppKpiGrid compact>
      <AppKpiCard
        icon="list_alt"
        tone="gold"
        :loading="pending"
        :value="kpi.total"
        :label="$t('tma.kpi_total')"
      />
      <AppKpiCard
        icon="pending"
        tone="blue"
        :loading="pending"
        :value="kpi.open"
        :label="$t('tma.kpi_open')"
      />
      <AppKpiCard
        icon="check_circle"
        tone="success"
        :loading="pending"
        :value="kpi.resolved"
        :label="$t('tma.kpi_resolved')"
      />
      <AppKpiCard
        v-if="canValidateTma"
        icon="hourglass_empty"
        tone="warn"
        :loading="pending"
        :value="kpi.awaiting"
        :label="$t('tma.kpi_awaiting')"
      />
    </AppKpiGrid>

    <AppCard v-if="showForm" padding="lg" class="mb">
      <DemandForm @submit="onCreate" />
    </AppCard>

    <AppCard padding="lg">
      <AppTable :columns="columns" :rows="rows" :loading="pending" :empty-title="$t('tma.empty')">
        <template #cell-status="{ value }">
          <AppBadge variant="neutral">{{ value }}</AppBadge>
        </template>
        <template #cell-actions="{ row }">
          <AppButton variant="ghost" size="sm" @click="navigateTo(`/tma/${row.id}`)">
            {{ $t('tma.open') }}
          </AppButton>
        </template>
      </AppTable>
    </AppCard>
  </div>
</template>

<script setup lang="ts">
import { countTmaByStatus, countTmaOpen } from '~/composables/useKpiMetrics'

definePageMeta({ layout: 'default' })

const { t } = useI18n()
const { list, create, exportXml, pickId, pickSubject, pickStatus } = useTma()

const { canValidateTma } = usePermissions()

const showForm = ref(false)
const creating = ref(false)

const { data, pending, refresh } = await useAsyncData('tma-demands', () => list())

const kpi = computed(() => {
  const items = data.value ?? []
  return {
    total: items.length,
    open: countTmaOpen(items),
    resolved: countTmaByStatus(items, 'resolue'),
    awaiting: countTmaByStatus(items, 'en_attente_creation')
  }
})

const columns = computed(() => [
  { key: 'title', label: t('tma.col_title') },
  { key: 'status', label: t('tma.col_status') },
  { key: 'actions', label: '' }
])

const rows = computed(() =>
  (data.value ?? []).map((d) => ({
    id: pickId(d),
    title: pickSubject(d),
    status: pickStatus(d)
  }))
)

const onCreate = async (payload: { applicationId: string; subject: string; requiresChefGate: boolean }) => {
  creating.value = true
  try {
    await create(payload)
    showForm.value = false
    await refresh()
  } finally {
    creating.value = false
  }
}
</script>

<style scoped>
.mb { margin-bottom: var(--kore-space-lg); }
</style>
