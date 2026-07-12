<template>
  <div class="page-shell">
    <PublicPageHero
      eyebrow="Kore SaaS"
      :title="$t('pricing.title')"
      :subtitle="$t('pricing.subtitle')"
      centered
    />

    <div class="card-grid card-grid--2">
      <PublicCard
        v-for="(mod, index) in modules"
        :key="mod.code"
        hoverable
        padding="lg"
        class="pricing-card feature-card"
        :class="{ 'pricing-card--featured': index === 1 }"
      >
        <span class="pricing-card__code">{{ mod.code }}</span>
        <h3 class="feature-card__title">{{ mod.name }}</h3>
        <p class="feature-card__desc">{{ mod.description }}</p>
        <div class="pricing-card__price">
          <span class="pricing-card__amount">{{ formatPrice(mod.unitAmount) }}</span>
          <span class="pricing-card__unit">{{ $t('pricing.per_seat') }}</span>
        </div>
      </PublicCard>
    </div>

    <section class="cta-band pricing-cta">
      <div>
        <h2 class="cta-band__title">{{ $t('pricing.cta') }}</h2>
        <p class="cta-band__text">{{ $t('pricing.subtitle') }}</p>
      </div>
      <PublicButton variant="primary" to="/reserver">{{ $t('brand.cta_book') }}</PublicButton>
    </section>
  </div>
</template>

<script setup lang="ts">
definePageMeta({ layout: 'public' })

const { data } = await useFetch('/api/public/pricing')

type ModulePrice = { code: string; name: string; description: string; unitAmount: number }

const modules = computed(() => {
  const catalog = (data.value as { data?: { modules?: ModulePrice[] } })?.data
  return catalog?.modules ?? []
})

const formatPrice = (cents: number) =>
  new Intl.NumberFormat('fr-FR', { style: 'currency', currency: 'EUR' }).format(cents / 100)
</script>

<style scoped>
.pricing-card { position: relative; }

.pricing-card--featured {
  border-color: var(--kore-brand-gold);
  box-shadow: var(--kore-gold-glow);
}

.pricing-card__code {
  display: inline-block;
  margin-bottom: var(--kore-space-sm);
  font-size: var(--kore-text-caption);
  font-weight: 600;
  letter-spacing: 0.06em;
  text-transform: uppercase;
  color: var(--kore-brand-blue);
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

.pricing-card__unit {
  font-size: var(--kore-text-small);
  color: var(--kore-text-muted);
}

.pricing-cta { margin-top: var(--kore-space-2xl); }
</style>
