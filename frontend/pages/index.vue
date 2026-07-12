<template>
  <div class="landing">
    <section class="landing-hero" aria-labelledby="landing-title">
      <div class="landing-hero__glow" aria-hidden="true" />
      <div class="landing-hero__watermark" aria-hidden="true">
        <KoreLogo variant="emblem" size="xl" tone="auto" alt="" />
      </div>
      <div class="landing-hero__logo">
        <KoreLogo variant="hero" size="xl" show-tagline :alt="$t('brand.name')" />
      </div>
      <div class="landing-hero__content">
        <span class="page-hero__eyebrow">{{ $t('brand.tagline') }}</span>
        <h1 id="landing-title" class="landing-hero__title">{{ $t('brand.hero_title') }}</h1>
        <p class="landing-hero__subtitle">{{ $t('brand.hero_subtitle') }}</p>
        <div class="landing-hero__actions">
          <PublicButton variant="primary" to="/reserver">{{ $t('brand.cta_book') }}</PublicButton>
          <PublicButton variant="secondary" to="/tarifs">{{ $t('landing.cta_subscribe') }}</PublicButton>
        </div>
        <p class="landing-hero__guarantee">{{ $t('landing.hero_guarantee') }}</p>
        <dl class="landing-hero__stats">
          <div v-for="stat in outcomeStats" :key="stat.label" class="stat-pill stat-pill--outcome">
            <dd>{{ stat.value }}</dd>
            <dt>{{ stat.label }}</dt>
          </div>
        </dl>
      </div>
    </section>

    <PublicSection
      :eyebrow="$t('landing.args_eyebrow')"
      :title="$t('landing.args_title')"
      :subtitle="$t('landing.args_subtitle')"
      centered
      variant="elevated"
    >
      <div class="card-grid card-grid--3">
        <PublicCard v-for="item in argItems" :key="item.icon" hoverable padding="lg" class="feature-card arg-card">
          <div class="feature-card__icon" :class="`feature-card__icon--${item.tone}`"><AppIcon :name="item.icon" /></div>
          <h3 class="feature-card__title">{{ item.title }}</h3>
          <p class="feature-card__desc">{{ item.desc }}</p>
        </PublicCard>
      </div>
    </PublicSection>

    <PublicSection
      :eyebrow="$t('landing.workflow_eyebrow')"
      :title="$t('landing.workflow_title')"
      :subtitle="$t('landing.workflow_subtitle')"
      centered
    >
      <figure class="landing-diagram landing-diagram--solo">
        <img
          src="/brand/schema-workflow.svg"
          :alt="$t('landing.diagram_workflow_alt')"
          width="900"
          height="220"
          loading="lazy"
          decoding="async"
          class="landing-diagram__img"
        />
      </figure>
    </PublicSection>

    <PublicSection
      :eyebrow="$t('landing.diagram_architecture_eyebrow')"
      :title="$t('landing.diagram_architecture_title')"
      :subtitle="$t('landing.diagram_architecture_subtitle')"
      centered
      variant="elevated"
    >
      <figure class="landing-diagram landing-diagram--architecture landing-diagram--solo">
        <img
          src="/brand/schema-architecture.svg"
          :alt="$t('landing.diagram_architecture_alt')"
          width="700"
          height="420"
          loading="lazy"
          decoding="async"
          class="landing-diagram__img"
        />
      </figure>
    </PublicSection>

    <section class="mid-cta" aria-labelledby="mid-cta-title">
      <div>
        <h2 id="mid-cta-title" class="mid-cta__title">{{ $t('landing.mid_cta_title') }}</h2>
        <p class="mid-cta__text">{{ $t('landing.mid_cta_text') }}</p>
      </div>
      <PublicButton variant="primary" to="/reserver">{{ $t('brand.cta_book') }}</PublicButton>
    </section>

    <PublicSection :title="$t('pillars.title')" centered>
      <figure class="landing-diagram landing-diagram--pillars landing-diagram--solo">
        <img
          src="/brand/schema-pillars.svg"
          :alt="$t('landing.diagram_pillars_alt')"
          width="800"
          height="260"
          loading="lazy"
          decoding="async"
          class="landing-diagram__img"
        />
      </figure>
    </PublicSection>

    <PublicSection
      :title="$t('landing.modules_title')"
      :subtitle="$t('landing.modules_subtitle')"
      centered
    >
      <div class="card-grid card-grid--2">
        <PublicCard v-for="mod in previewModules" :key="mod.code" hoverable padding="lg" class="feature-card">
          <div class="feature-card__icon"><AppIcon :name="moduleIcon(mod.code)" /></div>
          <span class="module-code">{{ mod.code }}</span>
          <h3 class="feature-card__title">{{ mod.name }}</h3>
          <p class="feature-card__desc">{{ mod.description }}</p>
        </PublicCard>
      </div>
      <div class="landing-modules-link">
        <PublicButton variant="secondary" to="/modules">{{ $t('landing.cta_modules') }}</PublicButton>
      </div>
    </PublicSection>

    <section v-if="minPrice !== null" class="pricing-teaser" aria-labelledby="pricing-teaser-title">
      <span class="pricing-teaser__badge">{{ $t('landing.pricing_teaser_badge') }}</span>
      <div class="pricing-teaser__content">
        <h2 id="pricing-teaser-title" class="pricing-teaser__title">{{ $t('landing.pricing_teaser_title') }}</h2>
        <p class="pricing-teaser__subtitle">{{ $t('landing.pricing_teaser_subtitle') }}</p>
        <p class="pricing-teaser__price">
          <span class="pricing-teaser__from">{{ $t('landing.pricing_from') }}</span>
          <strong>{{ formatPrice(minPrice) }}</strong>
          <span class="pricing-teaser__unit">{{ $t('landing.pricing_per_seat') }}</span>
        </p>
        <p class="pricing-teaser__guarantee">{{ $t('landing.pricing_guarantee') }}</p>
      </div>
      <div class="pricing-teaser__actions">
        <PublicButton variant="primary" to="/tarifs">{{ $t('landing.pricing_cta') }}</PublicButton>
        <PublicButton variant="secondary" to="/reserver">{{ $t('brand.cta_book') }}</PublicButton>
      </div>
    </section>

    <PublicSection
      :eyebrow="$t('landing.faq_eyebrow')"
      :title="$t('landing.faq_title')"
      centered
    >
      <div class="faq-list">
        <details v-for="item in faqItems" :key="item.q" class="faq-item">
          <summary class="faq-item__q">{{ item.q }}</summary>
          <p class="faq-item__a">{{ item.a }}</p>
        </details>
      </div>
    </PublicSection>

    <section class="trust-bar" aria-label="Conformité">
      <div class="trust-bar__logo">
        <KoreLogo variant="emblem" size="lg" tone="color" :alt="$t('brand.name')" />
      </div>
      <div class="trust-bar__icon"><AppIcon name="verified_user" /></div>
      <div>
        <p class="trust-bar__title">{{ $t('compliance.title') }}</p>
        <p class="trust-bar__text">{{ $t('compliance.desc') }}</p>
      </div>
    </section>

    <section class="cta-band cta-band--final">
      <div class="cta-band__brand">
        <KoreLogo variant="emblem" size="md" tone="color" :alt="$t('brand.name')" />
      </div>
      <div>
        <h2 class="cta-band__title">{{ $t('landing.cta_title') }}</h2>
        <p class="cta-band__text">{{ $t('landing.cta_text') }}</p>
        <p class="cta-band__guarantee">{{ $t('landing.hero_guarantee') }}</p>
      </div>
      <div class="cta-band__actions">
        <PublicButton variant="primary" to="/reserver">{{ $t('brand.cta_book') }}</PublicButton>
        <PublicButton variant="secondary" to="/tarifs">{{ $t('landing.cta_subscribe') }}</PublicButton>
        <PublicButton variant="ghost" to="/login">{{ $t('landing.cta_login') }}</PublicButton>
      </div>
    </section>

    <aside v-show="showStickyCta" class="sticky-cta" aria-label="Actions rapides">
      <div class="sticky-cta__copy">
        <p class="sticky-cta__text">{{ $t('landing.sticky_cta') }}</p>
        <p class="sticky-cta__sub">{{ $t('landing.sticky_sub') }}</p>
      </div>
      <div class="sticky-cta__actions">
        <PublicButton variant="primary" to="/reserver">{{ $t('brand.cta_book') }}</PublicButton>
        <PublicButton variant="secondary" to="/tarifs">{{ $t('brand.cta_pricing') }}</PublicButton>
      </div>
    </aside>
  </div>
