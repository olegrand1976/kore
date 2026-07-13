<template>
  <div>
    <AppPageHeader :title="displayNameHeader" :subtitle="$t('profile.subtitle')" />

    <AppCard v-if="pending" padding="lg">
      <p class="muted">{{ $t('profile.loading') }}</p>
    </AppCard>

    <AppCard v-else-if="error" padding="lg">
      <AppEmptyState icon="error" :title="$t('profile.load_error')" />
    </AppCard>

    <AppCard v-else-if="!user" padding="lg">
      <AppEmptyState icon="person" :title="$t('profile.not_available')" />
    </AppCard>

    <template v-else>
      <AppCard padding="lg" class="profile-hero">
        <div class="profile-hero__main">
          <div class="profile-hero__avatar" aria-hidden="true">
            <AppIcon name="person" />
          </div>
          <div>
            <h2 class="profile-hero__name">{{ displayName }}</h2>
            <p class="profile-hero__login">{{ user.login }}</p>
            <div class="profile-hero__badges">
              <AppBadge v-if="user.profil" variant="default">{{ user.profil }}</AppBadge>
              <AppBadge v-if="typeof user.active === 'boolean'" :variant="user.active ? 'success' : 'default'">
                {{ user.active ? $t('users.active') : $t('users.inactive') }}
              </AppBadge>
            </div>
          </div>
        </div>
      </AppCard>

      <div class="profile-grid">
        <AppCard padding="lg">
          <h3 class="profile-section-title">{{ $t('profile.section_account') }}</h3>
          <dl class="profile-dl">
            <div>
              <dt>{{ $t('profile.field_first_name') }}</dt>
              <dd>{{ user.prenom || $t('profile.none') }}</dd>
            </div>
            <div>
              <dt>{{ $t('profile.field_last_name') }}</dt>
              <dd>{{ user.nom || $t('profile.none') }}</dd>
            </div>
            <div>
              <dt>{{ $t('profile.field_login') }}</dt>
              <dd>{{ user.login || $t('profile.none') }}</dd>
            </div>
            <div>
              <dt>{{ $t('profile.field_email') }}</dt>
              <dd>{{ user.email || $t('profile.none') }}</dd>
            </div>
            <div>
              <dt>{{ $t('profile.field_account_type') }}</dt>
              <dd>{{ accountTypeLabel(user.typeCompte) }}</dd>
            </div>
            <div>
              <dt>{{ $t('profile.field_lang') }}</dt>
              <dd>{{ user.langue?.toUpperCase() || $t('profile.none') }}</dd>
            </div>
          </dl>
        </AppCard>

        <AppCard padding="lg">
          <h3 class="profile-section-title">{{ $t('profile.section_org') }}</h3>
          <dl class="profile-dl">
            <div>
              <dt>{{ $t('profile.field_team') }}</dt>
              <dd>{{ user.equipeLibelle || $t('profile.none') }}</dd>
            </div>
            <div>
              <dt>{{ $t('profile.field_profile') }}</dt>
              <dd>{{ user.profil || $t('profile.none') }}</dd>
            </div>
          </dl>
        </AppCard>
      </div>
    </template>
  </div>
</template>

<script setup lang="ts">
import { formatUserDisplayName } from '~/composables/useUserDisplay'

definePageMeta({ layout: 'default', narrow: true })

type UserDetail = {
  id?: string
  login?: string
  prenom?: string
  nom?: string
  email?: string
  profil?: string
  active?: boolean
  langue?: string
  typeCompte?: string
  equipeLibelle?: string
}

const { t } = useI18n()
const { user: sessionUser, fetchSession } = useAuth()

await fetchSession()

const sessionUserId = computed(() => String(sessionUser.value?.userId ?? ''))
const canLoad = computed(() => sessionUser.value?.ok === true && sessionUserId.value.length > 0)

const { data, pending, error } = await useFetch<UserDetail | { data?: UserDetail }>(
  () => `/api/org/users/${sessionUserId.value}`,
  { immediate: canLoad }
)

const user = computed(() => {
  if (!canLoad.value) return null
  const payload = (data.value as { data?: UserDetail } | null)?.data ?? data.value
  return payload && typeof payload === 'object' ? (payload as UserDetail) : null
})

const displayName = computed(() =>
  formatUserDisplayName(user.value?.prenom, user.value?.nom, user.value?.login)
)

const displayNameHeader = computed(() => user.value ? displayName.value : t('profile.title'))

const accountTypeLabel = (type?: string) => {
  switch (type) {
    case 'Interne':
      return t('profile.account_type_interne')
    case 'Client':
      return t('profile.account_type_client')
    case 'Prestataire':
      return t('profile.account_type_prestataire')
    default:
      return type || t('profile.none')
  }
}
</script>

<style scoped>
.muted {
  color: var(--kore-text-muted);
}

.profile-hero {
  margin-bottom: var(--kore-space-lg);
}

.profile-hero__main {
  display: flex;
  align-items: center;
  gap: var(--kore-space-lg);
}

.profile-hero__avatar {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 4rem;
  height: 4rem;
  border-radius: var(--kore-radius-full);
  background: var(--kore-bg-subtle);
  border: 1px solid var(--kore-border);
  color: var(--kore-brand-gold);
}

.profile-hero__name {
  margin: 0;
  font-size: var(--kore-text-xl);
  line-height: 1.2;
}

.profile-hero__login {
  margin: var(--kore-space-2xs) 0 0;
  color: var(--kore-text-muted);
}

.profile-hero__badges {
  display: flex;
  flex-wrap: wrap;
  gap: var(--kore-space-xs);
  margin-top: var(--kore-space-sm);
}

.profile-grid {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: var(--kore-space-lg);
}

.profile-section-title {
  margin: 0 0 var(--kore-space-md);
  font-size: var(--kore-text-sm);
  font-weight: 600;
  letter-spacing: 0.02em;
  color: var(--kore-text-muted);
  text-transform: uppercase;
}

.profile-dl {
  display: grid;
  gap: var(--kore-space-md);
  margin: 0;
}

.profile-dl > div {
  display: grid;
  grid-template-columns: 140px 1fr;
  gap: var(--kore-space-sm);
}

.profile-dl dt {
  color: var(--kore-text-muted);
}

.profile-dl dd {
  margin: 0;
  font-weight: 600;
}

@media (max-width: 768px) {
  .profile-grid {
    grid-template-columns: 1fr;
  }
  .profile-dl > div {
    grid-template-columns: 1fr;
    gap: var(--kore-space-2xs);
  }
}
</style>
