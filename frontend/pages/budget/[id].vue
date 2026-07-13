<template>
  <div>
    <AppPageHeader :title="pageTitle">
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

    <template v-else-if="budget">
      <AppCard padding="lg" class="mb">
        <dl class="meta">
          <div><dt>{{ $t('budget.col_type') }}</dt><dd>{{ budget.type ?? budget.Type }}</dd></div>
          <div><dt>{{ $t('budget.col_currency') }}</dt><dd>{{ budget.currency ?? budget.Currency ?? 'EUR' }}</dd></div>
        </dl>
        <BudgetTripleGauge
          class="mt"
          :planned-days="plannedDays"
          :consumed-days="consumedDays"
          :planned-u-o="plannedUO"
          :consumed-u-o="consumedUO"
          :planned-amount="plannedAmount"
          :consumed-amount="consumedAmount"
          :currency="budget.currency ?? budget.Currency ?? 'EUR'"
        />
      </AppCard>

      <AppCard v-if="isManager" padding="lg">
        <h2 class="section-title">{{ $t('budget.estimate_title') }}</h2>
        <form class="form" @submit.prevent="submitEstimate">
          <AppInput id="est-demand" v-model="estimateForm.demandId" :label="$t('budget.form_demand_id')" />
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

        <h2 class="section-title">{{ $t('budget.quote_title') }}</h2>
        <form class="form" @submit.prevent="submitQuote">
          <AppInput id="quote-demand" v-model="quoteForm.demandId" :label="$t('budget.form_demand_id')" />
          <AppInput id="quote-amount" v-model="quoteForm.amount" type="number" :label="$t('budget.form_amount')" />
          <AppInput id="quote-days" v-model="quoteForm.effortDays" type="number" step="0.5" :label="$t('budget.form_effort_days')" />
          <AppInput id="quote-uo" v-model="quoteForm.effortUO" type="number" step="0.5" :label="$t('budget.form_effort_uo')" />
          <AppButton variant="secondary" size="sm" type="submit" :disabled="busy">{{ $t('budget.quote_submit') }}</AppButton>
        </form>
      </AppCard>
    </template>
  </div>
</template>

<script setup lang="ts">
definePageMeta({ layout: 'default' })

const route = useRoute()
const { t } = useI18n()
const { isManager, fetchSession } = useAuth()
const { extractFetchError } = useApiError()
const { get, recompute, addEstimate, addQuote, tripleValue, pickId } = useBudget()
const { estimateBudgetEffort, suggestBudgetDemands, extractFetchError: aiError } = useAi()

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

onMounted(async () => {
  if (!isManager.value) return
  try {
    const suggestions = await suggestBudgetDemands(id.value)
    if (suggestions.length > 0 && !estimateForm.demandId) {
      estimateForm.demandId = suggestions[0].demandId
    }
  } catch {
    /* optional */
  }
})

const id = computed(() => String(route.params.id))
const busy = ref(false)
const errorMsg = ref('')

const { data, pending, refresh } = await useAsyncData(
  () => `budget-${id.value}`,
  () => get(id.value),
  { watch: [id] }
)

const budget = computed(() => data.value as any)
const plannedDays = computed(() => tripleValue(budget.value?.planned ?? budget.value?.Planned, 'days'))
const consumedDays = computed(() => tripleValue(budget.value?.consumed ?? budget.value?.Consumed, 'days'))
const plannedUO = computed(() => tripleValue(budget.value?.planned ?? budget.value?.Planned, 'uo'))
const consumedUO = computed(() => tripleValue(budget.value?.consumed ?? budget.value?.Consumed, 'uo'))
const plannedAmount = computed(() => tripleValue(budget.value?.planned ?? budget.value?.Planned, 'amount'))
const consumedAmount = computed(() => tripleValue(budget.value?.consumed ?? budget.value?.Consumed, 'amount'))
const pageTitle = computed(() => `${t('budget.title')} — ${pickId(budget.value ?? {}).slice(0, 8) || id.value.slice(0, 8)}`)

const estimateForm = reactive({ demandId: '', effortDays: '1', effortUO: '1' })
const quoteForm = reactive({ demandId: '', amount: '1000', effortDays: '1', effortUO: '1' })

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

const onRecompute = () => runAction(() => recompute(id.value))

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
      amount: Number(quoteForm.amount),
      effortDays: Number(quoteForm.effortDays),
      effortUO: Number(quoteForm.effortUO)
    })
  )
</script>

<style scoped>
.meta { display: grid; gap: var(--kore-space-md); margin: 0 0 var(--kore-space-lg); }
.meta div { display: flex; justify-content: space-between; gap: var(--kore-space-sm); }
.meta dt { color: var(--kore-text-muted); }
.muted { color: var(--kore-text-muted); }
.mb { margin-bottom: var(--kore-space-lg); }
.mt { margin-top: var(--kore-space-lg); }
.section-title { margin: 0 0 var(--kore-space-md); font-size: var(--kore-text-body); }
.form { display: grid; gap: var(--kore-space-md); max-width: 420px; margin-bottom: var(--kore-space-xl); }
.form-row { display: flex; flex-wrap: wrap; gap: var(--kore-space-sm); }
.ai-hint { margin: 0; font-size: var(--kore-text-caption); color: var(--kore-text-muted); }
.flash { margin: 0 0 var(--kore-space-md); padding: var(--kore-space-sm) var(--kore-space-md); border-radius: var(--kore-radius-md); font-size: var(--kore-text-small); }
.flash--error { background: color-mix(in srgb, var(--kore-danger) 12%, transparent); color: var(--kore-danger); }
@media (max-width: 640px) {
  .meta div { flex-direction: column; align-items: flex-start; }
}
</style>
