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

    <AppCard padding="lg" class="mb">
      <div class="section-head">
        <h2 class="section-title">{{ $t('integrations.taiga_mappings_title') }}</h2>
        <AppButton variant="secondary" size="sm" @click="toggleMappingForm">
          {{ $t('integrations.taiga_mappings_add') }}
        </AppButton>
      </div>
      <p class="section-hint">{{ $t('integrations.taiga_mappings_subtitle') }}</p>
      <form v-if="showMappingForm" class="mapping-form" @submit.prevent="saveMapping">
        <AppInput id="taiga-user-id" v-model="mappingForm.taigaUserId" type="number" min="1" required>
          <template #label>{{ $t('integrations.taiga_form_taiga_id') }}</template>
        </AppInput>
        <AppInput id="taiga-username" v-model="mappingForm.taigaUsername" required>
          <template #label>{{ $t('integrations.taiga_form_taiga_username') }}</template>
        </AppInput>
        <label class="field-label" for="kore-user-id">{{ $t('integrations.taiga_form_kore_user') }}</label>
        <select id="kore-user-id" v-model="mappingForm.koreUserId" class="field-select" required>
          <option value="">{{ $t('integrations.taiga_form_kore_user_none') }}</option>
          <option v-for="u in orgUsers" :key="u.id" :value="u.id">{{ u.label }}</option>
        </select>
        <label class="field-label" for="match-method">{{ $t('integrations.taiga_form_match') }}</label>
        <select id="match-method" v-model="mappingForm.matchMethod" class="field-select" required>
          <option value="email">{{ $t('integrations.taiga_form_match_email') }}</option>
          <option value="manual">{{ $t('integrations.taiga_form_match_manual') }}</option>
        </select>
        <AppButton type="submit" variant="primary" size="sm" :loading="mappingPending">
          {{ $t('integrations.taiga_mappings_save') }}
        </AppButton>
        <p v-if="mappingError" class="mapping-error">{{ mappingError }}</p>
        <p v-if="mappingSuccess" class="mapping-success">{{ $t('integrations.taiga_mappings_saved') }}</p>
      </form>
      <AppTable
        :columns="mappingColumns"
        :rows="mappingRows"
        :loading="mappingsPending"
        :empty-title="$t('integrations.taiga_mappings_empty')"
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

const { apiFetch } = useApiFetch()
const { t } = useI18n()

type Connection = { id: string; type: string; provider: string; status: string }
type ApiKey = { id: string; name: string; keyPrefix: string; createdAt: string }
type SyncLog = { id: string; connectionId: string; status: string; errorMessage?: string; startedAt: string }
type UserMapping = {
  id?: string
  externalUserId?: string
  externalUsername?: string
  koreUserId?: string
  matchMethod?: string
}
type OrgUser = { id?: string; ID?: string; login?: string; Login?: string; email?: string; Email?: string }

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
const { data: mappings, pending: mappingsPending, refresh: refreshMappings } = await useFetch('/api/integrations/taiga/user-mappings')
const { data: orgUsersRaw } = await useFetch<{ data?: OrgUser[] }>('/api/org/users')

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

const mappingColumns = computed(() => [
  { key: 'taigaUser', label: t('integrations.taiga_col_taiga_user') },
  { key: 'taigaId', label: t('integrations.taiga_col_taiga_id') },
  { key: 'koreUser', label: t('integrations.taiga_col_kore_user') },
  { key: 'matchMethod', label: t('integrations.taiga_col_match') }
])

const orgUsers = computed(() =>
  unwrap<OrgUser>(orgUsersRaw.value).map((u) => {
    const id = u.id ?? u.ID ?? ''
    const login = u.login ?? u.Login ?? ''
    const email = u.email ?? u.Email ?? ''
    return { id, label: email ? `${login} (${email})` : login || id }
  }).filter((u) => u.id)
)

const userLabelById = computed(() => {
  const map = new Map<string, string>()
  for (const u of orgUsers.value) {
    map.set(u.id, u.label)
  }
  return map
})

