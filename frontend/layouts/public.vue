<template>
  <div class="layout-public">
    <header class="header">
      <NuxtLink to="/" class="header__brand">
        <KoreLogo variant="horizontal" size="sm" alt="Kore" />
      </NuxtLink>
      <nav class="header__nav" aria-label="Navigation principale">
        <NuxtLink to="/modules">{{ $t('nav.modules') }}</NuxtLink>
        <NuxtLink to="/tarifs">{{ $t('nav.pricing') }}</NuxtLink>
        <NuxtLink to="/reserver">{{ $t('nav.book') }}</NuxtLink>
      </nav>
      <div class="header__actions">
        <button type="button" class="menu-btn" aria-label="Menu" @click="drawerOpen = true">
          <AppIcon name="menu" />
        </button>
        <ThemeToggle />
        <button type="button" class="chip-btn" @click="toggleLocale">{{ locale === 'fr' ? 'EN' : 'FR' }}</button>
        <PublicButton variant="primary" to="/login" class="header__login">{{ $t('nav.login') }}</PublicButton>
      </div>
    </header>
    <main class="main"><slot /></main>
    <PublicFooter />

    <MobileDrawer v-model:open="drawerOpen">
      <NuxtLink to="/modules" class="drawer-link" @click="drawerOpen = false">{{ $t('nav.modules') }}</NuxtLink>
      <NuxtLink to="/tarifs" class="drawer-link" @click="drawerOpen = false">{{ $t('nav.pricing') }}</NuxtLink>
      <NuxtLink to="/reserver" class="drawer-link" @click="drawerOpen = false">{{ $t('nav.book') }}</NuxtLink>
      <NuxtLink to="/login" class="drawer-link" @click="drawerOpen = false">{{ $t('nav.login') }}</NuxtLink>
    </MobileDrawer>
  </div>
</template>

<script setup lang="ts">
const { locale, setLocale } = useI18n()
const drawerOpen = ref(false)

const toggleLocale = () => setLocale(locale.value === 'fr' ? 'en' : 'fr')

useHead({
  meta: [
    { property: 'og:image', content: '/brand/kore-logo-hero.png' },
    { property: 'og:title', content: 'Kore — PSA/ESN modulaire' }
  ]
})
</script>

<style scoped>
.layout-public {
  min-height: 100vh;
  display: flex;
  flex-direction: column;
  background: var(--kore-bg);
}

.header {
  position: sticky;
  top: 0;
  z-index: 50;
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: var(--kore-space-md);
  padding: var(--kore-space-md) var(--kore-space-xl);
  background: color-mix(in srgb, var(--kore-bg-elevated) 85%, transparent);
  border-bottom: 1px solid var(--kore-border);
  backdrop-filter: blur(12px);
}

.header__brand { text-decoration: none; flex-shrink: 0; }

.header__nav { display: none; gap: var(--kore-space-xs); }

@media (min-width: 768px) { .header__nav { display: flex; } }

.header__nav a {
  padding: 0.5rem 0.875rem;
  color: var(--kore-text-muted);
  text-decoration: none;
  font-size: var(--kore-text-small);
  font-weight: 500;
  border-radius: var(--kore-radius-md);
  transition: color 0.15s, background 0.15s;
}

.header__nav a:hover,
.header__nav a.router-link-active {
  color: var(--kore-brand-gold);
  background: rgba(201, 162, 39, 0.08);
}

.header__actions {
  display: flex;
  align-items: center;
  gap: var(--kore-space-sm);
}

.menu-btn {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  width: 2.25rem;
  height: 2.25rem;
  border: 1px solid var(--kore-border);
  border-radius: var(--kore-radius-md);
  background: var(--kore-bg-elevated);
  color: var(--kore-text-muted);
  cursor: pointer;
}

@media (min-width: 768px) {
  .menu-btn { display: none; }
}

.chip-btn {
  padding: 0.375rem 0.625rem;
  font-size: var(--kore-text-caption);
  font-weight: 600;
  border: 1px solid var(--kore-border);
  border-radius: var(--kore-radius-full);
  background: var(--kore-bg-elevated);
  color: var(--kore-text-muted);
  cursor: pointer;
}

.chip-btn:hover {
  color: var(--kore-brand-gold);
  border-color: var(--kore-brand-gold);
}

.main {
  flex: 1;
  width: 100%;
  max-width: var(--kore-container-max);
  margin: 0 auto;
  padding: 0 var(--kore-space-xl);
}

.drawer-link {
  display: block;
  padding: 0.75rem 0.875rem;
  color: var(--kore-text);
  text-decoration: none;
  font-size: var(--kore-text-body);
  font-weight: 500;
  border-radius: var(--kore-radius-md);
}

.drawer-link:hover,
.drawer-link.router-link-active {
  color: var(--kore-brand-gold);
  background: rgba(201, 162, 39, 0.1);
}

@media (max-width: 767px) {
  .header { padding: var(--kore-space-sm) var(--kore-space-md); }
  .main { padding: 0 var(--kore-space-md); }
  .header__login :deep(.pub-btn) {
    padding: 0.5rem 0.875rem;
    font-size: var(--kore-text-caption);
  }
}
</style>
