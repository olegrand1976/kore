<template>
  <div class="integrations-page">
    <AppPageHeader :title="$t('integrations.title')" :subtitle="$t('integrations.subtitle')" />

    <AppCard padding="lg" class="mb">
      <div class="section-head">
        <h2 class="section-title">{{ $t('integrations.connections_title') }}</h2>
        <AppButton variant="secondary" size="sm" :loading="connectPending" @click="connectFec">
          {{ $t('integrations.connect_fec') }}
        </AppButton>
      </div>
      <AppTable
        :columns="connColumns"
        :rows="connRows"
        :loading="connPending"
        :empty-title="$t('integrations.connections_empty')"
      >
        <template #cell-status="{ row }">
          <AppBadge :variant="row.status === 'active' ? 'success' : 'neutral'">{{ row.status }}</AppBadge>
        </template>
        <template #cell-actions="{ row }">
          <AppButton variant="ghost" size="sm" :loading="syncingId === row.rawId" @click="syncConnection(row.rawId)">
            {{ $t('integrations.sync') }}
          </AppButton>
        </template>
      </AppTable>
    </AppCard>

    <AppCard padding="lg" class="mb">
      <div class="section-head">
        <h2 class="section-title">{{ $t('integrations.api_keys_title') }}</h2>
        <AppButton variant="secondary" size="sm" :loading="keyPending" @click="createKey">
          {{ $t('integrations.create_key') }}
        </AppButton>
      </div>
      <p v-if="plainKey" class="plain-key">{{ $t('integrations.plain_key_hint') }} <code>{{ plainKey }}</code></p>
      <AppTable
        :columns="keyColumns"
        :rows="keyRows"
        :loading="keysPending"
        :empty-title="$t('integrations.keys_empty')"
      />
    </AppCard>

    <AppCard padding="lg">
      <h2 class="section-title">{{ $t('integrations.sync_logs_title') }}</h2>
      <AppTable
        :columns="logColumns"
        :rows="logRows"
        :loading="logsPending"
        :empty-title="$t('integrations.logs_empty')"
      >
        <template #cell-status="{ row }">
          <AppBadge :variant="row.status === 'completed' ? 'success' : 'error'">{{ row.status }}</AppBadge>
        </template>
      </AppTable>
    </AppCard>
  </div>
</template>

<script setup lang="ts">
definePageMeta({ layout: 'default', middleware: 'admin' })

const { t } = useI18n()

type Connection = { id: string; type: string; provider: string; status: string }
type ApiKey = { id: string; name: string; keyPrefix: string; createdAt: string }
type SyncLog = { id: string; connectionId: string; status: string; errorMessage?: string; startedAt: string }

const unwrap = <T,>(raw: { data?: T } | T | null): T[] => {
  if (!raw) return []
  if (Array.isArray(raw)) return raw as T[]
  if (typeof raw === 'object' && 'data' in raw && Array.isArray((raw as { data?: T[] }).data)) {
    return (raw as { data: T[] }).data
  }
  return []
}

const { data: connections, pending: connPending, refresh: refreshConn } = await useFetch('/api/integrations/connections')
const { data: keys, pending: keysPending, refresh: refreshKeys } = await useFetch('/api/integrations/api-keys')
const { data: logs, pending: logsPending, refresh: refreshLogs } = await useFetch('/api/integrations/sync-logs')

const connColumns = computed(() => [
  { key: 'provider', label: t('integrations.col_provider') },
  { key: 'type', label: t('integrations.col_type') },
  { key: 'status', label: t('integrations.col_status') },
  { key: 'actions', label: t('integrations.col_actions'), nowrap: true }
])

const connRows = computed(() =>
  unwrap<Connection>(connections.value).map((c) => ({
    rawId: c.id,
    provider: c.provider,
    type: c.type,
    status: c.status
  }))
)

const keyColumns = computed(() => [
  { key: 'name', label: t('integrations.col_key_name') },
  { key: 'prefix', label: t('integrations.col_key_prefix') },
  { key: 'createdAt', label: t('integrations.col_created') }
])

const keyRows = computed(() =>
  unwrap<ApiKey>(keys.value).map((k) => ({
    name: k.name,
    prefix: k.keyPrefix,
    createdAt: new Date(k.createdAt).toLocaleDateString()
  }))
)

const logColumns = computed(() => [
  { key: 'status', label: t('integrations.col_status') },
  { key: 'message', label: t('integrations.col_message') },
  { key: 'startedAt', label: t('integrations.col_started') }
])

const logRows = computed(() =>
  unwrap<SyncLog>(logs.value).map((l) => ({
    status: l.status,
    message: l.errorMessage || '—',
    startedAt: new Date(l.startedAt).toLocaleString()
  }))
)

const connectPending = ref(false)
const keyPending = ref(false)
const plainKey = ref('')
const syncingId = ref('')

const connectFec = async () => {
  connectPending.value = true
  try {
    await $fetch('/api/integrations/connections', {
      method: 'POST',
      body: { type: 'accounting', provider: 'fec', credentialsRef: 'local' }
    })
    await refreshConn()
  } finally {
    connectPending.value = false
  }
}

const syncConnection = async (connId: string) => {
  syncingId.value = connId
  try {
    await $fetch(`/api/integrations/connections/${connId}/sync`, { method: 'POST' })
    await Promise.all([refreshConn(), refreshLogs()])
  } finally {
    syncingId.value = ''
  }
}

const createKey = async () => {
  keyPending.value = true
  plainKey.value = ''
  try {
    const res = await $fetch<{ data?: { plainKey?: string }; plainKey?: string }>('/api/integrations/api-keys', {
      method: 'POST',
      body: { name: `API ${new Date().toISOString().slice(0, 10)}` }
    })
    plainKey.value = res?.plainKey ?? res?.data?.plainKey ?? ''
    await refreshKeys()
  } finally {
    keyPending.value = false
  }
}
</script>

<style scoped>
.mb {
  margin-bottom: var(--kore-space-lg);
}

.section-head {
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  justify-content: space-between;
  gap: var(--kore-space-sm);
  margin-bottom: var(--kore-space-md);
}

.section-title {
  margin: 0;
  font-size: var(--kore-text-h3);
}

.plain-key {
  margin: 0 0 var(--kore-space-md);
  padding: var(--kore-space-sm);
  background: var(--kore-surface-muted);
  border-radius: var(--kore-radius-sm);
  font-size: var(--kore-text-small);
  word-break: break-all;
}

@media (max-width: 640px) {
  .section-head {
    flex-direction: column;
    align-items: stretch;
  }
}
</style>
