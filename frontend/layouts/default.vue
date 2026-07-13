<template>
  <div class="layout-app">
    <aside class="sidebar">
      <div class="sidebar__brand">
        <TenantLogo :logo-url="branding.logoUrl" :alt="branding.raisonSociale" size="md" />
        <p v-if="branding.raisonSociale" class="sidebar__company">{{ branding.raisonSociale }}</p>
      </div>
      <nav class="sidebar__nav" aria-label="Navigation applicative">
        <NuxtLink
          v-for="item in mainNavItems"
          :key="item.to"
          :to="item.to"
          :class="{ 'router-link-active': isNavActive(item) }"
        >
          <AppIcon :name="item.icon" />
          {{ item.label }}
        </NuxtLink>

        <div v-if="settingsNavItems.length > 0" class="sidebar__divider" />
        <p v-if="settingsNavItems.length > 0" class="sidebar__section-label">
          {{ t('nav.settings') }}
        </p>
        <NuxtLink
          v-for="item in settingsNavItems"
          :key="item.to"
          :to="item.to"
          :class="{ 'router-link-active': isNavActive(item) }"
        >
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
            <AppIcon :name="activeNavItem?.icon ?? 'grid_view'" />
            <span>{{ pageTitle }}</span>
          </div>
        </div>
        <div class="topbar__actions">
          <ThemeToggle variant="chip" />
          <button type="button" class="chip-btn" aria-label="Language" @click="toggleLocale">{{ locale === 'fr' ? 'EN' : 'FR' }}</button>
          <button
            type="button"
            class="chip-btn"
            :aria-label="$t('release_notes.open')"
            @click="openReleaseNotes()"
          >
            <AppIcon name="new_releases" />
          </button>
          <AppButton variant="ghost" size="sm" class="topbar__logout" @click="logout">{{ $t('nav.logout') }}</AppButton>
        </div>
      </header>
      <main class="main" :class="{ 'main--narrow': isNarrowMain }">
        <p v-if="isPastDue" class="past-due" role="alert">{{ $t('billing.past_due_banner') }}</p>
        <slot />
      </main>
    </div>

    <AppReleaseNotesModal
      v-model:open="releaseNotesOpen"
      v-model:selectedMonthKey="releaseNotesSelectedMonthKey"
      :months="releaseNotesMonths"
      :loading="releaseNotesLoading"
      :current-version="releaseNotesMeta?.currentVersion"
      :last-seen-version="releaseNotesMeta?.lastSeenVersion"
      :auto-show-enabled="releaseNotesAutoShowEnabled"
      @update:autoShowEnabled="onUpdateReleaseNotesAutoShow"
      @refresh="refreshReleaseNotes"
      @markSeen="markReleaseNotesSeen"
    />

    <AppBottomNav :items="bottomNavItems" />

    <MobileDrawer v-model:open="drawerOpen">
      <p class="drawer-title">{{ branding.raisonSociale || 'Kore' }}</p>
      <NuxtLink
        v-for="item in mainNavItems"
        :key="item.to"
        :to="item.to"
        class="drawer-link"
        :class="{ 'router-link-active': isNavActive(item) }"
        @click="drawerOpen = false"
      >
        <AppIcon :name="item.icon" />
        {{ item.label }}
      </NuxtLink>

      <div v-if="settingsNavItems.length > 0" class="drawer-divider" />
      <p v-if="settingsNavItems.length > 0" class="drawer-section-label">
        {{ t('nav.settings') }}
      </p>
      <NuxtLink
        v-for="item in settingsNavItems"
        :key="item.to"
        :to="item.to"
        class="drawer-link"
        :class="{ 'router-link-active': isNavActive(item) }"
        @click="drawerOpen = false"
      >
        <AppIcon :name="item.icon" />
        {{ item.label }}
      </NuxtLink>
      <button type="button" class="drawer-link drawer-link--btn" @click="onToggleTheme">
        <AppIcon :name="theme === 'dark' ? 'light_mode' : 'dark_mode'" />
        {{ theme === 'dark' ? $t('theme.switch_light') : $t('theme.switch_dark') }}
      </button>
      <button type="button" class="drawer-link drawer-link--btn" @click="logout">
        <AppIcon name="logout" />
        {{ $t('nav.logout') }}
      </button>
    </MobileDrawer>
  </div>
