<template>
  <div>
    <AppPageHeader :title="$t('missions.title')">
      <template #actions>
        <AppButton v-if="guideRef?.dismissed" variant="ghost" size="sm" type="button" @click="guideRef?.showAgain()">
          {{ $t('guides.show') }}
        </AppButton>
        <AppButton v-if="canCreate" variant="primary" size="sm" to="/missions/nouveau">
          {{ $t('missions.new') }}
        </AppButton>
      </template>
    </AppPageHeader>

    <AppSectionGuide ref="guideRef" guide-key="missions" />

    <AppCard v-if="pending" padding="lg">
      <p class="muted">{{ $t('fiche.loading') }}</p>
    </AppCard>

    <AppCard v-else-if="error" padding="lg">
      <p class="flash flash--error">{{ $t('missions.load_error') }}</p>
    </AppCard>

    <AppCard v-else padding="none">
      <AppTable
        :columns="columns"
        :rows="rows"
        :empty-title="$t('missions.empty')"
      >
        <template #cell-actions="{ row }">
          <AppButton variant="ghost" size="sm" :to="`/missions/${row.id}`">
            {{ $t('fiche.open_mission') }}
          </AppButton>
        </template>
      </AppTable>
    </AppCard>
  </div>
</template>

<script setup lang="ts">
definePageMeta({ layout: 'default' })

const { apiFetch } = useApiFetch()
const { can } = usePermissions()
const guideRef = ref<{ showAgain: () => void; dismissed: boolean } | null>(null)

const { t } = useI18n()
const { formatDate, formatMoney, missionStatusLabel } = useFicheFormat()

const canCreate = computed(() => can('ssii', 'E'))

type MissionRow = {
  id: string
  clientId: string
  clientName: string
  title?: string
  status: string
  startDate: string
  endDate?: string | null
  rateUnit?: string
  tjmAmount: number
  currency: string
}

const { data, pending, error } = await useAsyncData('missions-list', () =>
  apiFetch<{ data?: MissionRow[] }>('/api/ssii/missions')
)

const rateLabel = (mission: MissionRow) => {
  const amount = formatMoney(mission.tjmAmount, mission.currency || 'EUR')
  switch (mission.rateUnit) {
    case 'hourly':
      return `${amount}/h`
    case 'tjm':
    case undefined:
    case '':
      return `${amount}/j`
    default:
      return amount
  }
}

const columns = computed(() => [
  { key: 'title', label: t('missions.col_title') },
  { key: 'clientName', label: t('fiche.col_client') },
  { key: 'status', label: t('fiche.col_status') },
  { key: 'period', label: t('fiche.col_period') },
  { key: 'rate', label: t('missions.col_rate') },
  { key: 'actions', label: t('prestations.col_actions'), nowrap: true }
])

const rows = computed(() =>
  (data.value?.data ?? []).map((mission) => ({
    id: mission.id,
    title: mission.title?.trim() || '—',
    clientName: mission.clientName || '—',
    status: missionStatusLabel(mission.status),
    period: mission.endDate
      ? `${formatDate(mission.startDate)} → ${formatDate(mission.endDate)}`
      : `${formatDate(mission.startDate)} → …`,
    rate: rateLabel(mission)
  }))
)
</script>

<style scoped>
.muted {
  margin: 0;
  color: var(--kore-text-muted);
}

.flash--error {
  color: var(--kore-error);
}
</style>