</template>

<script setup lang="ts">
definePageMeta({ layout: 'public' })

type ModuleItem = { code: string; name: string; description: string; unitAmount?: number }
type ArgTone = 'blue' | 'gold' | 'success'

const { t, locale } = useI18n()

const { data: modulesData } = await useFetch('/api/public/modules')
const { data: pricingData } = await useFetch('/api/public/pricing')

const previewModules = computed(() => {
  const list = (modulesData.value as { data?: ModuleItem[] })?.data ?? []
  return list.slice(0, 3)
})

const minPrice = computed(() => minEditionPrice(pricingData.value))

const formatPrice = (cents: number) =>
  new Intl.NumberFormat(locale.value === 'fr' ? 'fr-FR' : 'en-US', {
    style: 'currency',
    currency: 'EUR'
  }).format(cents / 100)

const moduleIcon = (code: string) => {
  const icons: Record<string, string> = {
    org: 'corporate_fare',
    cra: 'schedule',
    conges: 'beach_access',
    budget: 'account_balance',
    tma: 'support_agent',
    workflow: 'account_tree'
  }
  return icons[code] ?? 'extension'
}

const outcomeStats = computed(() => [
  { value: t('landing.stats_outcome_1_value'), label: t('landing.stats_outcome_1_label') },
  { value: t('landing.stats_outcome_2_value'), label: t('landing.stats_outcome_2_label') },
  { value: t('landing.stats_outcome_3_value'), label: t('landing.stats_outcome_3_label') }
])

