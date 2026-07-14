<template>
  <div>
    <AppPageHeader :title="$t('missions.title')">
      <template #actions>
        <AppButton variant="primary" size="sm" to="/missions/nouveau">
          {{ $t('missions.new') }}
        </AppButton>
      </template>
    </AppPageHeader>

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

const { t } = useI18n()
const { formatDate, formatMoney, missionStatusLabel } = useFicheFormat()

type MissionRow = {
  id: string
  clientId: string
  clientName: string
  status: string
  startDate: string
  endDate?: string | null
  tjmAmount: number
  currency: string
}

const { data, pending, error } = await useAsyncData('missions-list', () =>
  $fetch<{ data?: MissionRow[] }>('/api/ssii/missions')
)

const columns = computed(() => [
  { key: 'clientName', label: t('fiche.col_client') },
  { key: 'status', label: t('fiche.col_status') },
  { key: 'period', label: t('fiche.col_period') },
  { key: 'tjm', label: t('fiche.col_tjm') },
  { key: 'actions', label: t('prestations.col_actions'), nowrap: true }
])

const rows = computed(() =>
  (data.value?.data ?? []).map((mission) => ({
    id: mission.id,
    clientName: mission.clientName || '—',
    status: missionStatusLabel(mission.status),
    period: mission.endDate
      ? `${formatDate(mission.startDate)} → ${formatDate(mission.endDate)}`
      : `${formatDate(mission.startDate)} → …`,
    tjm: formatMoney(mission.tjmAmount, mission.currency || 'EUR')
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
