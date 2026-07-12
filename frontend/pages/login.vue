<template>
  <div class="login-page">
    <div class="login-page__panel login-page__panel--brand">
      <KoreLogo variant="horizontal" size="md" />
      <p class="login-page__baseline">{{ $t('brand.tagline') }}</p>
      <ul class="login-page__features">
        <li><AppIcon name="check_circle" /> CRA, TMA & Budget unifiés</li>
        <li><AppIcon name="check_circle" /> Conformité européenne</li>
        <li><AppIcon name="check_circle" /> Déploiement modulaire</li>
      </ul>
    </div>
    <PublicCard padding="lg" class="login-card">
      <h1>{{ $t('login.title') }}</h1>
      <p class="login-card__subtitle">{{ $t('login.subtitle') }}</p>
      <form @submit.prevent="submit">
        <PublicInput id="login" v-model="login" :label="$t('nav.login')" placeholder="ADM_admin" required />
        <PublicInput id="password" v-model="password" type="password" :label="$t('login.password')" required />
        <PublicButton variant="primary" type="submit" class="login-card__submit">{{ $t('login.submit') }}</PublicButton>
      </form>
      <p v-if="error" class="login-card__error" role="alert">{{ error }}</p>
    </PublicCard>
  </div>
</template>

<script setup lang="ts">
definePageMeta({ layout: 'public' })

const { t } = useI18n()
const login = ref('ADM_admin')
const password = ref('Admin123!')
const error = ref('')

const submit = async () => {
  error.value = ''
  try {
    await $fetch('/api/auth/login', { method: 'POST', body: { login: login.value, password: password.value } })
    await navigateTo('/dashboard')
  } catch (e: unknown) {
    const err = e as { data?: { error?: { message?: string } } }
    error.value = err?.data?.error?.message || t('login.error')
  }
}
</script>

<style scoped>
.login-page {
  display: grid;
  gap: var(--kore-space-xl);
  align-items: stretch;
  min-height: calc(100vh - 8rem);
  padding: var(--kore-space-xl) 0 var(--kore-space-2xl);
}

@media (min-width: 900px) {
  .login-page {
    grid-template-columns: 1fr 420px;
    gap: var(--kore-space-2xl);
    align-items: center;
  }
}

.login-page__panel--brand {
  display: none;
  flex-direction: column;
  justify-content: center;
  padding: var(--kore-space-2xl);
  border-radius: var(--kore-radius-lg);
  background: var(--kore-hero-gradient);
  border: 1px solid var(--kore-border);
}

@media (min-width: 900px) {
  .login-page__panel--brand { display: flex; }
}

.login-page__baseline {
  margin: var(--kore-space-lg) 0 var(--kore-space-xl);
  font-size: var(--kore-text-small);
  letter-spacing: 0.1em;
  text-transform: uppercase;
  color: var(--kore-brand-gold);
  font-weight: 500;
}

.login-page__features {
  list-style: none;
  margin: 0;
  padding: 0;
  display: flex;
  flex-direction: column;
  gap: var(--kore-space-md);
}

.login-page__features li {
  display: flex;
  align-items: center;
  gap: var(--kore-space-sm);
  color: var(--kore-text-muted);
  font-size: var(--kore-text-small);
}

.login-page__features :deep(.material-symbols-outlined) {
  color: var(--kore-brand-gold);
  font-size: 1.125rem !important;
}

.login-card { width: 100%; }

.login-card h1 {
  margin: 0 0 var(--kore-space-xs);
  font-size: var(--kore-text-h2);
}

.login-card__subtitle {
  margin: 0 0 var(--kore-space-xl);
  color: var(--kore-text-muted);
  font-size: var(--kore-text-small);
}

form { display: flex; flex-direction: column; gap: var(--kore-space-md); }

.login-card__submit { width: 100%; margin-top: var(--kore-space-sm); }

.login-card__error {
  margin: var(--kore-space-md) 0 0;
  padding: var(--kore-space-sm) var(--kore-space-md);
  color: var(--kore-error);
  font-size: var(--kore-text-small);
  text-align: center;
  background: rgba(248, 113, 113, 0.08);
  border-radius: var(--kore-radius-md);
}
</style>