</template>

<script setup lang="ts">
const { locale, setLocale, t } = useI18n()
const { theme, toggleTheme } = useTheme()
const route = useRoute()
const { branding, fetchBranding } = useTenantBranding()
const { fetchSession, isAdmin, isPlatformAdmin } = useAuth()
const { fetchEntitlements, hasModule, isPastDue } = useEntitlements()
const { apiFetch } = useApiFetch()
const drawerOpen = ref(false)

onMounted(async () => {
  await Promise.all([fetchBranding(), fetchSession(), fetchEntitlements()])
  await maybeAutoOpenReleaseNotes()
})

const toggleLocale = () => setLocale(locale.value === 'fr' ? 'en' : 'fr')

const onToggleTheme = () => toggleTheme()

type ReleaseNotesMeta = {
  currentVersion: string
  lastSeenVersion: string | null
  autoShowEnabled: boolean
}

type ReleaseNotesMonth = {
  key: string
  label: string
  items: Array<{
    sha: string
    shortSha: string
    message: string
    authorName?: string
    date: string
    htmlUrl?: string
  }>
}

const releaseNotesOpen = ref(false)
const releaseNotesLoading = ref(false)
const releaseNotesMeta = ref<ReleaseNotesMeta | null>(null)
const releaseNotesMonths = ref<ReleaseNotesMonth[]>([])
const releaseNotesSelectedMonthKey = ref('')
const releaseNotesAutoShowEnabled = ref(true)

const fetchReleaseNotesMeta = async () => {
  releaseNotesMeta.value = await apiFetch<ReleaseNotesMeta>('/api/release-notes/meta')
  releaseNotesAutoShowEnabled.value = releaseNotesMeta.value.autoShowEnabled
}

const fetchReleaseNotesCommits = async () => {
  const res = await apiFetch<{ months: ReleaseNotesMonth[]; defaultMonthKey: string }>('/api/release-notes/commits')
  releaseNotesMonths.value = res.months
  releaseNotesSelectedMonthKey.value = res.defaultMonthKey
}

const refreshReleaseNotes = async () => {
  releaseNotesLoading.value = true
  try {
    await fetchReleaseNotesMeta()
    await fetchReleaseNotesCommits()
  } finally {
    releaseNotesLoading.value = false
  }
}

const openReleaseNotes = async () => {
  releaseNotesOpen.value = true
  if (releaseNotesMonths.value.length > 0) return
  await refreshReleaseNotes()
}

const maybeAutoOpenReleaseNotes = async () => {
  try {
    await fetchReleaseNotesMeta()
    const meta = releaseNotesMeta.value
    if (!meta) return
    if (!meta.autoShowEnabled) return
    if (meta.lastSeenVersion === meta.currentVersion) return
    await openReleaseNotes()
  } catch {
    // silent: release notes should never block app boot
  }
}

const onUpdateReleaseNotesAutoShow = async (enabled: boolean) => {
  releaseNotesAutoShowEnabled.value = enabled
  try {
    await apiFetch('/api/release-notes/auto-show', { method: 'POST', body: { enabled } })
  } catch {
    // best effort
  }
}

const markReleaseNotesSeen = async () => {
  try {
    await apiFetch('/api/release-notes/mark-seen', { method: 'POST' })
    if (releaseNotesMeta.value) {
      releaseNotesMeta.value.lastSeenVersion = releaseNotesMeta.value.currentVersion
    }
  } finally {
    releaseNotesOpen.value = false
  }
}

type NavItem = {
  to: string
  icon: string
  label: string
  adminOnly?: boolean
  platformOnly?: boolean
  module?: 'cra' | 'conges' | 'budget' | 'tma' | 'notifications' | 'billing'
  activePrefix?: string
}

