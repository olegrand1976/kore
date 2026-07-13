<template>
  <div>
    <AppPageHeader :title="$t('settings.title')" />

    <nav class="parametres-tabs" role="tablist" :aria-label="$t('settings.tabs_label')">
      <NuxtLink
        to="/admin/parametres/conges"
        role="tab"
        class="parametres-tab"
        :class="{ 'parametres-tab--active': route.path.startsWith('/admin/parametres/conges') }"
      >
        {{ $t('settings.tab_conges') }}
      </NuxtLink>
      <NuxtLink
        to="/admin/parametres/cra"
        role="tab"
        class="parametres-tab"
        :class="{ 'parametres-tab--active': route.path.startsWith('/admin/parametres/cra') }"
      >
        {{ $t('settings.tab_cra') }}
      </NuxtLink>
      <NuxtLink
        to="/admin/parametres/ia"
        role="tab"
        class="parametres-tab"
        :class="{ 'parametres-tab--active': route.path.startsWith('/admin/parametres/ia') }"
      >
        {{ $t('settings.tab_ia') }}
      </NuxtLink>
    </nav>

    <NuxtPage />
  </div>
</template>

<script setup lang="ts">
definePageMeta({ layout: 'default', middleware: 'admin' })

const route = useRoute()
</script>

<style scoped>
.parametres-tabs {
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

.parametres-tab {
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

.parametres-tab:hover {
  color: var(--kore-text);
}

.parametres-tab--active {
  color: var(--kore-text-inverse);
  background: var(--kore-brand-gold);
}

@media (max-width: 768px) {
  .parametres-tabs {
    width: 100%;
  }

  .parametres-tab {
    flex: 1 1 auto;
    min-width: 0;
  }
}
</style>
