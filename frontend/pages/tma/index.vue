<template>
  <div>
    <AppPageHeader :title="$t('tma.title')" />
    <AppCard padding="lg">
      <AppTable
        :columns="columns"
        :rows="rows"
        :loading="pending"
        :empty-title="$t('tma.empty')"
      />
    </AppCard>
  </div>
</template>

<script setup lang="ts">
definePageMeta({ layout: 'default' })

const { t } = useI18n()
const { data, pending } = await useFetch('/api/tma/demands')

const items = computed(() => (data.value as any)?.data ?? [])

const columns = computed(() => [
  { key: 'title', label: t('tma.col_title') },
  { key: 'status', label: t('tma.col_status') },
  { key: 'priority', label: t('tma.col_priority') }
])

const rows = computed(() =>
  items.value.map((d: any) => ({
    title: d.title ?? d.Title ?? d.subject ?? '-',
    status: d.status ?? d.Status ?? '-',
    priority: d.priority ?? d.Priority ?? '-'
  }))
)
</script>
