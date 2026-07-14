<template>
  <div class="dashboard">
    <section class="welcome-banner">
      <div>
        <p class="welcome-banner__eyebrow">{{ $t('dashboard.banner') }}</p>
        <h2 class="welcome-banner__title">{{ $t('dashboard.welcome') }}</h2>
      </div>
      <AppIcon name="insights" class="welcome-banner__icon" />
    </section>

    <AppCard v-if="briefingText" padding="lg" class="dashboard__briefing">
      <div class="dashboard__briefing-header">
        <h2 class="dashboard__briefing-title">{{ $t('ai.briefing_title') }}</h2>
        <AppAiBadge variant="generated" />
      </div>
      <p class="dashboard__briefing-text">{{ briefingText }}</p>
    </AppCard>

    <AppKpiGrid class="dashboard__kpis">
      <AppKpiCard
        icon="extension"
        tone="gold"
        :value="activeModulesCount"
        :label="$t('dashboard.modules_active')"
        clickable
        @click="showModulesPanel = true"
      />
      <AppKpiCard
        v-if="canValidateConges && showModule('conges')"
        icon="pending_actions"
        tone="warn"
        :loading="statsPending"
        :error="statErrors.conges"
        :value="stats.pendingValidations"
        :label="$t('dashboard.pending_validations')"
        to="/conges/validation"
      />
      <AppKpiCard
        v-if="showModule('cra') && canReadReporting"
        icon="payments"
        tone="gold"
        :loading="statsPending"
        :error="statErrors.billing"
        :value="billingAmountDisplay"
        :label="$t('dashboard.billing_amount')"
        to="/reporting/facturation"
      />
      <AppKpiCard
        v-if="showModule('cra') && canReadReporting"
        icon="timeline"
        tone="blue"
        :loading="statsPending"
        :value="Math.round(stats.billableHoursMonth)"
        :label="$t('dashboard.billable_hours')"
        to="/cra/planning"
      />
      <AppKpiCard
        v-if="showModule('cra') && canReadReporting"
        icon="dashboard"
        tone="success"
        :loading="statsPending"
        :value="stats.billingInvoiceCount"
        :label="$t('dashboard.invoice_count')"
        to="/reporting/dashboards/cra"
      />
      <AppKpiCard
        v-if="showModule('cra') && stats.craRequired"
        icon="schedule"
        tone="blue"
        :loading="statsPending"
        :error="statErrors.cra"
        :value="craCurrentDisplay"
        :label="$t('dashboard.cra_current')"
        :hint="currentMonthLabel"
        to="/cra"
      />
      <AppKpiCard
        v-if="showModule('cra') && stats.craRequired && stats.craPrefillRatio != null"
        icon="auto_fix_high"
        :tone="stats.craPrefillLow ? 'warn' : 'success'"
        :loading="statsPending"
        :error="statErrors.cra"
        :value="`${stats.craPrefillRatio}%`"
        :label="$t('dashboard.cra_prefill_ratio')"
        :hint="stats.craPrefillLow ? $t('dashboard.cra_prefill_low') : $t('dashboard.cra_prefill_ok')"
        to="/cra"
      />
      <AppKpiCard
        v-if="showModule('conges') && !canValidateConges"
        icon="event"
        tone="warn"
        :loading="statsPending"
        :error="statErrors.conges"
        :value="stats.leavePending"
        :label="$t('dashboard.leave_pending')"
        to="/conges"
      />
      <AppKpiCard
        v-if="showModule('tma')"
        icon="support_agent"
        tone="blue"
        :loading="statsPending"
        :error="statErrors.tma"
        :value="stats.tmaOpen"
        :label="$t('dashboard.tma_open')"
        :hint="tmaTotalHint"
        to="/tma"
      />
      <AppKpiCard
        v-if="showModule('budget')"
        icon="account_balance"
        :tone="stats.budgetOverrun > 0 ? 'warn' : 'success'"
        :loading="statsPending"
        :error="statErrors.budget"
        :value="budgetConsumptionDisplay"
        :label="$t('dashboard.budget_consumption')"
        :hint="budgetOverrunHint"
        to="/budget"
      />
      <AppCard v-if="showModule('cra') && stats.craRequired && stats.craAlert" padding="lg" class="dashboard__cra-alert">
        <p class="dashboard__cra-alert-text">{{ $t('dashboard.cra_alert') }}</p>
        <AppButton variant="primary" size="sm" to="/cra">{{ $t('dashboard.quick_cra') }}</AppButton>
      </AppCard>

      <AppCard v-if="showModule('cra')" padding="lg" hoverable class="kpi-card kpi-card--action">
        <div class="feature-card__icon"><AppIcon name="edit_calendar" /></div>
        <p class="kpi-card__label">{{ $t('nav.cra') }}</p>
        <AppButton variant="primary" size="sm" to="/cra">{{ $t('dashboard.quick_cra') }}</AppButton>
      </AppCard>
    </AppKpiGrid>

    <DashboardCharts
      :charts="charts"
      :errors="statErrors"
      :loading="statsPending"
      :show-module="showModule"
    />

    <DashboardModulesPanel
      :open="showModulesPanel"
      :active-modules="modules"
      :subscription-status="status"
      :seats="seats"
      @close="showModulesPanel = false"
    />
  </div>
</template>

