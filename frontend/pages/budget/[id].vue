<template>
  <div>
    <AppPageHeader :title="pageTitle" :subtitle="pageSubtitle">
      <template #actions>
        <AppButton variant="ghost" size="sm" @click="navigateTo('/budget')">
          {{ $t('budget.back') }}
        </AppButton>
        <AppButton variant="secondary" size="sm" :disabled="busy" @click="onRecompute">
          {{ $t('budget.recompute') }}
        </AppButton>
      </template>
    </AppPageHeader>

    <p v-if="errorMsg" class="flash flash--error" role="alert">{{ errorMsg }}</p>

    <AppCard v-if="pending" padding="lg"><p class="muted">{{ $t('budget.loading') }}</p></AppCard>

    <AppCard v-else-if="loadError" padding="lg">
      <AppEmptyState icon="error" :title="$t('budget.not_found')" />
    </AppCard>

    <template v-else-if="budget">
      <BudgetContextCard
        class="mb"
        :application="application"
        :budget-type="budgetType"
        :currency="currency"
      />

      <BudgetIntegrationPanel class="mb" />

      <AppCard padding="lg" class="mb">
        <div class="consumption-head">
          <h2 class="section-title">{{ $t('budget.consumption_title') }}</h2>
          <AppBadge :variant="statusBadgeVariant(overallStatus)">{{ statusLabel(overallStatus) }}</AppBadge>
        </div>
        <p class="recompute-hint">{{ $t('budget.recompute_period_hint') }}</p>
        <BudgetTripleGauge
          class="mt"
          :planned-days="plannedDays"
          :consumed-days="consumedDays"
          :remaining-days="remainingDays"
          :planned-u-o="plannedUO"
          :consumed-u-o="consumedUO"
          :remaining-u-o="remainingUO"
          :planned-amount="plannedAmount"
          :consumed-amount="consumedAmount"
          :remaining-amount="remainingAmount"
          :currency="currency"
        />
      </AppCard>

      <AppCard v-if="isManager" padding="lg">
        <h2 class="section-title">{{ $t('budget.manager_section_title') }}</h2>
        <p class="section-help">{{ $t('budget.manager_section_help') }}</p>

        <h3 class="subsection-title">{{ $t('budget.estimate_title') }}</h3>
        <form class="form" @submit.prevent="submitEstimate">
          <label class="field-label" for="est-demand">{{ $t('budget.form_demand') }}</label>
          <select id="est-demand" v-model="estimateForm.demandId" class="field-select" required>
            <option value="" disabled>{{ $t('budget.form_demand_empty') }}</option>
            <option v-for="d in tmaOptions" :key="d.id" :value="d.id">{{ d.label }}</option>
          </select>
          <p class="field-hint">{{ $t('budget.form_demand_help') }}</p>
          <AppButton
            v-if="estimateForm.demandId"
            variant="ghost"
            size="sm"
            type="button"
            @click="navigateTo(`/tma/${estimateForm.demandId}`)"
          >
            {{ $t('budget.view_demand') }}
          </AppButton>
          <div class="form-row">
            <AppButton variant="ghost" size="sm" type="button" :disabled="busy || !estimateForm.demandId" @click="onAiEstimate">
              {{ $t('ai.budget_estimate') }}
            </AppButton>
          </div>
          <p v-if="estimateRationale" class="ai-hint">{{ estimateRationale }}</p>
          <AppInput id="est-days" v-model="estimateForm.effortDays" type="number" step="0.5" :label="$t('budget.form_effort_days')" />
          <AppInput id="est-uo" v-model="estimateForm.effortUO" type="number" step="0.5" :label="$t('budget.form_effort_uo')" />
          <AppButton variant="primary" size="sm" type="submit" :disabled="busy">{{ $t('budget.estimate_submit') }}</AppButton>
        </form>

        <h3 class="subsection-title">{{ $t('budget.quote_title') }}</h3>
        <form class="form" @submit.prevent="submitQuote">
          <label class="field-label" for="quote-demand">{{ $t('budget.form_demand') }}</label>
          <select id="quote-demand" v-model="quoteForm.demandId" class="field-select" required>
            <option value="" disabled>{{ $t('budget.form_demand_empty') }}</option>
            <option v-for="d in tmaOptions" :key="`q-${d.id}`" :value="d.id">{{ d.label }}</option>
          </select>
          <AppButton
            v-if="quoteForm.demandId"
            variant="ghost"
            size="sm"
            type="button"
            @click="navigateTo(`/tma/${quoteForm.demandId}`)"
          >
            {{ $t('budget.view_demand') }}
          </AppButton>
          <AppInput id="quote-amount" v-model="quoteForm.amountEur" type="number" step="0.01" min="0" :label="$t('budget.form_amount_eur')" />
          <AppInput id="quote-days" v-model="quoteForm.effortDays" type="number" step="0.5" :label="$t('budget.form_effort_days')" />
          <AppInput id="quote-uo" v-model="quoteForm.effortUO" type="number" step="0.5" :label="$t('budget.form_effort_uo')" />
          <AppButton variant="secondary" size="sm" type="submit" :disabled="busy">{{ $t('budget.quote_submit') }}</AppButton>
        </form>
      </AppCard>
    </template>
  </div>
