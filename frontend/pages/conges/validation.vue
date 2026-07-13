<template>
  <div>
    <p v-if="errorMsg" class="flash flash--error" role="alert">{{ errorMsg }}</p>
    <AppCard v-if="!canValidateConges" padding="lg">
      <AppEmptyState icon="lock" :title="$t('conges.validation_forbidden')" />
    </AppCard>
    <AppCard v-else padding="lg">
      <AppTable
        :columns="columns"
        :rows="rows"
        :loading="pending"
        :empty-title="$t('conges.validation_empty')"
      >
        <template #cell-requester="{ value }">
          <span class="requester">{{ value }}</span>
        </template>
        <template #cell-actions="{ row }">
          <div class="actions">
            <AppButton variant="ghost" size="sm" type="button" @click="toggleContext(row.id)">
              {{ $t('ai.manager_context') }}
            </AppButton>
            <AppButton variant="primary" size="sm" :disabled="busyId === row.id" @click="decide(row.id, 'approve')">
              {{ $t('conges.validation_approve') }}
            </AppButton>
            <AppButton variant="ghost" size="sm" :disabled="busyId === row.id" @click="decide(row.id, 'reject')">
              {{ $t('conges.validation_reject') }}
            </AppButton>
          </div>
        </template>
      </AppTable>
      <AppCard v-if="contextRowId && managerContext" padding="md" class="context-panel">
        <AppAiBadge variant="generated" />
        <p class="context-text">{{ managerContext }}</p>
        <p class="context-disclaimer">{{ $t('ai.disclaimer') }}</p>
      </AppCard>
    </AppCard>
  </div>
</template>

<script setup lang="ts">
import type { LeaveRequest } from '~/composables/useLeave'

const { t } = useI18n()
const { fetchSession } = useAuth()
const { canValidateConges } = usePermissions()
const { extractFetchError } = useApiError()
const { list, pending: pendingFn, approve, reject, pickId, pickFrom, pickTo, pickType, pickMotif, pickUserId, formatLeaveUserLogin } = useLeave()
const { fetchMine } = useLeaveTypeConfigs()
const { typeLabel } = useLeaveLabels()
const { fetchManagerContext } = useAi()

await fetchSession()
await fetchMine()

type OrgUser = { id?: string; ID?: string; login?: string; Login?: string }

type ValidationPayload = {
  items: LeaveRequest[]
  logins: Record<string, string>
}

function normalizeUserId(id: string) {
  return id.trim().toLowerCase()
}

function buildLoginMap(users: OrgUser[]) {
  const map: Record<string, string> = {}
  for (const user of users) {
    const id = user.id ?? user.ID
    const login = user.login ?? user.Login
    if (id && login) map[normalizeUserId(id)] = login
  }
  return map
}

async function loadValidationData(): Promise<ValidationPayload> {
  const items = pendingFn(await list())
  if (!canValidateConges.value) {
    return { items, logins: {} }
  }
  try {
    const res = await $fetch<{ data?: OrgUser[] }>('/api/org/users')
    return { items, logins: buildLoginMap(res?.data ?? []) }
  } catch {
    return { items, logins: {} }
  }
}

const { data, pending, refresh } = await useAsyncData('leave-validation', loadValidationData)

const resolveRequester = (item: LeaveRequest, logins: Record<string, string>) => {
  const userId = normalizeUserId(pickUserId(item))
  if (!userId) return '—'
  const login = logins[userId]
  return login ? formatLeaveUserLogin(login) : userId.slice(0, 8)
}

const contextRowId = ref('')
const managerContext = ref('')

const toggleContext = async (leaveId: string) => {
  if (contextRowId.value === leaveId) {
    contextRowId.value = ''
    managerContext.value = ''
    return
  }
  contextRowId.value = leaveId
  managerContext.value = ''
  try {
    const res = await fetchManagerContext(leaveId)
    managerContext.value = res.context
  } catch {
    managerContext.value = ''
  }
}

const errorMsg = ref('')
const busyId = ref('')

const columns = computed(() => [
  { key: 'requester', label: t('conges.col_requester') },
  { key: 'type', label: t('conges.col_type') },
  { key: 'from', label: t('conges.from') },
  { key: 'to', label: t('conges.to') },
  { key: 'motif', label: t('conges.motif') },
  { key: 'actions', label: '' }
])

const rows = computed(() => {
  const payload = data.value
  const items = payload?.items ?? []
  const logins = payload?.logins ?? {}
  return items.map((item) => ({
    id: pickId(item),
    requester: resolveRequester(item, logins),
    type: typeLabel(pickType(item)),
    from: pickFrom(item),
    to: pickTo(item),
    motif: pickMotif(item) || '—'
  }))
})

const decide = async (id: string, action: 'approve' | 'reject') => {
  busyId.value = id
  errorMsg.value = ''
  try {
    if (action === 'approve') await approve(id)
    else await reject(id)
    await refresh()
  } catch (err) {
    errorMsg.value = extractFetchError(err)
  } finally {
    busyId.value = ''
  }
}
</script>

<style scoped>
.requester {
  font-weight: 500;
  color: var(--kore-text);
  white-space: nowrap;
}

.actions {
  display: flex;
  flex-wrap: wrap;
  gap: var(--kore-space-xs);
}

.context-panel {
  margin-top: var(--kore-space-md);
  display: grid;
  gap: var(--kore-space-sm);
}

.context-text {
  margin: 0;
  font-size: var(--kore-text-small);
  line-height: 1.5;
}

.context-disclaimer {
  margin: 0;
  font-size: var(--kore-text-caption);
  color: var(--kore-text-muted);
}
.flash { margin: 0 0 var(--kore-space-md); padding: var(--kore-space-sm) var(--kore-space-md); border-radius: var(--kore-radius-md); font-size: var(--kore-text-small); }
.flash--error { background: color-mix(in srgb, var(--kore-danger) 12%, transparent); color: var(--kore-danger); }
</style>
