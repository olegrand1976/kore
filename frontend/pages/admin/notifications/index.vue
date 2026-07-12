<template>
  <div>
    <AppPageHeader :title="$t('notifications.title')" />
    <AppCard padding="lg" class="mb">
      <h3 class="section-title">{{ $t('notifications.rules') }}</h3>
      <AppTable :columns="ruleColumns" :rows="ruleRows" :loading="rulesPending" :empty-title="$t('notifications.rules_empty')" />
    </AppCard>
    <AppCard padding="lg">
      <h3 class="section-title">{{ $t('notifications.journal') }}</h3>
      <AppTable :columns="journalColumns" :rows="journalRows" :loading="journalPending" :empty-title="$t('notifications.journal_empty')" />
    </AppCard>
  </div>
</template>

<script setup lang="ts">
definePageMeta({ layout: 'default', middleware: ['admin'] })

const { t } = useI18n()
const { data: rulesData, pending: rulesPending } = await useFetch('/api/notifications/rules')
const { data: journalData, pending: journalPending } = await useFetch('/api/notifications/journal')

const ruleColumns = computed(() => [
  { key: 'code', label: t('notifications.col_code') },
  { key: 'trigger', label: t('notifications.col_trigger') },
  { key: 'frequency', label: t('notifications.col_frequency') }
])

const journalColumns = computed(() => [
  { key: 'subject', label: t('notifications.col_subject') },
  { key: 'status', label: t('notifications.col_status') }
])

const ruleRows = computed(() => {
  const items = (rulesData.value as any)?.data ?? []
  return items.map((r: any) => ({
    code: r.code ?? r.Code,
    trigger: r.trigger ?? r.Trigger,
    frequency: r.frequency ?? r.Frequency
  }))
})

const journalRows = computed(() => {
  const items = (journalData.value as any)?.data ?? []
  return items.map((m: any) => ({
    subject: m.subject ?? m.Subject,
    status: m.status ?? m.Status
  }))
})
</script>

<style scoped>
.mb { margin-bottom: var(--kore-space-lg); }
.section-title { margin: 0 0 var(--kore-space-md); font-size: var(--kore-text-h3); }
</style>
