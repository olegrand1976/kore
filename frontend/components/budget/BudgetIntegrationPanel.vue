<script setup lang="ts">
const steps = computed(() => [
  {
    key: 'cra',
    icon: 'schedule',
    titleKey: 'budget.integration_cra_title',
    descKey: 'budget.integration_cra_desc',
    to: '/cra'
  },
  {
    key: 'budget',
    icon: 'account_balance_wallet',
    titleKey: 'budget.integration_budget_title',
    descKey: 'budget.integration_budget_desc',
    to: null
  },
  {
    key: 'tma',
    icon: 'support_agent',
    titleKey: 'budget.integration_tma_title',
    descKey: 'budget.integration_tma_desc',
    to: '/tma'
  }
])
</script>

<template>
  <AppCard padding="lg" class="integration-panel">
    <h2 class="integration-panel__title">{{ $t('budget.integration_title') }}</h2>
    <ol class="integration-panel__steps">
      <li v-for="(step, index) in steps" :key="step.key" class="integration-panel__step">
        <div class="integration-panel__icon-wrap" aria-hidden="true">
          <AppIcon :name="step.icon" />
        </div>
        <div class="integration-panel__body">
          <div class="integration-panel__head">
            <span class="integration-panel__index">{{ index + 1 }}</span>
            <strong>{{ $t(step.titleKey) }}</strong>
          </div>
          <p class="integration-panel__desc">{{ $t(step.descKey) }}</p>
          <AppButton
            v-if="step.to"
            variant="ghost"
            size="sm"
            class="integration-panel__link"
            @click="navigateTo(step.to!)"
          >
            {{ $t('budget.open') }}
          </AppButton>
        </div>
      </li>
    </ol>
  </AppCard>
</template>

<style scoped>
.integration-panel__title {
  margin: 0 0 var(--kore-space-lg);
  font-size: var(--kore-text-body);
}

.integration-panel__steps {
  list-style: none;
  margin: 0;
  padding: 0;
  display: grid;
  gap: var(--kore-space-lg);
}

.integration-panel__step {
  display: grid;
  grid-template-columns: auto 1fr;
  gap: var(--kore-space-md);
  align-items: start;
}

.integration-panel__icon-wrap {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 2.5rem;
  height: 2.5rem;
  border-radius: var(--kore-radius-md);
  background: var(--kore-bg-subtle);
  color: var(--kore-accent);
}

.integration-panel__head {
  display: flex;
  align-items: center;
  gap: var(--kore-space-sm);
  margin-bottom: 0.25rem;
}

.integration-panel__index {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  width: 1.25rem;
  height: 1.25rem;
  border-radius: var(--kore-radius-full);
  background: var(--kore-brand-gold);
  color: var(--kore-text-inverse);
  font-size: var(--kore-text-caption);
  font-weight: 700;
}

.integration-panel__desc {
  margin: 0 0 var(--kore-space-sm);
  font-size: var(--kore-text-small);
  color: var(--kore-text-muted);
}

.integration-panel__link {
  margin-top: var(--kore-space-xs);
}

@media (max-width: 768px) {
  .integration-panel__step {
    grid-template-columns: 1fr;
  }

  .integration-panel__icon-wrap {
    width: 2rem;
    height: 2rem;
  }
}
</style>
