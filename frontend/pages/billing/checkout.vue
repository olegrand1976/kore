<template>
  <div>
    <AppPageHeader :title="$t('billing.checkout_title')" :subtitle="$t('billing.checkout_subtitle')" />

    <p v-if="pending" class="muted">{{ $t('billing.checkout_loading') }}</p>

    <div v-else class="checkout-grid">
      <AppCard padding="lg" class="checkout-modules">
        <p v-if="activeEdition" class="edition-banner">
          {{ $t(`pricing.editions.${activeEdition.code}.title`) }}
          <span class="edition-banner__price">{{ formatPrice(activeEdition.unitAmount) }}{{ $t('pricing.per_seat') }}</span>
        </p>
        <h3 class="section-title">{{ $t('billing.checkout_select_modules') }}</h3>
        <ul class="module-list">
          <li v-for="mod in visibleModules" :key="mod.code" class="module-item">
            <label class="module-item__label">
              <input
                type="checkbox"
                :value="mod.code"
                :checked="selected.has(mod.code)"
                :disabled="Boolean(activeEdition)"
                @change="toggle(mod.code)"
              />
              <span class="module-item__text">
                <span class="module-item__name">{{ mod.name }}</span>
                <span class="module-item__desc">{{ mod.description }}</span>
              </span>
            </label>
            <span v-if="!activeEdition" class="module-item__price">{{ formatPrice(mod.unitAmount) }}</span>
            <span v-else-if="selected.has(mod.code)" class="module-item__included">{{ $t('pricing.included') }}</span>
          </li>
        </ul>
      </AppCard>

      <AppCard padding="lg" class="checkout-summary">
        <h3 class="section-title">{{ $t('billing.checkout_seats') }}</h3>
        <div class="seat-row">
          <AppButton variant="secondary" size="sm" :disabled="seats <= 1" @click="seats--">−</AppButton>
          <input v-model.number="seats" type="number" min="1" class="seat-input" />
          <AppButton variant="secondary" size="sm" @click="seats++">+</AppButton>
        </div>
        <p class="muted seat-hint">{{ $t('billing.checkout_seats_hint') }}</p>

        <dl class="total">
          <dt>{{ $t('billing.checkout_total') }}</dt>
          <dd class="total__amount">{{ formatPrice(totalCents) }}</dd>
        </dl>

        <p v-if="errorMsg" class="error-text">{{ errorMsg }}</p>

        <AppButton
          variant="primary"
          class="checkout-submit"
          :disabled="selected.size === 0 || loading"
          @click="onSubmit"
        >
          {{ $t('billing.checkout_submit') }}
        </AppButton>
      </AppCard>
    </div>
  </div>
</template>

<script setup lang="ts">
definePageMeta({ layout: 'default', middleware: ['admin'] })

type ModulePrice = { code: string; name: string; description: string; unitAmount: number }

const { t } = useI18n()
const { startCheckout, loading } = useBilling()

const route = useRoute()
const { data, pending } = await useFetch('/api/public/pricing')

const modules = computed<ModulePrice[]>(() => parsePricingModules(data.value))
const editions = computed(() => parsePricingEditions(data.value))

const selected = ref(new Set<string>())
const seats = ref(1)
const errorMsg = ref<string | null>(null)

function modulesForEdition(code: string | undefined) {
  if (!code) {
    return []
  }
  const edition = editions.value.find((item) => item.code === code)
  return edition?.modules ?? []
}

watch(
  [modules, editions, () => route.query.edition],
  ([list, editionList, editionCode]) => {
    const preset = modulesForEdition(typeof editionCode === 'string' ? editionCode : undefined)
    if (preset.length > 0) {
      selected.value = new Set(preset.filter((code) => list.some((mod) => mod.code === code)))
      return
    }
    if (selected.value.size === 0 && list.length > 0) {
      const pro = editionList.find((edition) => edition.code === 'pro')
      const fallback = pro?.modules?.length ? pro.modules : list.map((mod) => mod.code)
      selected.value = new Set(fallback.filter((code) => list.some((mod) => mod.code === code)))
    }
  },
  { immediate: true }
)

const toggle = (code: string) => {
  if (activeEdition.value) {
    return
  }
  const next = new Set(selected.value)
  if (next.has(code)) {
    next.delete(code)
  } else {
    next.add(code)
  }
  selected.value = next
}

