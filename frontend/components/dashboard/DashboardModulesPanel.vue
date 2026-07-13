<script setup lang="ts">
import type { ModuleCode } from '~/composables/useEntitlements'
import {
  matchEdition,
  modulesMissingFromEdition,
  parsePricingEditions,
  parsePricingModules,
  suggestUpgradeEdition,
  type EditionPrice,
  type ModulePrice
} from '~/composables/usePricingCatalog'

const props = defineProps<{
  open: boolean
  activeModules: ModuleCode[]
  subscriptionStatus: string
  seats?: number | null
}>()

const emit = defineEmits<{ close: [] }>()

const { t } = useI18n()
const { isAdmin } = useAuth()
const { openPortal, loading: portalLoading } = useBilling()

const { data: pricingData, pending: pricingPending } = await useFetch('/api/public/pricing')

const catalogModules = computed(() => parsePricingModules(pricingData.value))
const editions = computed(() => parsePricingEditions(pricingData.value))

const matchedEdition = computed(() => matchEdition(props.activeModules, editions.value))
const upgradeEdition = computed(() => suggestUpgradeEdition(matchedEdition.value, editions.value))
const missingForUpgrade = computed(() =>
  upgradeEdition.value ? modulesMissingFromEdition(props.activeModules, upgradeEdition.value) : []
)

const activeModuleDetails = computed(() =>
  props.activeModules.map((code) => {
    const fromCatalog = catalogModules.value.find((mod) => mod.code === code)
    if (fromCatalog) {
      return fromCatalog
    }
    return {
      code,
      name: t(`dashboard.modules_panel.codes.${code}`, code),
      description: '',
      unitAmount: 0
    }
  })
)

const lockedModules = computed(() => {
  const active = new Set(props.activeModules.map((code) => code.toLowerCase()))
  return catalogModules.value.filter((mod) => !active.has(mod.code.toLowerCase()))
})

const editionTitle = computed(() => {
  if (matchedEdition.value) {
    return t(`pricing.editions.${matchedEdition.value.code}.title`)
  }
  return t('dashboard.modules_panel.custom_plan')
})

const editionDesc = computed(() => {
  if (matchedEdition.value) {
    return t(`pricing.editions.${matchedEdition.value.code}.desc`)
  }
  return t('dashboard.modules_panel.custom_plan_desc')
})

const statusLabel = computed(() => {
  const key = props.subscriptionStatus.toLowerCase()
  switch (key) {
    case 'trial':
    case 'active':
    case 'past_due':
    case 'suspended':
    case 'canceled':
      return t(`dashboard.modules_panel.status.${key}`)
    default:
      return props.subscriptionStatus || '—'
  }
})

const statusVariant = computed(() => {
  switch (props.subscriptionStatus.toLowerCase()) {
    case 'active':
    case 'trial':
      return 'success'
    case 'past_due':
      return 'warning'
    case 'suspended':
    case 'canceled':
      return 'default'
    default:
      return 'default'
  }
})

const checkoutUpgradeUrl = computed(() =>
  upgradeEdition.value ? `/billing/checkout?edition=${upgradeEdition.value.code}` : '/tarifs'
)

const moduleLabel = (mod: ModulePrice) => mod.name || t(`dashboard.modules_panel.codes.${mod.code}`, mod.code)

const onBackdropClick = () => emit('close')

const onOpenPortal = async () => {
  if (!import.meta.client) {
    return
  }
  await openPortal(window.location.href)
}

const formatPrice = (cents: number) =>
  new Intl.NumberFormat('fr-FR', { style: 'currency', currency: 'EUR' }).format(cents / 100)

const editionPriceLabel = (edition: EditionPrice) =>
  `${formatPrice(edition.unitAmount)}${t('pricing.per_seat')}`
</script>