const argItems = computed(() => [
  { icon: 'link_off', tone: 'blue' as ArgTone, title: t('landing.args_1_title'), desc: t('landing.args_1_desc') },
  { icon: 'trending_up', tone: 'gold' as ArgTone, title: t('landing.args_2_title'), desc: t('landing.args_2_desc') },
  { icon: 'verified', tone: 'success' as ArgTone, title: t('landing.args_3_title'), desc: t('landing.args_3_desc') }
])

const faqItems = computed(() => [
  { q: t('landing.faq_1_q'), a: t('landing.faq_1_a') },
  { q: t('landing.faq_2_q'), a: t('landing.faq_2_a') },
  { q: t('landing.faq_3_q'), a: t('landing.faq_3_a') }
])

const showStickyCta = ref(false)

onMounted(() => {
  const onScroll = () => {
    showStickyCta.value = window.scrollY > 480
  }
  onScroll()
  window.addEventListener('scroll', onScroll, { passive: true })
  onUnmounted(() => window.removeEventListener('scroll', onScroll))
})
</script>

<style scoped>
.landing {
  padding-bottom: calc(var(--kore-space-2xl) + 5rem);
}

.landing-hero {
  position: relative;
  display: grid;
  gap: var(--kore-space-2xl);
  align-items: center;
  margin: var(--kore-space-lg) 0 var(--kore-space-2xl);
  padding: var(--kore-space-2xl) var(--kore-space-xl);
  border-radius: var(--kore-radius-lg);
  background: var(--kore-hero-gradient);
  border: 1px solid var(--kore-border);
  overflow: hidden;
}

@media (min-width: 900px) {
  .landing-hero {
    grid-template-columns: minmax(260px, 1fr) minmax(320px, 1.1fr);
    gap: var(--kore-space-xl);
    padding: var(--kore-space-2xl);
  }
}

.landing-hero__glow {
  position: absolute;
  inset: 0;
  background:
    radial-gradient(circle at 15% 40%, rgba(43, 108, 176, 0.18), transparent 45%),
    radial-gradient(circle at 85% 60%, rgba(201, 162, 39, 0.14), transparent 50%);
  pointer-events: none;
}

