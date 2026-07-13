<script setup lang="ts">
import type { DashboardCharts, DashboardStatErrors } from '~/composables/useDashboardStats'
import type { ModuleCode } from '~/composables/useEntitlements'

const props = defineProps<{
  charts: DashboardCharts
  errors: DashboardStatErrors
  loading?: boolean
  showModule: (code: ModuleCode) => boolean
}>()

const { t } = useI18n()

const hasAnyChart = computed(
  () =>
    props.loading ||
    props.showModule('tma') ||
    props.showModule('budget') ||
    props.showModule('cra') ||
    props.showModule('conges')
)
</script>

<template>
  <section v-if="hasAnyChart || loading" class="dashboard-charts" aria-labelledby="dashboard-charts-title">
    <h2 id="dashboard-charts-title" class="dashboard-charts__title">{{ $t('dashboard.charts.title') }}</h2>

    <div class="dashboard-charts__grid">
      <AppCard v-if="showModule('tma')" padding="lg" class="dashboard-charts__panel">
        <header class="dashboard-charts__head">
          <AppIcon name="support_agent" />
          <div>
            <h3>{{ $t('dashboard.charts.tma_title') }}</h3>
            <p>{{ $t('dashboard.charts.tma_desc') }}</p>
          </div>
          <AppButton variant="ghost" size="sm" to="/tma">{{ $t('dashboard.charts.view_module') }}</AppButton>
        </header>
        <AppBarChart
          :bars="charts.tmaStatus"
          :loading="loading"
          :empty-label="$t('dashboard.charts.empty_tma')"
        />
        <p v-if="errors.tma" class="dashboard-charts__error">{{ $t('common.unavailable') }}</p>
      </AppCard>

      <AppCard v-if="showModule('budget')" padding="lg" class="dashboard-charts__panel">
        <header class="dashboard-charts__head">
          <AppIcon name="account_balance" />
          <div>
            <h3>{{ $t('dashboard.charts.budget_title') }}</h3>
            <p>{{ $t('dashboard.charts.budget_desc') }}</p>
          </div>
          <AppButton variant="ghost" size="sm" to="/budget">{{ $t('dashboard.charts.view_module') }}</AppButton>
        </header>
        <AppBarChart
          :bars="charts.budgetConsumption"
          :loading="loading"
          :max-value="100"
          value-suffix="%"
          :empty-label="$t('dashboard.charts.empty_budget')"
        />
        <p v-if="errors.budget" class="dashboard-charts__error">{{ $t('common.unavailable') }}</p>
      </AppCard>

      <AppCard v-if="showModule('cra')" padding="lg" class="dashboard-charts__panel">
        <header class="dashboard-charts__head">
          <AppIcon name="schedule" />
          <div>
            <h3>{{ $t('dashboard.charts.cra_title') }}</h3>
            <p>{{ $t('dashboard.charts.cra_desc') }}</p>
          </div>
          <AppButton variant="ghost" size="sm" to="/cra">{{ $t('dashboard.charts.view_module') }}</AppButton>
        </header>
        <CraStatusTimeline
          :months="charts.craMonths"
          :loading="loading"
          :empty-label="$t('dashboard.charts.empty_cra')"
        />
        <p v-if="errors.cra" class="dashboard-charts__error">{{ $t('common.unavailable') }}</p>
      </AppCard>

      <AppCard v-if="showModule('conges')" padding="lg" class="dashboard-charts__panel">
        <header class="dashboard-charts__head">
          <AppIcon name="beach_access" />
          <div>
            <h3>{{ $t('dashboard.charts.leave_title') }}</h3>
            <p>{{ $t('dashboard.charts.leave_desc') }}</p>
          </div>
          <AppButton variant="ghost" size="sm" to="/conges">{{ $t('dashboard.charts.view_module') }}</AppButton>
        </header>
        <AppBarChart
          :bars="charts.leaveStatus"
          :loading="loading"
          :empty-label="$t('dashboard.charts.empty_leave')"
        />
        <p v-if="errors.conges" class="dashboard-charts__error">{{ $t('common.unavailable') }}</p>
      </AppCard>
    </div>
  </section>
</template>

<style scoped>
.dashboard-charts {
  margin-top: var(--kore-space-xl);
}

.dashboard-charts__title {
  margin: 0 0 var(--kore-space-lg);
  font-size: var(--kore-text-h3);
  font-weight: 700;
}

.dashboard-charts__grid {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: var(--kore-space-lg);
}

.dashboard-charts__panel {
  display: flex;
  flex-direction: column;
  gap: var(--kore-space-md);
  min-height: 14rem;
}

.dashboard-charts__head {
  display: grid;
  grid-template-columns: auto 1fr auto;
  gap: var(--kore-space-sm) var(--kore-space-md);
  align-items: start;
}

.dashboard-charts__head :deep(.material-symbols-outlined) {
  font-size: 1.5rem !important;
  color: var(--kore-brand-gold);
  margin-top: 0.125rem;
}

.dashboard-charts__head h3 {
  margin: 0;
  font-size: var(--kore-text-body);
  font-weight: 600;
}

.dashboard-charts__head p {
  margin: 0.15rem 0 0;
  font-size: var(--kore-text-caption);
  color: var(--kore-text-muted);
}

.dashboard-charts__error {
  margin: 0;
  font-size: var(--kore-text-caption);
  color: var(--kore-error);
}

@media (max-width: 900px) {
  .dashboard-charts__grid {
    grid-template-columns: 1fr;
  }
}

@media (max-width: 640px) {
  .dashboard-charts__head {
    grid-template-columns: auto 1fr;
  }

  .dashboard-charts__head :deep(.app-btn) {
    grid-column: 1 / -1;
    width: 100%;
  }
}
</style>
