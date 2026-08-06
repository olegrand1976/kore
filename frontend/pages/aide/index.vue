<template>
  <div class="help-hub">
    <AppPageHeader :title="$t('help.title')" :subtitle="$t('help.subtitle')" />

    <AppCard v-if="effectiveProfiles.length" padding="lg" class="help-hub__you">
      <h2 class="help-hub__section-title">{{ $t('help.your_profiles') }}</h2>
      <div class="help-hub__badges">
        <AppBadge v-for="p in effectiveProfiles" :key="p" variant="default">{{ p }}</AppBadge>
      </div>
      <p class="help-hub__hint">{{ $t('help.your_profiles_hint') }}</p>
    </AppCard>

    <div class="help-hub__grid">
      <NuxtLink to="/aide/acces" class="help-hub__card">
        <AppCard padding="lg" hoverable class="help-hub__card-inner">
          <div class="help-hub__card-icon" aria-hidden="true">
            <AppIcon name="lock_person" />
          </div>
          <h2 class="help-hub__card-title">{{ $t('help.topics.access.title') }}</h2>
          <p class="help-hub__card-desc">{{ $t('help.topics.access.description') }}</p>
          <span class="help-hub__card-cta">{{ $t('help.open') }}</span>
        </AppCard>
      </NuxtLink>
    </div>
  </div>
</template>

<script setup lang="ts">
definePageMeta({ layout: 'default' })

const { effectiveProfiles } = useAuth()
</script>

<style scoped>
.help-hub {
  display: flex;
  flex-direction: column;
  gap: var(--kore-space-lg);
}

.help-hub__section-title {
  margin: 0 0 var(--kore-space-sm);
  font-size: var(--kore-text-h3);
  font-weight: 600;
  color: var(--kore-text);
}

.help-hub__badges {
  display: flex;
  flex-wrap: wrap;
  gap: var(--kore-space-sm);
}

.help-hub__hint {
  margin: var(--kore-space-md) 0 0;
  font-size: var(--kore-text-small);
  color: var(--kore-text-muted);
}

.help-hub__grid {
  display: grid;
  grid-template-columns: 1fr;
  gap: var(--kore-space-md);
}

@media (min-width: 769px) {
  .help-hub__grid {
    grid-template-columns: repeat(auto-fill, minmax(280px, 1fr));
  }
}

.help-hub__card {
  text-decoration: none;
  color: inherit;
  display: block;
}

.help-hub__card:focus-visible {
  outline: 2px solid var(--kore-focus-ring);
  outline-offset: 2px;
  border-radius: var(--kore-radius-md);
}

.help-hub__card-inner {
  height: 100%;
}

.help-hub__card-icon {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  width: 2.5rem;
  height: 2.5rem;
  margin-bottom: var(--kore-space-md);
  border-radius: var(--kore-radius-md);
  background: var(--kore-bg-subtle);
  color: var(--kore-brand-gold);
}

.help-hub__card-title {
  margin: 0 0 var(--kore-space-sm);
  font-size: var(--kore-text-h3);
  font-weight: 600;
  color: var(--kore-text);
}

.help-hub__card-desc {
  margin: 0 0 var(--kore-space-md);
  font-size: var(--kore-text-small);
  color: var(--kore-text-muted);
  line-height: 1.45;
}

.help-hub__card-cta {
  font-size: var(--kore-text-small);
  font-weight: 600;
  color: var(--kore-link);
}
</style>
