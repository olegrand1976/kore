<template>
  <div>
    <AppPageHeader :title="$t('nav.conges')">
      <template v-if="showIndexActions" #actions>
        <AppButton v-if="guideRef?.dismissed" variant="ghost" size="sm" type="button" @click="guideRef?.showAgain()">
          {{ $t('guides.show') }}
        </AppButton>
        <AppButton variant="ghost" size="sm" @click="navigateTo('/conges/soldes')">
          {{ $t('conges.balances_link') }}
        </AppButton>
        <AppButton variant="primary" size="sm" @click="indexActions?.toggleForm()">
          {{ $t('conges.new_request') }}
        </AppButton>
      </template>
    </AppPageHeader>

    <nav class="conges-tabs" role="tablist" :aria-label="$t('conges.tabs_label')">
      <NuxtLink
        to="/conges"
        role="tab"
        class="conges-tab"
        :class="{ 'conges-tab--active': isRequestsTab }"
        :aria-selected="isRequestsTab"
      >
        {{ $t('conges.tab_requests') }}
      </NuxtLink>
      <NuxtLink
        v-if="canValidateConges"
        to="/conges/validation"
        role="tab"
        class="conges-tab"
        :class="{ 'conges-tab--active': isValidationTab }"
        :aria-selected="isValidationTab"
      >
        {{ $t('conges.tab_validation') }}
      </NuxtLink>
    </nav>

    <AppSectionGuide v-if="showIndexActions" ref="guideRef" guide-key="conges" />

    <NuxtPage />
  </div>
</template>

<script setup lang="ts">
definePageMeta({ layout: 'default' })

const route = useRoute()
const guideRef = ref<{ showAgain: () => void; dismissed: boolean } | null>(null)
const { fetchSession } = useAuth()
const { canValidateConges } = usePermissions()

await fetchSession()

const indexActions = useState<{ toggleForm: () => void } | null>('conges-index-actions', () => null)

const isRequestsTab = computed(() => {
  const path = route.path
  return path === '/conges' || path === '/conges/soldes'
})

const isValidationTab = computed(() => route.path === '/conges/validation')

const showIndexActions = computed(() => route.path === '/conges')
</script>

<style scoped>
.conges-tabs {
  display: flex;
  flex-wrap: wrap;
  gap: var(--kore-space-xs);
  margin: calc(-1 * var(--kore-space-md)) 0 var(--kore-space-lg);
  padding: 0.25rem;
  border: 1px solid var(--kore-border);
  border-radius: var(--kore-radius-md);
  background: var(--kore-bg-elevated);
  width: fit-content;
  max-width: 100%;
}

.conges-tab {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  padding: 0.5rem 1rem;
  border-radius: calc(var(--kore-radius-md) - 2px);
  color: var(--kore-text-muted);
  font-size: var(--kore-text-small);
  font-weight: 600;
  text-decoration: none;
  transition: background 0.15s, color 0.15s;
  white-space: nowrap;
}

.conges-tab:hover {
  color: var(--kore-text);
}

.conges-tab--active {
  color: var(--kore-text-inverse);
  background: var(--kore-brand-gold);
}

@media (max-width: 768px) {
  .conges-tabs {
    width: 100%;
  }

  .conges-tab {
    flex: 1 1 calc(50% - var(--kore-space-xs));
    min-width: 0;
  }
}
</style>