.landing-hero__watermark {
  position: absolute;
  right: -2rem;
  bottom: -2rem;
  opacity: 0.06;
  pointer-events: none;
  z-index: 0;
}

.landing-hero__watermark :deep(.kore-logo__img) {
  width: min(280px, 40vw) !important;
}

.landing-diagram {
  margin: 0 0 var(--kore-space-xl);
  padding: 0;
  border-radius: var(--kore-radius-lg);
  overflow: hidden;
  border: 1px solid var(--kore-border);
  background: var(--kore-bg-subtle);
  box-shadow: var(--kore-shadow-sm);
}

.landing-diagram__img {
  display: block;
  width: 100%;
  height: auto;
}

.landing-diagram--solo {
  margin-bottom: 0;
}

.landing-diagram--architecture,
.landing-diagram--pillars {
  max-width: 800px;
  margin-inline: auto;
}

.arg-card .feature-card__desc {
  margin: 0;
}

.feature-card__icon--blue {
  background: rgba(43, 108, 176, 0.12);
  color: var(--kore-brand-blue);
}

.feature-card__icon--gold {
  background: rgba(201, 162, 39, 0.12);
  color: var(--kore-brand-gold);
}

.feature-card__icon--success {
  background: rgba(74, 222, 128, 0.12);
  color: var(--kore-success);
}

.landing-hero__logo,
.landing-hero__content {
  position: relative;
  z-index: 1;
}

.landing-hero__content {
  text-align: center;
}

@media (min-width: 900px) {
  .landing-hero__content {
    text-align: left;
  }
}

.landing-hero__logo {
  display: flex;
  justify-content: center;
}

.landing-hero__title {
  margin: 0 0 var(--kore-space-md);
  font-size: clamp(1.875rem, 4vw, var(--kore-text-display));
  font-weight: 700;
  line-height: 1.12;
  color: var(--kore-text);
}

.landing-hero__subtitle {
  margin: 0 auto var(--kore-space-xl);
  max-width: 540px;
  color: var(--kore-text-muted);
  line-height: 1.65;
  font-size: var(--kore-text-body);
}

@media (min-width: 900px) {
  .landing-hero__subtitle {
    margin: 0 0 var(--kore-space-xl);
  }
}

.landing-hero__actions {
  display: flex;
  flex-wrap: wrap;
  gap: var(--kore-space-md);
  margin-bottom: var(--kore-space-sm);
}

.landing-hero__guarantee {
  margin: 0 0 var(--kore-space-xl);
  font-size: var(--kore-text-caption);
  font-weight: 500;
  color: var(--kore-brand-gold);
  letter-spacing: 0.02em;
}

.landing-hero__stats {
  display: grid;
  grid-template-columns: repeat(3, 1fr);
  gap: var(--kore-space-sm);
  margin: 0;
}

.stat-pill {
  padding: var(--kore-space-md);
  text-align: center;
  border-radius: var(--kore-radius-md);
  background: color-mix(in srgb, var(--kore-bg-elevated) 75%, transparent);
  border: 1px solid var(--kore-border);
}

.stat-pill--outcome dd {
  margin: 0 0 var(--kore-space-xs);
  font-size: var(--kore-text-h2);
  font-weight: 800;
  color: var(--kore-brand-gold);
  line-height: 1.1;
}

.stat-pill--outcome dt {
  margin: 0;
  font-size: var(--kore-text-caption);
  font-weight: 500;
  color: var(--kore-text-muted);
  line-height: 1.35;
}

.feature-card__icon--warn {
  background: rgba(248, 113, 113, 0.12);
  color: var(--kore-error);
}

.value-card {
  position: relative;
  padding-top: calc(var(--kore-space-lg) + 1.5rem) !important;
}

.value-card__metric {
  position: absolute;
  top: var(--kore-space-md);
  right: var(--kore-space-md);
  padding: 0.2rem 0.6rem;
  font-size: var(--kore-text-caption);
  font-weight: 700;
  color: var(--kore-brand-gold);
  background: rgba(201, 162, 39, 0.12);
  border-radius: var(--kore-radius-full);
  border: 1px solid rgba(201, 162, 39, 0.25);
}