<template>
  <Teleport to="body">
    <div
      v-if="open"
      class="modules-panel-overlay"
      role="presentation"
      @click.self="onBackdropClick"
    >
      <AppCard
        padding="lg"
        class="modules-panel"
        role="dialog"
        aria-modal="true"
        :aria-label="$t('dashboard.modules_panel.title')"
      >
        <header class="modules-panel__header">
          <div>
            <p class="modules-panel__eyebrow">{{ $t('dashboard.modules_panel.title') }}</p>
            <h2 class="modules-panel__title">{{ editionTitle }}</h2>
            <p class="modules-panel__desc">{{ editionDesc }}</p>
          </div>
          <AppButton variant="ghost" size="sm" :aria-label="$t('common.close')" @click="emit('close')">
            <AppIcon name="close" />
          </AppButton>
        </header>

        <div class="modules-panel__meta">
          <AppBadge :variant="statusVariant">{{ statusLabel }}</AppBadge>
          <span v-if="seats != null" class="modules-panel__seats">
            {{ $t('dashboard.modules_panel.seats', { n: seats }) }}
          </span>
        </div>

        <section class="modules-panel__section">
          <h3 class="modules-panel__section-title">{{ $t('dashboard.modules_panel.active_title') }}</h3>
          <p v-if="pricingPending" class="modules-panel__muted">{{ $t('billing.checkout_loading') }}</p>
          <ul v-else class="modules-panel__list">
            <li v-for="mod in activeModuleDetails" :key="mod.code" class="modules-panel__item modules-panel__item--active">
              <AppIcon name="check_circle" class="modules-panel__icon modules-panel__icon--active" />
              <span class="modules-panel__item-text">
                <span class="modules-panel__item-name">{{ moduleLabel(mod) }}</span>
                <span class="modules-panel__item-desc">{{ mod.description }}</span>
              </span>
            </li>
          </ul>
        </section>

        <section v-if="lockedModules.length > 0" class="modules-panel__section">
          <h3 class="modules-panel__section-title">{{ $t('dashboard.modules_panel.extend_title') }}</h3>
          <p class="modules-panel__hint">
            <template v-if="upgradeEdition">
              {{ $t('dashboard.modules_panel.extend_hint_edition', { edition: $t(`pricing.editions.${upgradeEdition.code}.title`) }) }}
            </template>
            <template v-else>
              {{ $t('dashboard.modules_panel.extend_hint_custom') }}
            </template>
          </p>
          <ul class="modules-panel__list">
            <li v-for="mod in lockedModules" :key="mod.code" class="modules-panel__item">
              <AppIcon name="lock" class="modules-panel__icon" />
              <span class="modules-panel__item-text">
                <span class="modules-panel__item-name">{{ moduleLabel(mod) }}</span>
                <span class="modules-panel__item-desc">{{ mod.description }}</span>
              </span>
              <span v-if="missingForUpgrade.includes(mod.code)" class="modules-panel__tag">
                {{ $t('dashboard.modules_panel.in_upgrade', { edition: upgradeEdition ? $t(`pricing.editions.${upgradeEdition.code}.title`) : '' }) }}
              </span>
            </li>
          </ul>
          <p v-if="upgradeEdition" class="modules-panel__upgrade-price">
            {{ $t('dashboard.modules_panel.upgrade_from', { price: editionPriceLabel(upgradeEdition) }) }}
          </p>
        </section>

        <footer class="modules-panel__actions">
          <template v-if="isAdmin">
            <AppButton variant="primary" size="sm" :to="checkoutUpgradeUrl">
              {{ $t('dashboard.modules_panel.cta_upgrade') }}
            </AppButton>
            <AppButton variant="secondary" size="sm" to="/billing/abonnement">
              {{ $t('dashboard.modules_panel.cta_manage') }}
            </AppButton>
            <AppButton variant="ghost" size="sm" :disabled="portalLoading" @click="onOpenPortal">
              {{ $t('billing.portal') }}
            </AppButton>
          </template>
          <template v-else>
            <p class="modules-panel__admin-hint">{{ $t('dashboard.modules_panel.admin_only') }}</p>
            <AppButton variant="secondary" size="sm" to="/tarifs">
              {{ $t('brand.cta_pricing') }}
            </AppButton>
          </template>
          <AppButton variant="ghost" size="sm" @click="emit('close')">
            {{ $t('common.close') }}
          </AppButton>
        </footer>
      </AppCard>
    </div>
  </Teleport>
