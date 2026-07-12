<template>
  <div class="layout-app">
    <aside class="sidebar">
      <div class="sidebar__brand">
        <TenantLogo :logo-url="branding.logoUrl" :alt="branding.raisonSociale" size="md" />
        <p v-if="branding.raisonSociale" class="sidebar__company">{{ branding.raisonSociale }}</p>
      </div>
      <nav class="sidebar__nav" aria-label="Navigation applicative">
        <NuxtLink v-for="item in navItems" :key="item.to" :to="item.to">
          <AppIcon :name="item.icon" />
          {{ item.label }}
        </NuxtLink>
      </nav>
    </aside>
    <div class="content">
      <header class="topbar">
        <div class="topbar__left">
          <button type="button" class="menu-btn" aria-label="Menu" @click="drawerOpen = true">
            <AppIcon name="menu" />
          </button>
          <div class="topbar__breadcrumb">
            <AppIcon name="grid_view" />
            <span>{{ pageTitle }}</span>
          </div>
        </div>
        <div class="topbar__actions">
          <ThemeToggle />
          <button type="button" class="chip-btn" @click="toggleLocale">{{ locale === 'fr' ? 'EN' : 'FR' }}</button>
          <AppButton variant="ghost" size="sm" class="topbar__logout" @click="logout">{{ $t('nav.logout') }}</AppButton>
        </div>
      </header>
      <main class="main"><slot /></main>
    </div>

    <AppBottomNav :items="bottomNavItems" />

    <MobileDrawer v-model:open="drawerOpen">
      <p class="drawer-title">{{ branding.raisonSociale || 'Kore' }}</p>
      <NuxtLink
        v-for="item in navItems"
        :key="item.to"
        :to="item.to"
        class="drawer-link"
        @click="drawerOpen = false"
      >
        <AppIcon :name="item.icon" />
        {{ item.label }}
      </NuxtLink>
      <button type="button" class="drawer-link drawer-link--btn" @click="logout">
        <AppIcon name="logout" />
        {{ $t('nav.logout') }}
      </button>
    </MobileDrawer>
  </div>
</template>

<script setup lang="ts">
const { locale, setLocale, t } = useI18n()
const route = useRoute()
const { branding, fetchBranding } = useTenantBranding()
const { fetchSession, isAdmin } = useAuth()
const drawerOpen = ref(false)

onMounted(async () => {
  await Promise.all([fetchBranding(), fetchSession()])
})

const toggleLocale = () => setLocale(locale.value === 'fr' ? 'en' : 'fr')

const allNavItems = computed(() => [
  { to: '/dashboard', icon: 'dashboard', label: t('nav.dashboard'), adminOnly: false },
  { to: '/cra', icon: 'schedule', label: t('nav.cra'), adminOnly: false },
  { to: '/admin/organisation', icon: 'corporate_fare', label: t('nav.organisation'), adminOnly: true },
  { to: '/admin/users', icon: 'group', label: t('nav.users'), adminOnly: true }
])

const navItems = computed(() =>
  allNavItems.value.filter(item => !item.adminOnly || isAdmin.value)
)

const bottomNavItems = computed(() => navItems.value.slice(0, 4))

const pageTitle = computed(() => {
  if (route.path.startsWith('/cra/') && route.params.id) {
    return t('nav.cra')
  }
  const item = allNavItems.value.find(n => route.path === n.to || route.path.startsWith(n.to + '/'))
  return item?.label ?? 'Kore'
})

const logout = async () => {
  drawerOpen.value = false
  await $fetch('/api/auth/logout', { method: 'POST' })
  await navigateTo('/login')
}
</script>

<style scoped>
.layout-app {
  display: grid;
  grid-template-columns: 260px 1fr;
  min-height: 100vh;
  background: var(--kore-bg);
}

.sidebar {
  display: flex;
  flex-direction: column;
  gap: var(--kore-space-xl);
  padding: var(--kore-space-lg);
  background: var(--kore-bg-elevated);
  border-right: 1px solid var(--kore-border);
}

