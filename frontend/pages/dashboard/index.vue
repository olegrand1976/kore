<template>
  <div>
    <section class="welcome-banner">
      <div>
        <p class="welcome-banner__eyebrow">{{ $t('dashboard.banner') }}</p>
        <h2 class="welcome-banner__title">{{ $t('dashboard.welcome') }}</h2>
      </div>
      <AppIcon name="insights" class="welcome-banner__icon" />
    </section>

    <div class="dashboard-grid">
      <AppCard padding="lg" hoverable class="kpi-card">
        <div class="feature-card__icon"><AppIcon name="extension" /></div>
        <p class="kpi-card__value">6</p>
        <p class="kpi-card__label">{{ $t('dashboard.modules_active') }}</p>
      </AppCard>
      <AppCard padding="lg" hoverable class="kpi-card">
        <div class="feature-card__icon"><AppIcon name="pending_actions" /></div>
        <p class="kpi-card__value">2</p>
        <p class="kpi-card__label">{{ $t('dashboard.pending_validations') }}</p>
      </AppCard>
      <AppCard padding="lg" hoverable class="kpi-card kpi-card--action">
        <div class="feature-card__icon"><AppIcon name="schedule" /></div>
        <p class="kpi-card__label">{{ $t('nav.cra') }}</p>
        <AppButton variant="primary" size="sm" to="/cra">{{ $t('dashboard.quick_cra') }}</AppButton>
      </AppCard>
    </div>
  </div>
</template>

<script setup lang="ts">
definePageMeta({ layout: 'default' })
</script>

<style scoped>
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

.dashboard-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(220px, 1fr));
  gap: var(--kore-space-lg);
}

.kpi-card--action {
  display: flex;
  flex-direction: column;
  gap: var(--kore-space-md);
}

.kpi-card--action .kpi-card__label { margin-bottom: auto; }
</style>
