<template>
  <div class="page-shell">
    <PublicPageHero
      eyebrow="Suite modulaire"
      :title="$t('modules_page.title')"
      :subtitle="$t('modules_page.subtitle')"
      centered
    />

    <div class="card-grid card-grid--2">
      <PublicCard v-for="mod in modules" :key="mod.code" hoverable padding="lg" class="feature-card">
        <div class="feature-card__icon"><AppIcon :name="moduleIcon(mod.code)" /></div>
        <span class="module-code">{{ mod.code }}</span>
        <h3 class="feature-card__title">{{ mod.name }}</h3>
        <p class="feature-card__desc">{{ mod.description }}</p>
      </PublicCard>
    </div>
  </div>
</template>

<script setup lang="ts">
definePageMeta({ layout: 'public' })

const { data } = await useFetch('/api/public/modules')
const modules = computed(() => (data.value as { data?: Array<{ code: string; name: string; description: string }> })?.data ?? [])

const moduleIcon = (code: string) => {
  const icons: Record<string, string> = {
    org: 'corporate_fare', cra: 'schedule', conges: 'beach_access',
    budget: 'account_balance', tma: 'support_agent', workflow: 'account_tree'
  }
  return icons[code] ?? 'extension'
}
</script>

<style scoped>
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
</style>
