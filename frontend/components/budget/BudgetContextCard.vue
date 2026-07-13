<script setup lang="ts">
import type { OrgApplication } from '~/composables/useApplications'

const props = defineProps<{
  application?: OrgApplication | null
  budgetType: string
  currency: string
}>()

const { t } = useI18n()
const { budgetTypeLabel, normalizeBudgetType } = useBudgetDisplay()
const { pickAppLabel, pickAppClient } = useApplications()

const clientLabel = computed(() => pickAppClient(props.application) || t('budget.context_empty_client'))
const isDefault = computed(() => normalizeBudgetType(props.budgetType) === 'defaut')
</script>

<template>
  <AppCard padding="lg" class="budget-context">
    <h2 class="budget-context__title">{{ $t('budget.context_title') }}</h2>
    <dl class="budget-context__dl">
      <div>
        <dt>{{ $t('budget.context_application') }}</dt>
        <dd>{{ pickAppLabel(application) || '—' }}</dd>
      </div>
      <div>
        <dt>{{ $t('budget.context_client') }}</dt>
        <dd>{{ clientLabel }}</dd>
      </div>
      <div>
        <dt>{{ $t('budget.col_type') }}</dt>
        <dd><AppBadge variant="gold">{{ budgetTypeLabel(budgetType) }}</AppBadge></dd>
      </div>
      <div>
        <dt>{{ $t('budget.col_currency') }}</dt>
        <dd>{{ currency }}</dd>
      </div>
      <div v-if="isDefault">
        <dt>{{ $t('budget.context_role') }}</dt>
        <dd class="budget-context__role">{{ $t('budget.type_defaut_help') }}</dd>
      </div>
    </dl>
  </AppCard>
</template>

<style scoped>
.budget-context__title {
  margin: 0 0 var(--kore-space-md);
  font-size: var(--kore-text-body);
}

.budget-context__dl {
  display: grid;
  gap: var(--kore-space-md);
  margin: 0;
}

.budget-context__dl div {
  display: flex;
  justify-content: space-between;
  gap: var(--kore-space-sm);
}

.budget-context__dl dt {
  color: var(--kore-text-muted);
  flex-shrink: 0;
}

.budget-context__dl dd {
  margin: 0;
  text-align: right;
}

.budget-context__role {
  max-width: 28rem;
  font-size: var(--kore-text-small);
  color: var(--kore-text-muted);
}

@media (max-width: 640px) {
  .budget-context__dl div {
    flex-direction: column;
    align-items: flex-start;
  }

  .budget-context__dl dd {
    text-align: left;
  }
}
</style>