const activeEdition = computed(() => {
  const code = route.query.edition
  if (typeof code !== 'string') {
    return null
  }
  return editions.value.find((edition) => edition.code === code) ?? null
})

const visibleModules = computed(() => {
  if (activeEdition.value) {
    return modules.value.filter((mod) => activeEdition.value!.modules.includes(mod.code))
  }
  return modules.value
})

const totalCents = computed(() => {
  const seatCount = Math.max(1, seats.value)
  if (activeEdition.value) {
    return activeEdition.value.unitAmount * seatCount
  }
  const unit = modules.value
    .filter((m) => selected.value.has(m.code))
    .reduce((sum, m) => sum + m.unitAmount, 0)
  return unit * seatCount
})

const formatPrice = (cents: number) =>
  new Intl.NumberFormat('fr-FR', { style: 'currency', currency: 'EUR' }).format(cents / 100)

const onSubmit = async () => {
  errorMsg.value = null
  if (selected.value.size === 0) {
    errorMsg.value = t('billing.checkout_empty_modules')
    return
  }
  const origin = import.meta.client ? window.location.origin : ''
  try {
    await startCheckout({
      modules: Array.from(selected.value),
      seats: Math.max(1, seats.value),
      successUrl: `${origin}/billing/success`,
      cancelUrl: `${origin}/billing/cancel`
    })
  } catch {
    errorMsg.value = t('billing.checkout_error')
  }
}
</script>

<style scoped>
.checkout-grid {
  display: grid;
  grid-template-columns: 2fr 1fr;
  gap: var(--kore-space-lg);
  align-items: start;
}

.section-title {
  margin: 0 0 var(--kore-space-md);
  font-size: var(--kore-text-h3);
  font-weight: 600;
}

.edition-banner {
  display: flex;
  flex-wrap: wrap;
  align-items: baseline;
  justify-content: space-between;
  gap: var(--kore-space-sm);
  margin: 0 0 var(--kore-space-md);
  padding: var(--kore-space-sm) var(--kore-space-md);
  border-radius: var(--kore-radius-md);
  background: rgba(201, 162, 39, 0.1);
  border: 1px solid rgba(201, 162, 39, 0.25);
  font-weight: 600;
}

.edition-banner__price {
  font-size: var(--kore-text-small);
  color: var(--kore-brand-gold);
}

.module-item__included {
  font-size: var(--kore-text-caption);
  font-weight: 600;
  color: var(--kore-success);
  white-space: nowrap;
}

.module-list { margin: 0; padding: 0; list-style: none; }

.module-item {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: var(--kore-space-md);
  padding: var(--kore-space-md) 0;
  border-bottom: 1px solid var(--kore-border);
}

.module-item__label {
  display: flex;
  align-items: flex-start;
  gap: var(--kore-space-sm);
  cursor: pointer;
}

.module-item__text { display: flex; flex-direction: column; }
.module-item__name { font-weight: 600; }
.module-item__desc { font-size: var(--kore-text-small); color: var(--kore-text-muted); }
.module-item__price { font-weight: 600; color: var(--kore-brand-gold); white-space: nowrap; }

.seat-row { display: flex; align-items: center; gap: var(--kore-space-sm); }

.seat-input {
  width: 4rem;
  text-align: center;
  padding: var(--kore-space-xs) var(--kore-space-sm);
  border: 1px solid var(--kore-border);
  border-radius: var(--kore-radius-sm);
  background: var(--kore-surface);
  color: var(--kore-text);
}

.seat-hint { margin: var(--kore-space-sm) 0 0; }

.total {
  display: flex;
  align-items: baseline;
  justify-content: space-between;
  margin: var(--kore-space-lg) 0;
  padding-top: var(--kore-space-lg);
  border-top: 1px solid var(--kore-border);
}

.total dt { color: var(--kore-text-muted); }
.total__amount { font-size: var(--kore-text-h2); font-weight: 700; color: var(--kore-brand-gold); }

.error-text { color: var(--kore-danger, #c0392b); margin-bottom: var(--kore-space-sm); }

.checkout-submit { width: 100%; }

.muted { color: var(--kore-text-muted); }

@media (max-width: 768px) {
  .checkout-grid { grid-template-columns: 1fr; }
}
</style>