const mappingRows = computed(() =>
  unwrap<UserMapping>(mappings.value).map((m) => ({
    taigaUser: m.externalUsername ?? '—',
    taigaId: m.externalUserId ?? '—',
    koreUser: userLabelById.value.get(m.koreUserId ?? '') ?? m.koreUserId ?? '—',
    matchMethod: m.matchMethod ?? '—'
  }))
)

const connectPending = ref(false)
const keyPending = ref(false)
const plainKey = ref('')
const syncingId = ref('')
const showMappingForm = ref(false)
const mappingPending = ref(false)
const mappingError = ref('')
const mappingSuccess = ref(false)
const mappingForm = reactive({
  taigaUserId: '',
  taigaUsername: '',
  koreUserId: '',
  matchMethod: 'email'
})

const connectFec = async () => {
  connectPending.value = true
  try {
    await apiFetch('/api/integrations/connections', {
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
    await apiFetch(`/api/integrations/connections/${connId}/sync`, { method: 'POST' })
    await Promise.all([refreshConn(), refreshLogs()])
  } finally {
    syncingId.value = ''
  }
}

const createKey = async () => {
  keyPending.value = true
  plainKey.value = ''
  try {
    const res = await apiFetch<{ data?: { plainKey?: string }; plainKey?: string }>('/api/integrations/api-keys', {
      method: 'POST',
      body: { name: `API ${new Date().toISOString().slice(0, 10)}` }
    })
    plainKey.value = res?.plainKey ?? res?.data?.plainKey ?? ''
    await refreshKeys()
  } finally {
    keyPending.value = false
  }
}

const toggleMappingForm = () => {
  showMappingForm.value = !showMappingForm.value
  mappingError.value = ''
  mappingSuccess.value = false
}

const saveMapping = async () => {
  mappingPending.value = true
  mappingError.value = ''
  mappingSuccess.value = false
  const taigaUserId = Number(mappingForm.taigaUserId)
  if (!Number.isInteger(taigaUserId) || taigaUserId <= 0) {
    mappingError.value = t('integrations.taiga_form_taiga_id_invalid')
    mappingPending.value = false
    return
  }
  try {
    await apiFetch('/api/integrations/taiga/user-mappings', {
      method: 'POST',
      body: {
        taigaUserId,
        taigaUsername: mappingForm.taigaUsername.trim(),
        koreUserId: mappingForm.koreUserId,
        matchMethod: mappingForm.matchMethod
      }
    })
    mappingSuccess.value = true
    mappingForm.taigaUserId = ''
    mappingForm.taigaUsername = ''
    mappingForm.koreUserId = ''
    mappingForm.matchMethod = 'email'
    await refreshMappings()
  } catch (err: unknown) {
    const status = (err as { statusCode?: number; status?: number; response?: { status?: number } })?.statusCode
      ?? (err as { status?: number })?.status
      ?? (err as { response?: { status?: number } })?.response?.status
    mappingError.value = status === 409
      ? t('integrations.taiga_mappings_conflict')
      : t('integrations.taiga_mappings_error')
  } finally {
    mappingPending.value = false
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

.section-hint {
  margin: 0 0 var(--kore-space-md);
  color: var(--kore-text-muted);
  font-size: var(--kore-text-small);
}

.mapping-form {
  display: grid;
  gap: var(--kore-space-sm);
  margin-bottom: var(--kore-space-md);
  max-width: 32rem;
}

.field-label {
  font-size: var(--kore-text-small);
  color: var(--kore-text-muted);
}

.field-select {
  width: 100%;
  padding: var(--kore-space-sm);
  border: 1px solid var(--kore-border);
  border-radius: var(--kore-radius-sm);
  background: var(--kore-surface);
  color: var(--kore-text);
}

.mapping-error {
  margin: 0;
  color: var(--kore-danger);
  font-size: var(--kore-text-small);
}

.mapping-success {
  margin: 0;
  color: var(--kore-success);
  font-size: var(--kore-text-small);
}

@media (max-width: 640px) {
  .section-head {
    flex-direction: column;
    align-items: stretch;
  }
}
</style>