</template>

<style scoped>
.modules-panel-overlay {
  position: fixed;
  inset: 0;
  z-index: 50;
  display: flex;
  align-items: flex-end;
  justify-content: center;
  padding: var(--kore-space-md);
  padding-bottom: calc(var(--kore-space-md) + env(safe-area-inset-bottom, 0px));
  background: rgba(0, 0, 0, 0.45);
}

.modules-panel {
  width: min(100%, 36rem);
  max-height: min(90vh, 720px);
  overflow: auto;
}

.modules-panel__header {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  gap: var(--kore-space-md);
  margin-bottom: var(--kore-space-md);
}

.modules-panel__eyebrow {
  margin: 0 0 var(--kore-space-xs);
  font-size: var(--kore-text-caption);
  font-weight: 600;
  letter-spacing: 0.06em;
  text-transform: uppercase;
  color: var(--kore-brand-gold);
}

.modules-panel__title {
  margin: 0;
  font-size: var(--kore-text-h3);
  font-weight: 700;
}

.modules-panel__desc {
  margin: var(--kore-space-xs) 0 0;
  font-size: var(--kore-text-small);
  color: var(--kore-text-muted);
  line-height: 1.45;
}

.modules-panel__meta {
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  gap: var(--kore-space-sm);
  margin-bottom: var(--kore-space-lg);
}

.modules-panel__seats {
  font-size: var(--kore-text-small);
  color: var(--kore-text-muted);
}

.modules-panel__section {
  margin-bottom: var(--kore-space-lg);
}

.modules-panel__section-title {
  margin: 0 0 var(--kore-space-sm);
  font-size: var(--kore-text-body);
  font-weight: 600;
}

.modules-panel__hint,
.modules-panel__muted,
.modules-panel__admin-hint,
.modules-panel__upgrade-price {
  margin: 0 0 var(--kore-space-sm);
  font-size: var(--kore-text-small);
  color: var(--kore-text-muted);
  line-height: 1.45;
}

.modules-panel__list {
  margin: 0;
  padding: 0;
  list-style: none;
  display: flex;
  flex-direction: column;
  gap: var(--kore-space-sm);
}

.modules-panel__item {
  display: flex;
  align-items: flex-start;
  gap: var(--kore-space-sm);
  padding: var(--kore-space-sm) var(--kore-space-md);
  border: 1px solid var(--kore-border);
  border-radius: var(--kore-radius-md);
  background: var(--kore-surface);
}

.modules-panel__item--active {
  border-color: rgba(201, 162, 39, 0.25);
  background: rgba(201, 162, 39, 0.06);
}

.modules-panel__item-text {
  display: flex;
  flex-direction: column;
  gap: 0.15rem;
  flex: 1;
  min-width: 0;
}

.modules-panel__item-name {
  font-weight: 600;
  font-size: var(--kore-text-small);
}

.modules-panel__item-desc {
  font-size: var(--kore-text-caption);
  color: var(--kore-text-muted);
  line-height: 1.35;
}

.modules-panel__icon {
  flex-shrink: 0;
  font-size: 1.125rem !important;
  color: var(--kore-text-muted);
  margin-top: 0.1rem;
}

.modules-panel__icon--active {
  color: var(--kore-success);
}

.modules-panel__tag {
  flex-shrink: 0;
  align-self: center;
  font-size: var(--kore-text-caption);
  font-weight: 600;
  color: var(--kore-brand-gold);
  white-space: nowrap;
}

.modules-panel__actions {
  display: flex;
  flex-wrap: wrap;
  gap: var(--kore-space-sm);
  padding-top: var(--kore-space-md);
  border-top: 1px solid var(--kore-border);
}

@media (min-width: 640px) {
  .modules-panel-overlay {
    align-items: center;
  }
}

@media (max-width: 768px) {
  .modules-panel__actions :deep(.app-button) {
    flex: 1 1 calc(50% - var(--kore-space-sm));
    justify-content: center;
  }

  .modules-panel__tag {
    display: none;
  }
}
</style>
