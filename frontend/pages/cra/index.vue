<template>
  <div>
    <AppPageHeader :title="$t('cra.title')">
      <template #actions>
        <AppButton variant="primary" size="sm" :disabled="creating" @click="openCurrentMonth">
          <AppIcon name="add" /> {{ $t('cra.new') }}
        </AppButton>
      </template>
    </AppPageHeader>

    <AppCard v-if="pending" padding="lg">
      <p class="muted">{{ $t('cra.loading') }}</p>
    </AppCard>

    <AppCard v-else-if="rows.length" padding="none" class="cra-table-wrap">
      <AppTable :columns="columns" :rows="rows" row-key="id">
        <template #cell-month="{ value }">
          <span class="cra-month">{{ formatMonth(String(value)) }}</span>
        </template>
        <template #cell-status="{ value }">
          <AppBadge :variant="statusVariant(String(value))">{{ statusLabel(String(value)) }}</AppBadge>
        </template>
        <template #cell-actions="{ row }">
          <AppButton variant="ghost" size="sm" @click="navigateTo(`/cra/${row.id}`)">{{ $t('cra.open') }}</AppButton>
        </template>
      </AppTable>
    </AppCard>

    <AppCard v-else padding="lg">
      <AppEmptyState icon="schedule" :title="$t('cra.empty')" :description="$t('cra.empty_desc')">
        <AppButton variant="primary" size="sm" :disabled="creating" @click="openCurrentMonth">{{ $t('cra.new') }}</AppButton>
      </AppEmptyState>
    </AppCard>

    <p v-if="errorMsg" class="flash flash--error" role="alert">{{ errorMsg }}</p>
  </div>
</template>

<script setup lang="ts">
definePageMeta({ layout: 'default' })

const { t, locale } = useI18n()
const { statusLabel, statusVariant, currentMonthKey } = useCraStatus()

const columns = computed(() => [
  { key: 'month', label: t('cra.col_period') },
  { key: 'status', label: t('cra.col_status') },
  { key: 'actions', label: '' }
])

const creating = ref(false)
const errorMsg = ref('')

const { data, pending, refresh } = await useFetch('/api/cra/timesheets/recent')

const rows = computed(() => {
  const payload = (data.value as any)?.data ?? data.value
  if (!Array.isArray(payload)) return []
  return payload.map((ts: { id: string; month: string; status: string }) => ({
    id: ts.id,
    month: ts.month,
    status: ts.status,
    actions: ''
  }))
})

const formatMonth = (raw: string) => {
  const [y, m] = raw.split('-').map(Number)
  return new Date(y, m - 1, 1).toLocaleDateString(locale.value === 'en' ? 'en-US' : 'fr-FR', {
    month: 'long', year: 'numeric'
  })
}

const openCurrentMonth = async () => {
  creating.value = true
  errorMsg.value = ''
  try {
    const res = await $fetch<any>(`/api/cra/timesheets?month=${currentMonthKey()}`)
    const ts = res?.data ?? res
    if (ts?.id) {
      await navigateTo(`/cra/${ts.id}`)
      return
    }
    await refresh()
  } catch {
    errorMsg.value = t('cra.download_error')
  } finally {
    creating.value = false
  }
}
</script>

<style scoped>
.cra-table-wrap { overflow: hidden; }

.cra-month {
  font-weight: 600;
  color: var(--kore-text);
}

.muted { color: var(--kore-text-muted); }

.flash {
  margin-top: var(--kore-space-md);
  font-size: var(--kore-text-small);
}

.flash--error { color: var(--kore-error); }
</style>