.workflow-steps {
  display: grid;
  gap: var(--kore-space-lg);
  margin: 0;
  padding: 0;
  list-style: none;
}

@media (min-width: 768px) {
  .workflow-steps {
    grid-template-columns: repeat(3, 1fr);
  }
}

.workflow-step {
  position: relative;
  padding: var(--kore-space-xl);
  border-radius: var(--kore-radius-lg);
  background: var(--kore-bg-elevated);
  border: 1px solid var(--kore-border);
  text-align: center;
}

.workflow-step__badge {
  position: absolute;
  top: var(--kore-space-md);
  right: var(--kore-space-md);
  width: 1.75rem;
  height: 1.75rem;
  display: flex;
  align-items: center;
  justify-content: center;
  border-radius: var(--kore-radius-full);
  background: var(--kore-brand-gold);
  color: var(--kore-text-inverse);
  font-size: var(--kore-text-caption);
  font-weight: 700;
}

.mid-cta {
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  justify-content: space-between;
  gap: var(--kore-space-lg);
  margin: var(--kore-space-xl) 0 var(--kore-space-2xl);
  padding: var(--kore-space-xl);
  border-radius: var(--kore-radius-lg);
  background: linear-gradient(90deg, rgba(43, 108, 176, 0.2), rgba(201, 162, 39, 0.15));
  border: 1px solid var(--kore-brand-gold);
}

.mid-cta__title {
  margin: 0 0 var(--kore-space-xs);
  font-size: var(--kore-text-h2);
  font-weight: 700;
}

.mid-cta__text {
  margin: 0;
  max-width: 520px;
  color: var(--kore-text-muted);
  line-height: 1.55;
}

.audience-card {
  border-left: 3px solid var(--kore-brand-blue);
}

.audience-card__bullets {
  margin: var(--kore-space-sm) 0 0;
  padding: 0;
  list-style: none;
  display: flex;
  flex-direction: column;
  gap: var(--kore-space-xs);
}

.audience-card__bullets li {
  display: flex;
  align-items: center;
  gap: var(--kore-space-xs);
  font-size: var(--kore-text-small);
  color: var(--kore-text-muted);
}

.audience-card__bullets :deep(.material-symbols-outlined) {
  font-size: 1rem !important;
  color: var(--kore-success);
}

.module-code {
  display: inline-block;
  margin-bottom: var(--kore-space-sm);
  padding: 0.125rem 0.5rem;
  font-size: var(--kore-text-caption);
  font-weight: 600;
  color: var(--kore-brand-blue);
  background: rgba(43, 108, 176, 0.1);
  border-radius: var(--kore-radius-full);
}

.landing-modules-link {
  display: flex;
  justify-content: center;
  margin-top: var(--kore-space-xl);
}

.pricing-teaser {
  position: relative;
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  justify-content: space-between;
  gap: var(--kore-space-lg);
  margin: var(--kore-space-2xl) 0;
  padding: var(--kore-space-xl);
  border-radius: var(--kore-radius-lg);
  background: linear-gradient(135deg, rgba(43, 108, 176, 0.15), rgba(201, 162, 39, 0.12));
  border: 1px solid var(--kore-brand-gold);
  box-shadow: var(--kore-gold-glow);
}

.pricing-teaser__badge {
  position: absolute;
  top: calc(-1 * var(--kore-space-sm));
  left: var(--kore-space-xl);
  padding: 0.25rem 0.75rem;
  font-size: var(--kore-text-caption);
  font-weight: 700;
  text-transform: uppercase;
  letter-spacing: 0.06em;
  color: var(--kore-text-inverse);
  background: var(--kore-brand-gold);
  border-radius: var(--kore-radius-full);
}

.pricing-teaser__title {
  margin: var(--kore-space-sm) 0 var(--kore-space-xs);
  font-size: var(--kore-text-h2);
  font-weight: 700;
}

.pricing-teaser__subtitle {
  margin: 0 0 var(--kore-space-md);
  color: var(--kore-text-muted);
}

