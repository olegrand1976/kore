<template>
  <div>
    <AppPageHeader :title="pageTitle">
      <template #actions>
        <AppButton variant="ghost" size="sm" @click="navigateTo('/tma')">
          {{ $t('tma.back') }}
        </AppButton>
      </template>
    </AppPageHeader>

    <p v-if="errorMsg" class="flash flash--error" role="alert">{{ errorMsg }}</p>

    <AppCard v-if="pending" padding="lg"><p class="muted">{{ $t('tma.loading') }}</p></AppCard>

    <template v-else-if="demand">
      <AppCard padding="lg" class="mb">
        <dl class="meta">
          <div><dt>{{ $t('tma.col_status') }}</dt><dd><AppBadge variant="neutral">{{ status }}</AppBadge></dd></div>
          <div v-if="workflowState"><dt>{{ $t('tma.workflow_state') }}</dt><dd>{{ workflowState }}</dd></div>
        </dl>
        <WorkflowActions
          :status="status"
          :actions="workflowActions"
          :can-validate-tma="can('tma', 'V')"
          :assignee-id="assigneeId"
          :requires-chef-gate="requiresChefGate"
          :busy="busy"
          :users="teamUsers"
          @validate-creation="onValidateCreation"
          @assign="onAssign"
          @take-over="onTakeOver"
          @resolve="onResolve"
          @reopen="onReopen"
        />
      </AppCard>

      <RequestAttachmentsPanel
        resource="tma"
        :resource-id="id"
        :can-upload="can('tma', 'E')"
      />

      <AppCard padding="lg">
        <h2 class="section-title">{{ $t('tma.analysis_title') }}</h2>
        <AnalysisEditor
          :analysis="analysis"
          :disabled="busy"
          :demand-id="id"
          :subject="pickSubject(demand ?? {})"
          :application-id="String(demand?.applicationId ?? demand?.ApplicationID ?? '')"
          @save="onSaveAnalysis"
        />
      </AppCard>
    </template>
  </div>
</template>

<script setup lang="ts">
definePageMeta({ layout: 'default' })

const route = useRoute()
const { t } = useI18n()
const { fetchSession } = useAuth()
const { can } = usePermissions()
const { extractFetchError } = useApiError()
const {
  get,
  getAnalysis,
  validateCreation,
  assign,
  takeOver,
  resolve,
  reopen,
  saveAnalysis,
  pickSubject,
  pickStatus,
  pickWorkflowId
} = useTma()
const { getInstance, availableActions, pickState } = useWorkflow()

await fetchSession()

const id = computed(() => String(route.params.id))
const busy = ref(false)
const errorMsg = ref('')
const workflowState = ref('')
const workflowActions = ref<string[]>([])
const teamUsers = ref<{ id: string; label: string }[]>([])

const analysis = reactive({
  functional: '',
  technical: '',
  risks: '',
  testScenario: ''
})

const loadDetail = async () => {
  const d = await get(id.value)
  const wfId = pickWorkflowId(d)
  workflowState.value = ''
  workflowActions.value = []
  if (wfId) {
    try {
      const inst = await getInstance(wfId)
      workflowState.value = pickState(inst)
      workflowActions.value = await availableActions(wfId)
    } catch {
      workflowState.value = ''
      workflowActions.value = []
    }
  }
  try {
    const dossier = await getAnalysis(id.value)
    analysis.functional = dossier.functional ?? dossier.Functional ?? ''
    analysis.technical = dossier.technical ?? dossier.Technical ?? ''
    analysis.risks = dossier.risks ?? dossier.Risks ?? ''
    analysis.testScenario = dossier.testScenario ?? dossier.TestScenario ?? ''
  } catch {
    analysis.functional = ''
    analysis.technical = ''
    analysis.risks = ''
    analysis.testScenario = ''
  }
  return d
}

const { data: demand, pending, refresh } = await useAsyncData(
  () => `tma-${id.value}`,
  () => loadDetail(),
  { watch: [id] }
)

if (can('tma', 'V')) {
  try {
    const res = await $fetch<{ data?: Array<{ id?: string; ID?: string; login?: string; Login?: string }> }>('/api/org/users')
    const list = res?.data ?? []
    teamUsers.value = list.map((u) => ({
      id: u.id ?? u.ID ?? '',
      label: u.login ?? u.Login ?? (u.id ?? u.ID ?? '')
    })).filter((u) => u.id)
  } catch {
    teamUsers.value = []
  }
}

const status = computed(() => pickStatus(demand.value ?? {}))
const assigneeId = computed(() => demand.value?.assigneeId ?? demand.value?.AssigneeID)
const requiresChefGate = computed(() => demand.value?.requiresChefGate ?? demand.value?.RequiresChefGate ?? false)
const pageTitle = computed(() => pickSubject(demand.value ?? {}) || t('tma.detail_title'))

const runAction = async (fn: () => Promise<unknown>) => {
  errorMsg.value = ''
  busy.value = true
  try {
    await fn()
    await refresh()
  } catch (err) {
    errorMsg.value = extractFetchError(err)
  } finally {
    busy.value = false
  }
}

const onValidateCreation = () => runAction(() => validateCreation(id.value))
const onAssign = (assigneeId: string) => runAction(() => assign(id.value, assigneeId))
const onTakeOver = () => runAction(() => takeOver(id.value))
const onResolve = () => runAction(() => resolve(id.value))
const onReopen = (reason: string) => runAction(() => reopen(id.value, reason))
const onSaveAnalysis = (payload: typeof analysis) =>
  runAction(async () => {
    await saveAnalysis(id.value, payload)
    Object.assign(analysis, payload)
  })
</script>

<style scoped>
.meta { display: grid; gap: var(--kore-space-md); margin: 0 0 var(--kore-space-lg); }
.meta div { display: flex; justify-content: space-between; gap: var(--kore-space-sm); }
.meta dt { color: var(--kore-text-muted); }
.muted { color: var(--kore-text-muted); }
.mb { margin-bottom: var(--kore-space-lg); }
.mb { margin-bottom: var(--kore-space-lg); }
.section-title { margin: 0 0 var(--kore-space-md); font-size: var(--kore-text-body); }
.flash { margin: 0 0 var(--kore-space-md); padding: var(--kore-space-sm) var(--kore-space-md); border-radius: var(--kore-radius-md); font-size: var(--kore-text-small); }
.flash--error { background: color-mix(in srgb, var(--kore-danger) 12%, transparent); color: var(--kore-danger); }
@media (max-width: 640px) {
  .meta div { flex-direction: column; align-items: flex-start; }
}
</style>