<script setup lang="ts">
import type { DashboardStatErrors } from '~/composables/useDashboardStats'

definePageMeta({ layout: 'default' })

const { t, locale } = useI18n()
const { modules, status, seats, loaded, fetchEntitlements } = useEntitlements()
const { fetchSession } = useAuth()
const { load, craCurrentLabel, showModule, currentMonthKey, canValidateConges, emptyStats, emptyCharts } =
  useDashboardStats()
const { fetchBriefing } = useAi()
const { canReadReporting } = usePermissions()

await Promise.all([fetchEntitlements(), fetchSession()])

const showModulesPanel = ref(false)

const { data: dashboardData, pending: statsPending } = await useAsyncData('dashboard-stats', () => load())

const briefingText = ref('')
watch(
  () => dashboardData.value?.stats,
  async (s) => {
    if (!s) return
    try {
      const briefing = await fetchBriefing({
        craStatus: s.craCurrentStatus ?? '',
        leavePending: s.leavePending,
        tmaOpen: s.tmaOpen,
        budgetConsumption: s.budgetConsumptionPct,
        budgetOverrun: s.budgetOverrun,
        pendingValidations: s.pendingValidations
      })
      briefingText.value = briefing.text
    } catch {
      briefingText.value = ''
    }
  },
  { immediate: true }
)

const stats = computed(() => dashboardData.value?.stats ?? emptyStats())
const charts = computed(() => dashboardData.value?.charts ?? emptyCharts())
const statErrors = computed<DashboardStatErrors>(() => dashboardData.value?.errors ?? {})

const activeModulesCount = computed(() => {
  if (!loaded.value) return '—'
  if (modules.value.length === 0) return 0
  return modules.value.length
})

const craCurrentDisplay = computed(() => craCurrentLabel(stats.value.craCurrentStatus))

const currentMonthLabel = computed(() => {
  const [y, m] = currentMonthKey().split('-').map(Number)
  return new Date(y, m - 1, 1).toLocaleDateString(locale.value === 'en' ? 'en-US' : 'fr-FR', {
    month: 'long',
    year: 'numeric'
  })
})

const tmaTotalHint = computed(() =>
  stats.value.tmaTotal > 0 ? t('dashboard.tma_total_hint', { n: stats.value.tmaTotal }) : undefined
)

const budgetOverrunHint = computed(() =>
  stats.value.budgetOverrun > 0 ? t('dashboard.budget_overrun_hint', { n: stats.value.budgetOverrun }) : undefined
)

const budgetConsumptionDisplay = computed(() => `${stats.value.budgetConsumptionPct}%`)

const billingAmountDisplay = computed(() => {
  const cents = stats.value.billingAmountCents
  if (!cents) return '—'
  return new Intl.NumberFormat(locale.value === 'en' ? 'en-US' : 'fr-FR', {
    style: 'currency',
    currency: 'EUR'
  }).format(cents / 100)
})
</script>

<style scoped>
.dashboard {
  width: 100%;
}

.dashboard__kpis {
  grid-template-columns: repeat(auto-fill, minmax(180px, 1fr));
}

@media (min-width: 1200px) {
  .dashboard__kpis {
    grid-template-columns: repeat(6, minmax(0, 1fr));
  }
}

.dashboard__briefing {
  margin-bottom: var(--kore-space-xl);
}

.dashboard__briefing-header {
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  gap: var(--kore-space-sm);
  margin-bottom: var(--kore-space-sm);
}

.dashboard__briefing-title {
  margin: 0;
  font-size: var(--kore-text-body);
  font-weight: 600;
}

.dashboard__briefing-text {
  margin: 0;
  color: var(--kore-text-muted);
  line-height: 1.5;
}

.welcome-banner {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: var(--kore-space-lg);
  margin-bottom: var(--kore-space-xl);
  padding: var(--kore-space-xl);
  border-radius: var(--kore-radius-lg);
  background: var(--kore-hero-gradient);
  border: 1px solid var(--kore-border);
}

.welcome-banner__eyebrow {
  margin: 0 0 var(--kore-space-xs);
  font-size: var(--kore-text-caption);
  font-weight: 600;
  letter-spacing: 0.08em;
  text-transform: uppercase;
  color: var(--kore-brand-gold);
}

.welcome-banner__title {
  margin: 0;
  font-size: var(--kore-text-h2);
  font-weight: 700;
}

.welcome-banner__icon {
  font-size: 3rem !important;
  color: var(--kore-brand-gold);
  opacity: 0.6;
}

.kpi-card--action {
  display: flex;
  flex-direction: column;
  gap: var(--kore-space-md);
}

.kpi-card--action .kpi-card__label {
  margin-bottom: auto;
}

.dashboard__cra-alert {
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  justify-content: space-between;
  gap: var(--kore-space-md);
  margin-bottom: var(--kore-space-lg);
  border-left: 4px solid var(--kore-warning);
}

.dashboard__cra-alert-text {
  margin: 0;
  color: var(--kore-text);
}

@media (max-width: 768px) {
  .welcome-banner {
    flex-direction: column;
    align-items: flex-start;
    padding: var(--kore-space-lg);
  }

  .welcome-banner__icon {
    align-self: flex-end;
  }

  .dashboard__kpis {
    grid-template-columns: repeat(2, minmax(0, 1fr));
  }
}
</style>