.pricing-teaser__price {
  margin: 0 0 var(--kore-space-sm);
  display: flex;
  flex-wrap: wrap;
  align-items: baseline;
  gap: var(--kore-space-sm);
}

.pricing-teaser__from,
.pricing-teaser__unit {
  font-size: var(--kore-text-small);
  color: var(--kore-text-muted);
}

.pricing-teaser__price strong {
  font-size: var(--kore-text-h1);
  color: var(--kore-brand-gold);
}

.pricing-teaser__guarantee {
  margin: 0;
  font-size: var(--kore-text-caption);
  color: var(--kore-text-muted);
}

.pricing-teaser__actions {
  display: flex;
  flex-wrap: wrap;
  gap: var(--kore-space-md);
}

.faq-list {
  max-width: 720px;
  margin: 0 auto;
  display: flex;
  flex-direction: column;
  gap: var(--kore-space-sm);
}

.faq-item {
  border: 1px solid var(--kore-border);
  border-radius: var(--kore-radius-md);
  background: var(--kore-bg-elevated);
  overflow: hidden;
}

.faq-item__q {
  padding: var(--kore-space-md) var(--kore-space-lg);
  font-size: var(--kore-text-body);
  font-weight: 600;
  color: var(--kore-text);
  cursor: pointer;
  list-style: none;
}

.faq-item__q::-webkit-details-marker {
  display: none;
}

.faq-item[open] .faq-item__q {
  color: var(--kore-brand-gold);
  border-bottom: 1px solid var(--kore-border);
}

.faq-item__a {
  margin: 0;
  padding: var(--kore-space-md) var(--kore-space-lg);
  font-size: var(--kore-text-small);
  color: var(--kore-text-muted);
  line-height: 1.6;
}

.readiness-list {
  display: flex;
  flex-direction: column;
  gap: var(--kore-space-md);
  margin: 0;
  padding: 0;
  list-style: none;
  max-width: 720px;
  margin-inline: auto;
}

.readiness-item {
  display: flex;
  align-items: flex-start;
  gap: var(--kore-space-md);
  padding: var(--kore-space-md) var(--kore-space-lg);
  border-radius: var(--kore-radius-md);
  background: var(--kore-bg);
  border: 1px solid var(--kore-border);
}

.readiness-item__badge {
  flex-shrink: 0;
  padding: 0.2rem 0.6rem;
  font-size: var(--kore-text-caption);
  font-weight: 700;
  text-transform: uppercase;
  letter-spacing: 0.04em;
  border-radius: var(--kore-radius-full);
}

.readiness-item__badge--live {
  color: var(--kore-success);
  background: rgba(74, 222, 128, 0.12);
  border: 1px solid rgba(74, 222, 128, 0.3);
}

.readiness-item__badge--building {
  color: var(--kore-brand-gold);
  background: rgba(201, 162, 39, 0.12);
  border: 1px solid rgba(201, 162, 39, 0.3);
}

.readiness-item__badge--planned {
  color: var(--kore-brand-blue);
  background: rgba(43, 108, 176, 0.12);
  border: 1px solid rgba(43, 108, 176, 0.3);
}

.readiness-item__title {
  margin: 0 0 var(--kore-space-xs);
  font-size: var(--kore-text-body);
  font-weight: 600;
}

.readiness-item__desc {
  margin: 0;
  font-size: var(--kore-text-small);
  color: var(--kore-text-muted);
  line-height: 1.55;
}

.early-access {
  margin: var(--kore-space-2xl) 0;
  padding: var(--kore-space-2xl) var(--kore-space-xl);
  border-radius: var(--kore-radius-lg);
  background: linear-gradient(180deg, rgba(43, 108, 176, 0.1), rgba(201, 162, 39, 0.08));
  border: 1px dashed var(--kore-brand-gold);
  text-align: center;
}

.early-access__header {
  max-width: 640px;
  margin: 0 auto var(--kore-space-xl);
}

.early-access__title {
  margin: 0 0 var(--kore-space-sm);
  font-size: var(--kore-text-h2);
  font-weight: 700;
}

