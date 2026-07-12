<template>
  <div>
    <AppPageHeader :title="$t('users.title')" :subtitle="$t('users.subtitle')" />

    <AppCard v-if="pending" padding="lg">
      <p class="muted">{{ $t('users.loading') }}</p>
    </AppCard>

    <AppCard v-else-if="forbidden" padding="lg">
      <AppEmptyState icon="lock" :title="$t('users.forbidden')" />
    </AppCard>

    <AppCard v-else-if="rows.length" padding="none" class="users-table-wrap">
      <AppTable :columns="columns" :rows="rows" row-key="id">
        <template #cell-profil="{ value }">
          <AppBadge variant="default">{{ value }}</AppBadge>
        </template>
        <template #cell-active="{ value }">
          <AppBadge :variant="value ? 'success' : 'default'">{{ value ? $t('users.active') : $t('users.inactive') }}</AppBadge>
        </template>
      </AppTable>
    </AppCard>

    <AppCard v-else padding="lg">
      <AppEmptyState icon="group" :title="$t('users.empty')" />
    </AppCard>
  </div>
</template>

<script setup lang="ts">
definePageMeta({ layout: 'default', middleware: 'admin' })

const { t } = useI18n()

const columns = computed(() => [
  { key: 'login', label: t('users.login') },
  { key: 'profil', label: t('users.profile') },
  { key: 'active', label: t('users.status') }
])

const { data, pending, error } = await useFetch('/api/org/users')
const forbidden = computed(() => (error.value as any)?.statusCode === 403)
const rows = computed(() => {
  if (forbidden.value) return []
  const payload = (data.value as any)?.data ?? data.value
  return Array.isArray(payload) ? payload : []
})
</script>

<style scoped>
.users-table-wrap { overflow: hidden; }
.muted { color: var(--kore-text-muted); }
</style>
