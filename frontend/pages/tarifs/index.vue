<template>
  <div class="page-shell">
    <PublicPageHero
      eyebrow="Kore SaaS"
      :title="$t('pricing.title')"
      :subtitle="$t('pricing.subtitle')"
      centered
    />

    <div class="card-grid card-grid--3">
      <PublicCard
        v-for="edition in editions"
        :key="edition.code"
        hoverable
        padding="lg"
        class="pricing-card feature-card"
        :class="{ 'pricing-card--featured': edition.highlight }"
      >
        <span v-if="edition.highlight" class="pricing-card__badge">{{ $t('pricing.featured_badge') }}</span>
        <span class="pricing-card__code">{{ $t(`pricing.editions.${edition.code}.eyebrow`) }}</span>
        <h3 class="feature-card__title">{{ $t(`pricing.editions.${edition.code}.title`) }}</h3>
        <p class="feature-card__desc">{{ $t(`pricing.editions.${edition.code}.desc`) }}</p>
        <ul class="pricing-card__features">
          <li v-for="(feature, index) in tm(`pricing.editions.${edition.code}.features`)" :key="index">
            <AppIcon name="check_circle" />
            <span>{{ rt(feature) }}</span>
          </li>
        </ul>
        <div class="pricing-card__price">
          <span class="pricing-card__amount">{{ formatPrice(edition.unitAmount) }}</span>
          <span class="pricing-card__unit">{{ $t('pricing.per_seat') }}</span>
          <span class="pricing-card__annual">{{ $t('pricing.annual_hint') }}</span>
        </div>
        <PublicButton
          :variant="edition.highlight ? 'primary' : 'secondary'"
          :to="`/billing/checkout?edition=${edition.code}`"
          class="pricing-card__cta"
        >
          {{ $t(`pricing.editions.${edition.code}.cta`) }}
        </PublicButton>
      </PublicCard>
    </div>

    <p class="pricing-note">{{ $t('pricing.note') }}</p>

    <section class="cta-band pricing-cta">
      <div>
        <h2 class="cta-band__title">{{ $t('pricing.cta_title') }}</h2>
        <p class="cta-band__text">{{ $t('pricing.cta_text') }}</p>
      </div>
      <div class="cta-actions">
        <PublicButton variant="primary" to="/reserver">{{ $t('brand.cta_book') }}</PublicButton>
        <PublicButton variant="secondary" to="/billing/checkout?edition=pro">{{ $t('billing.checkout') }}</PublicButton>
      </div>
    </section>
  </div>
</template>

<script setup lang="ts">
definePageMeta({ layout: 'public' })

const { tm, rt } = useI18n()
const { data } = await useFetch('/api/public/pricing')

const editions = computed(() => parsePricingEditions(data.value))

const formatPrice = (cents: number) =>
  new Intl.NumberFormat('fr-FR', { style: 'currency', currency: 'EUR' }).format(cents / 100)
</script>

<style scoped>
.pricing-card {
  position: relative;
  display: flex;
  flex-direction: column;
  gap: var(--kore-space-sm);
}

.pricing-card--featured {
  border-color: var(--kore-brand-gold);
  box-shadow: var(--kore-gold-glow);
}

.pricing-card__badge {
  position: absolute;
  top: calc(-1 * var(--kore-space-sm));
  right: var(--kore-space-md);
  padding: 0.2rem 0.65rem;
  font-size: var(--kore-text-caption);
  font-weight: 700;
  letter-spacing: 0.04em;
  text-transform: uppercase;
  color: var(--kore-text-inverse);
  background: var(--kore-brand-gold);
  border-radius: var(--kore-radius-full);
}

.pricing-card__code {
  display: inline-block;
  margin-bottom: var(--kore-space-xs);
  font-size: var(--kore-text-caption);
  font-weight: 600;
  letter-spacing: 0.06em;
  text-transform: uppercase;
  color: var(--kore-brand-blue);
}

.pricing-card__features {
  margin: var(--kore-space-sm) 0 0;
  padding: 0;
  list-style: none;
  display: flex;
  flex-direction: column;
  gap: var(--kore-space-xs);
  flex: 1;
}

.pricing-card__features li {
  display: flex;
  align-items: flex-start;
  gap: var(--kore-space-xs);
  font-size: var(--kore-text-small);
  color: var(--kore-text-muted);
  line-height: 1.45;
}

.pricing-card__features :deep(.material-symbols-outlined) {
  flex-shrink: 0;
  font-size: 1rem !important;
  color: var(--kore-success);
}

.pricing-card__price {
  margin-top: auto;
  padding-top: var(--kore-space-lg);
  border-top: 1px solid var(--kore-border);
}

.pricing-card__amount {
  display: block;
  font-size: var(--kore-text-h2);
  font-weight: 700;
  color: var(--kore-brand-gold);
}

.pricing-card__unit,
.pricing-card__annual {
  display: block;
  font-size: var(--kore-text-small);
  color: var(--kore-text-muted);
}

.pricing-card__cta {
  width: 100%;
  margin-top: var(--kore-space-sm);
}

.pricing-note {
  margin: var(--kore-space-xl) auto 0;
  max-width: 720px;
  text-align: center;
  font-size: var(--kore-text-small);
  color: var(--kore-text-muted);
  line-height: 1.55;
}

.pricing-cta {
  margin-top: var(--kore-space-2xl);
}

@media (max-width: 640px) {
  .pricing-card__cta :deep(.pub-btn) {
    width: 100%;
  }
}
</style>