</template>

<script setup lang="ts">
import type { BudgetItem } from '~/composables/useBudget'
import type { OrgApplication } from '~/composables/useApplications'

definePageMeta({ layout: 'default' })

const route = useRoute()
const { isManager, fetchSession } = useAuth()
const { extractFetchError } = useApiError()
const { get, recompute, addEstimate, addQuote, tripleValue, pickId } = useBudget()
const { estimateBudgetEffort, suggestBudgetDemands, extractFetchError: aiError } = useAi()
const { list: listTma, pickId: pickTmaId, pickSubject } = useTma()
const { get: getApplication, pickAppLabel, pickAppClient } = useApplications()
const {
  budgetTypeLabel,
  budgetStatus,
  statusLabel,
  statusBadgeVariant,
  currentMonthPeriod,
  worstBudgetStatus,
  budgetPageTitle,
  pickApplicationId,
  eurosToCentimes
} = useBudgetDisplay()

await fetchSession()

const estimateRationale = ref('')

const onAiEstimate = async () => {
  errorMsg.value = ''
  try {
    const res = await estimateBudgetEffort(estimateForm.demandId, id.value)
    estimateForm.effortDays = String(res.effortDays)
    estimateForm.effortUO = String(res.effortUO)
    estimateRationale.value = res.rationale
  } catch (err) {
    errorMsg.value = aiError(err)
  }
}

const id = computed(() => String(route.params.id))
const busy = ref(false)
const errorMsg = ref('')

const loadError = ref('')

const { data, pending, refresh } = await useAsyncData(
  () => `budget-${id.value}`,
  async () => {
    loadError.value = ''
    try {
      const budget = await get(id.value)
      const appId = pickApplicationId(budget as BudgetItem)
      let application: OrgApplication | null = null
      if (appId) {
        try {
          application = await getApplication(appId)
        } catch {
          application = null
        }
      }
      return { budget, application }
    } catch (err) {
      loadError.value = extractFetchError(err)
      return null
    }
  },
  { watch: [id] }
)

const budget = computed(() => data.value?.budget as BudgetItem | undefined)
const application = computed(() => data.value?.application ?? null)
const budgetType = computed(() => budget.value?.type ?? budget.value?.Type ?? '')
const currency = computed(() => budget.value?.currency ?? budget.value?.Currency ?? 'EUR')

const plannedDays = computed(() => tripleValue(budget.value?.planned ?? budget.value?.Planned, 'days'))
const consumedDays = computed(() => tripleValue(budget.value?.consumed ?? budget.value?.Consumed, 'days'))
const remainingDays = computed(() => tripleValue(budget.value?.remaining ?? budget.value?.Remaining, 'days'))
const plannedUO = computed(() => tripleValue(budget.value?.planned ?? budget.value?.Planned, 'uo'))
const consumedUO = computed(() => tripleValue(budget.value?.consumed ?? budget.value?.Consumed, 'uo'))
const remainingUO = computed(() => tripleValue(budget.value?.remaining ?? budget.value?.Remaining, 'uo'))
const plannedAmount = computed(() => tripleValue(budget.value?.planned ?? budget.value?.Planned, 'amount'))
const consumedAmount = computed(() => tripleValue(budget.value?.consumed ?? budget.value?.Consumed, 'amount'))
const remainingAmount = computed(() => tripleValue(budget.value?.remaining ?? budget.value?.Remaining, 'amount'))

const overallStatus = computed(() =>
  worstBudgetStatus(
    budgetStatus(consumedDays.value, plannedDays.value),
    budgetStatus(consumedUO.value, plannedUO.value),
    budgetStatus(consumedAmount.value, plannedAmount.value)
  )
)

const pageTitle = computed(() => budgetPageTitle(pickAppLabel(application.value), pickId(budget.value ?? {}) || id.value))
const pageSubtitle = computed(() => {
  const type = budgetTypeLabel(budgetType.value)
  const client = pickAppClient(application.value)
  return client ? `${type} · ${client}` : type
})