const allNavItems = computed<NavItem[]>(() => [
  { to: '/dashboard', icon: 'dashboard', label: t('nav.dashboard') },
  { to: '/compte', icon: 'person', label: t('nav.profile'), activePrefix: '/compte' },
  { to: '/cra', icon: 'schedule', label: t('nav.cra'), module: 'cra' },
  { to: '/prestations', icon: 'fact_check', label: t('nav.prestations'), module: 'cra', adminOnly: true },
  { to: '/conges', icon: 'beach_access', label: t('nav.conges'), module: 'conges', activePrefix: '/conges' },
  { to: '/budget', icon: 'account_balance', label: t('nav.budget'), module: 'budget' },
  { to: '/tma', icon: 'support_agent', label: t('nav.tma'), module: 'tma' },
  { to: '/platform', icon: 'hub', label: t('nav.platform'), platformOnly: true, activePrefix: '/platform' },
  { to: '/billing/abonnement', icon: 'payments', label: t('nav.billing'), adminOnly: true, module: 'billing' },
  { to: '/admin/notifications', icon: 'notifications', label: t('nav.notifications'), adminOnly: true, module: 'notifications' },
  { to: '/admin/organisation', icon: 'corporate_fare', label: t('nav.organisation'), adminOnly: true },
  { to: '/admin/users', icon: 'group', label: t('nav.users'), adminOnly: true },
  { to: '/admin/parametres', icon: 'settings', label: t('nav.settings'), adminOnly: true, activePrefix: '/admin/parametres' }
])

const navItems = computed(() =>
  allNavItems.value.filter((item) => {
    if (item.platformOnly && !isPlatformAdmin.value) return false
    if (item.adminOnly && !isAdmin.value) return false
    if (item.module && !hasModule(item.module)) return false
    return true
  })
)

const mainNavItems = computed(() =>
  navItems.value.filter(
    (item) =>
      ![
        '/compte',
        '/admin/notifications',
        '/admin/organisation',
        '/admin/users',
        '/admin/parametres',
        '/platform',
        '/billing/abonnement'
      ].includes(item.to)
  )
)

const settingsNavItems = computed(() => {
  const byTo = new Map(navItems.value.map((item) => [item.to, item]))

  return [
    byTo.get('/compte'),
    byTo.get('/admin/notifications'),
    byTo.get('/admin/organisation'),
    byTo.get('/admin/users'),
    byTo.get('/admin/parametres'),
    byTo.get('/platform'),
    byTo.get('/billing/abonnement')
  ].filter((item): item is NavItem => item !== undefined)
})

const isNavActive = (item: NavItem) => {
  const prefix = item.activePrefix ?? item.to
  if (route.path === item.to) return true
  if (prefix !== '/' && route.path.startsWith(`${prefix}/`)) return true
  return false
}

const activeNavItem = computed(() => navItems.value.find((item) => isNavActive(item)))

const bottomNavItems = computed(() => mainNavItems.value.slice(0, 4))

const isNarrowMain = computed(() => route.meta.narrow === true)

const pageTitle = computed(() => {
  const active = activeNavItem.value
  if (active) return active.label

  if (route.path.startsWith('/cra/') && route.params.id) {
    return t('nav.cra')
  }
  if (route.path === '/conges' || route.path.startsWith('/conges/')) {
    return t('nav.conges')
  }
  const item = allNavItems.value.find(
    (n) => route.path === n.to || route.path.startsWith(`${n.to}/`)
  )
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

.sidebar__divider {
  height: 1px;
  background: var(--kore-border);
  margin: var(--kore-space-sm) 0;
}

.sidebar__section-label {
  margin: var(--kore-space-xs) 0;
  font-size: var(--kore-text-caption);
  font-weight: 700;
  color: var(--kore-text-muted);
  letter-spacing: 0.02em;
  text-transform: uppercase;
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
  width: 100%;
  padding: var(--kore-space-xl);
  max-width: var(--kore-app-main-max);
}

.main--narrow {
  max-width: var(--kore-container-max);
}

.past-due {
  margin: 0 0 var(--kore-space-md);
  padding: var(--kore-space-sm) var(--kore-space-md);
  border-radius: var(--kore-radius-md);
  background: rgba(220, 53, 69, 0.12);
  color: var(--kore-error);
  font-size: var(--kore-text-small);
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

.drawer-divider {
  height: 1px;
  background: var(--kore-border);
  margin: var(--kore-space-sm) 0;
}

.drawer-section-label {
  margin: var(--kore-space-xs) 0;
  font-size: var(--kore-text-caption);
  font-weight: 700;
  color: var(--kore-text-muted);
  letter-spacing: 0.02em;
  text-transform: uppercase;
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
