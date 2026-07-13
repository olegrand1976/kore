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

const userLoginById = ref<Record<string, string>>({})

if (canValidateConges.value) {
  try {
    const res = await $fetch<{ data?: OrgUser[] }>('/api/org/users')
    const map: Record<string, string> = {}
    for (const user of res?.data ?? []) {
      const id = user.id ?? user.ID
      const login = user.login ?? user.Login
      if (id && login) map[id] = login
    }
    userLoginById.value = map
  } catch {
    userLoginById.value = {}
  }
}

const resolveRequester = (item: Parameters<typeof pickUserId>[0]) => {
  const userId = pickUserId(item)
  if (!userId) return '—'
  const login = userLoginById.value[userId]
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

const { data, pending, refresh } = await useAsyncData('leave-validation', async () => {
  const items = await list()
  return pendingFn(items)
})

const busyId = ref('')

const columns = computed(() => [
  { key: 'requester', label: t('conges.col_requester') },
  { key: 'type', label: t('conges.col_type') },
  { key: 'from', label: t('conges.from') },
  { key: 'to', label: t('conges.to') },
  { key: 'motif', label: t('conges.motif') },
  { key: 'actions', label: '' }
])

const rows = computed(() =>
  (data.value ?? []).map((item) => ({
    id: pickId(item),
    requester: resolveRequester(item),
    type: typeLabel(pickType(item)),
    from: pickFrom(item),
    to: pickTo(item),
    motif: pickMotif(item) || '—'
  }))
)

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