const applicationId = computed(() => pickApplicationId(budget.value ?? {}))
const tmaOptions = ref<Array<{ id: string; label: string }>>([])

const loadTmaOptions = async () => {
  if (!applicationId.value) {
    tmaOptions.value = []
    return
  }
  try {
    const demands = await listTma()
    tmaOptions.value = demands
      .filter((d) => (d.applicationId ?? d.ApplicationID) === applicationId.value)
      .map((d) => ({
        id: pickTmaId(d),
        label: pickSubject(d) || pickTmaId(d).slice(0, 8)
      }))
      .filter((d) => d.id)
  } catch {
    tmaOptions.value = []
  }
}

const ensureTmaOption = (demandId: string, subject: string) => {
  if (!demandId || tmaOptions.value.some((d) => d.id === demandId)) return
  tmaOptions.value.unshift({
    id: demandId,
    label: subject || demandId.slice(0, 8)
  })
}

const applyDemandSuggestion = (demandId: string, subject: string) => {
  ensureTmaOption(demandId, subject)
  if (!estimateForm.demandId) estimateForm.demandId = demandId
  if (!quoteForm.demandId) quoteForm.demandId = demandId
}

onMounted(async () => {
  await loadTmaOptions()
  if (!isManager.value) return
  try {
    const suggestions = await suggestBudgetDemands(id.value)
    const first = suggestions[0]
    if (first?.demandId) {
      applyDemandSuggestion(first.demandId, first.subject)
    }
  } catch {
    /* optional */
  }
})

watch(applicationId, () => {
  loadTmaOptions()
})

const estimateForm = reactive({ demandId: '', effortDays: '1', effortUO: '1' })
const quoteForm = reactive({ demandId: '', amountEur: '10', effortDays: '1', effortUO: '1' })

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

const onRecompute = () => runAction(() => recompute(id.value, currentMonthPeriod()))

const submitEstimate = () =>
  runAction(() =>
    addEstimate(id.value, {
      demandId: estimateForm.demandId,
      effortDays: Number(estimateForm.effortDays),
      effortUO: Number(estimateForm.effortUO)
    })
  )

const submitQuote = () =>
  runAction(() =>
    addQuote(id.value, {
      demandId: quoteForm.demandId,
      amount: eurosToCentimes(Number(quoteForm.amountEur)),
      effortDays: Number(quoteForm.effortDays),
      effortUO: Number(quoteForm.effortUO)
    })
  )
</script>

<style scoped>
.muted { color: var(--kore-text-muted); }
.mb { margin-bottom: var(--kore-space-lg); }
.mt { margin-top: var(--kore-space-lg); }
.section-title { margin: 0 0 var(--kore-space-sm); font-size: var(--kore-text-body); }
.subsection-title { margin: var(--kore-space-lg) 0 var(--kore-space-md); font-size: var(--kore-text-small); font-weight: 600; }
.section-help { margin: 0 0 var(--kore-space-lg); font-size: var(--kore-text-small); color: var(--kore-text-muted); }
.consumption-head { display: flex; flex-wrap: wrap; align-items: center; gap: var(--kore-space-sm); margin-bottom: var(--kore-space-sm); }
.recompute-hint { margin: 0; font-size: var(--kore-text-caption); color: var(--kore-text-muted); }
.form { display: grid; gap: var(--kore-space-md); max-width: var(--kore-form-max); margin-bottom: var(--kore-space-xl); }
.form-row { display: flex; flex-wrap: wrap; gap: var(--kore-space-sm); }
.field-label { font-size: var(--kore-text-small); font-weight: 600; }
.field-select {
  width: 100%;
  max-width: var(--kore-form-max);
  padding: var(--kore-space-sm) var(--kore-space-md);
  border: 1px solid var(--kore-border);
  border-radius: var(--kore-radius-md);
  background: var(--kore-bg);
  color: var(--kore-text);
  font: inherit;
}
.field-hint { margin: calc(-1 * var(--kore-space-sm)) 0 0; font-size: var(--kore-text-caption); color: var(--kore-text-muted); }
.ai-hint { margin: 0; font-size: var(--kore-text-caption); color: var(--kore-text-muted); }
.flash { margin: 0 0 var(--kore-space-md); padding: var(--kore-space-sm) var(--kore-space-md); border-radius: var(--kore-radius-md); font-size: var(--kore-text-small); }
.flash--error { background: color-mix(in srgb, var(--kore-danger) 12%, transparent); color: var(--kore-danger); }
@media (max-width: 640px) {
  .field-select { max-width: none; }
}
</style>