.early-access__subtitle {
  margin: 0;
  color: var(--kore-text-muted);
  line-height: 1.6;
}

.early-access__cta {
  display: flex;
  justify-content: center;
  margin-top: var(--kore-space-xl);
}

.trust-bar {
  display: flex;
  align-items: flex-start;
  gap: var(--kore-space-lg);
  padding: var(--kore-space-xl);
  border-radius: var(--kore-radius-lg);
  background: var(--kore-bg-elevated);
  border: 1px solid var(--kore-border);
  border-left: 4px solid var(--kore-brand-gold);
}

.trust-bar__logo {
  flex-shrink: 0;
  display: none;
}

@media (min-width: 640px) {
  .trust-bar__logo {
    display: block;
  }
}

.trust-bar__icon {
  display: flex;
  align-items: center;
  justify-content: center;
  flex-shrink: 0;
  width: 3rem;
  height: 3rem;
  border-radius: var(--kore-radius-full);
  background: rgba(201, 162, 39, 0.12);
  color: var(--kore-brand-gold);
}

.trust-bar__icon :deep(.material-symbols-outlined) {
  font-size: 1.5rem !important;
}

.trust-bar__title {
  margin: 0 0 var(--kore-space-xs);
  font-size: var(--kore-text-h3);
  font-weight: 600;
}

.trust-bar__text {
  margin: 0;
  font-size: var(--kore-text-small);
  color: var(--kore-text-muted);
  line-height: 1.55;
}

.cta-band--final {
  border: 1px solid var(--kore-brand-gold);
  box-shadow: var(--kore-gold-glow);
}

.cta-band__brand {
  flex-shrink: 0;
  display: flex;
  align-items: center;
  justify-content: center;
}

.cta-band__guarantee {
  margin: var(--kore-space-sm) 0 0;
  font-size: var(--kore-text-caption);
  font-weight: 500;
  color: var(--kore-brand-gold);
}

.cta-band__actions {
  display: flex;
  flex-wrap: wrap;
  gap: var(--kore-space-md);
}

.sticky-cta {
  position: fixed;
  bottom: 0;
  left: 0;
  right: 0;
  z-index: 60;
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  justify-content: space-between;
  gap: var(--kore-space-md);
  padding: var(--kore-space-md) var(--kore-space-xl);
  background: color-mix(in srgb, var(--kore-bg-elevated) 94%, transparent);
  border-top: 1px solid var(--kore-brand-gold);
  backdrop-filter: blur(12px);
  box-shadow: 0 -4px 24px rgba(0, 0, 0, 0.35);
}

.sticky-cta__text {
  margin: 0;
  font-size: var(--kore-text-small);
  font-weight: 700;
  color: var(--kore-text);
}

.sticky-cta__sub {
  margin: 0;
  font-size: var(--kore-text-caption);
  color: var(--kore-text-muted);
}

.sticky-cta__actions {
  display: flex;
  flex-wrap: wrap;
  gap: var(--kore-space-sm);
}

@media (max-width: 640px) {
  .landing-hero {
    margin: var(--kore-space-md) 0 var(--kore-space-xl);
    padding: var(--kore-space-xl) var(--kore-space-md);
  }

  .landing-hero__stats {
    grid-template-columns: 1fr;
  }

  .landing-hero__actions,
  .mid-cta,
  .pricing-teaser,
  .pricing-teaser__actions,
  .cta-band__actions,
  .sticky-cta,
  .sticky-cta__actions {
    flex-direction: column;
    align-items: stretch;
  }

  .landing-hero__actions .pub-btn,
  .pricing-teaser__actions .pub-btn,
  .cta-band__actions .pub-btn,
  .sticky-cta__actions .pub-btn,
  .mid-cta .pub-btn {
    width: 100%;
  }

  .trust-bar {
    flex-direction: column;
    padding: var(--kore-space-lg);
  }

  .sticky-cta {
    padding: var(--kore-space-md);
    padding-bottom: calc(var(--kore-space-md) + env(safe-area-inset-bottom, 0px));
  }
}
</style>
