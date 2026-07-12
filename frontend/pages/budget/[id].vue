<template>
  <div>
    <AppPageHeader :title="pageTitle" />
    <AppCard v-if="pending" padding="lg"><p class="muted">{{ $t('budget.loading') }}</p></AppCard>
    <AppCard v-else-if="budget" padding="lg">
      <dl class="meta">
        <div><dt>{{ $t('budget.col_type') }}</dt><dd>{{ budget.type ?? budget.Type }}</dd></div>
        <div><dt>{{ $t('budget.col_planned') }}</dt><dd>{{ plannedDays }}</dd></div>
        <div><dt>{{ $t('budget.col_consumed') }}</dt><dd>{{ consumedDays }}</dd></div>
      </dl>
    </AppCard>
  </div>
</template>

<script setup lang="ts">
definePageMeta({ layout: 'default' })

const route = useRoute()
const { t } = useI18n()
const id = computed(() => String(route.params.id))

const { data, pending } = await useFetch(() => `/api/budget/budgets/${id.value}`)
const budget = computed(() => (data.value as any)?.data ?? data.value)
const plannedDays = computed(() => budget.value?.planned?.days ?? budget.value?.Planned?.Days ?? '-')
const consumedDays = computed(() => budget.value?.consumed?.days ?? budget.value?.Consumed?.Days ?? '-')
const pageTitle = computed(() => `${t('budget.title')} — ${id.value.slice(0, 8)}`)
</script>

<style scoped>
.meta { display: grid; gap: var(--kore-space-md); margin: 0; }
.meta div { display: flex; justify-content: space-between; }
.meta dt { color: var(--kore-text-muted); }
.muted { color: var(--kore-text-muted); }
</style>