.sidebar__brand {
  padding-bottom: var(--kore-space-md);
  border-bottom: 1px solid var(--kore-border);
}

.sidebar__company {
  margin: var(--kore-space-sm) 0 0;
  font-size: var(--kore-text-caption);
  font-weight: 500;
  color: var(--kore-text-muted);
}

.sidebar__nav {
  display: flex;
  flex-direction: column;
  gap: var(--kore-space-xs);
}

.sidebar__nav a {
  display: flex;
  align-items: center;
  gap: var(--kore-space-sm);
  padding: 0.625rem 0.875rem;
  color: var(--kore-text-muted);
  text-decoration: none;
  border-radius: var(--kore-radius-md);
  font-size: var(--kore-text-small);
  font-weight: 500;
  transition: background 0.15s, color 0.15s;
}

.sidebar__nav a:hover,
.sidebar__nav a.router-link-active {
  color: var(--kore-brand-gold);
  background: rgba(201, 162, 39, 0.1);
}

.content { display: flex; flex-direction: column; min-width: 0; }

.topbar {
  display: flex;
  justify-content: space-between;
  align-items: center;
  gap: var(--kore-space-sm);
  padding: var(--kore-space-md) var(--kore-space-xl);
  border-bottom: 1px solid var(--kore-border);
  background: color-mix(in srgb, var(--kore-bg-elevated) 90%, transparent);
  backdrop-filter: blur(8px);
}

.topbar__left {
  display: flex;
  align-items: center;
  gap: var(--kore-space-sm);
  min-width: 0;
}

.menu-btn {
  display: none;
  align-items: center;
  justify-content: center;
  width: 2.25rem;
  height: 2.25rem;
  border: 1px solid var(--kore-border);
  border-radius: var(--kore-radius-md);
  background: var(--kore-bg);
  color: var(--kore-text-muted);
  cursor: pointer;
}

.topbar__breadcrumb {
  display: flex;
  align-items: center;
  gap: var(--kore-space-sm);
  font-size: var(--kore-text-small);
  font-weight: 600;
  color: var(--kore-text-muted);
  min-width: 0;
}

.topbar__breadcrumb span {
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.topbar__breadcrumb :deep(.material-symbols-outlined) {
  font-size: 1.125rem !important;
  color: var(--kore-brand-gold);
  flex-shrink: 0;
}

.topbar__actions {
  display: flex;
  align-items: center;
  gap: var(--kore-space-sm);
  flex-shrink: 0;
}

.chip-btn {
  padding: 0.375rem 0.625rem;
  font-size: var(--kore-text-caption);
  font-weight: 600;
  border: 1px solid var(--kore-border);
  border-radius: var(--kore-radius-full);
  background: var(--kore-bg);
  color: var(--kore-text-muted);
  cursor: pointer;
}

.main {
  flex: 1;
  padding: var(--kore-space-xl);
  max-width: 1200px;
}

.drawer-title {
  margin: 0 0 var(--kore-space-lg);
  font-size: var(--kore-text-small);
  font-weight: 600;
  color: var(--kore-text-muted);
}

.drawer-link {
  display: flex;
  align-items: center;
  gap: var(--kore-space-sm);
  padding: 0.75rem 0.875rem;
  color: var(--kore-text);
  text-decoration: none;
  font-size: var(--kore-text-small);
  font-weight: 500;
  border: none;
  background: transparent;
  border-radius: var(--kore-radius-md);
  cursor: pointer;
  width: 100%;
  text-align: left;
}

.drawer-link:hover,
.drawer-link.router-link-active {
  color: var(--kore-brand-gold);
  background: rgba(201, 162, 39, 0.1);
}

@media (max-width: 768px) {
  .layout-app { grid-template-columns: 1fr; }
  .sidebar { display: none; }
  .menu-btn { display: inline-flex; }
  .topbar { padding: var(--kore-space-md); }
  .topbar__logout { display: none; }
  .main {
    padding: var(--kore-space-md);
    padding-bottom: calc(4.5rem + env(safe-area-inset-bottom, 0px));
  }
}
</style>
