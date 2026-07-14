<template>
  <div class="login-page">
    <div class="login-page__panel login-page__panel--brand">
      <KoreLogo variant="horizontal" size="md" tone="auto" alt="Kore" />
      <p class="login-page__baseline">{{ $t('brand.tagline') }}</p>
      <ul class="login-page__features">
        <li><AppIcon name="check_circle" /> CRA, TMA & Budget unifiés</li>
        <li><AppIcon name="check_circle" /> Conformité européenne</li>
        <li><AppIcon name="check_circle" /> Déploiement modulaire</li>
      </ul>
    </div>
    <PublicCard padding="lg" class="login-card">
      <h1>{{ stepTitle }}</h1>
      <p class="login-card__subtitle">{{ stepSubtitle }}</p>

      <form v-if="step === 'credentials'" @submit.prevent="submit">
        <PublicInput id="login" v-model="login" :label="$t('login.identifier')" placeholder="ADM_admin" required />
        <PublicInput id="password" v-model="password" type="password" :label="$t('login.password')" required />
        <PublicButton variant="primary" type="submit" class="login-card__submit">{{ $t('login.submit') }}</PublicButton>
      </form>

      <form v-else-if="step === 'totp'" @submit.prevent="submitTotp">
        <PublicInput
          id="totp-code"
          v-model="totpCode"
          :label="$t('login.2fa.code_label')"
          inputmode="numeric"
          autocomplete="one-time-code"
          maxlength="8"
          required
        />
        <PublicButton variant="primary" type="submit" class="login-card__submit">{{ $t('login.2fa.verify') }}</PublicButton>
        <button type="button" class="login-card__link" @click="step = 'credentials'">{{ $t('login.2fa.back') }}</button>
      </form>

      <div v-else-if="step === 'enrollment'" class="login-enrollment">
        <p class="login-card__info">{{ $t('login.2fa.enrollment_intro') }}</p>
        <img v-if="qrCodeDataUrl" :src="qrCodeDataUrl" width="200" height="200" class="login-enrollment__qr" :alt="$t('login.2fa.enrollment_subtitle')" />
        <p v-if="manualSecret" class="login-enrollment__secret">
          <span class="login-enrollment__secret-label">{{ $t('profile.2fa.manual_secret') }}</span>
          <code>{{ manualSecret }}</code>
        </p>
        <form @submit.prevent="submitEnrollment">
          <PublicInput
            id="enroll-code"
            v-model="totpCode"
            :label="$t('login.2fa.code_label')"
            inputmode="numeric"
            autocomplete="one-time-code"
            maxlength="8"
            required
          />
          <PublicButton variant="primary" type="submit" class="login-card__submit">{{ $t('login.2fa.enrollment_confirm') }}</PublicButton>
        </form>
        <div v-if="backupCodes.length" class="login-enrollment__backup">
          <p>{{ $t('profile.2fa.backup_title') }}</p>
          <ul>
            <li v-for="code in backupCodes" :key="code"><code>{{ code }}</code></li>
          </ul>
          <PublicButton variant="primary" class="login-card__submit" @click="finishEnrollment">{{ $t('profile.2fa.continue') }}</PublicButton>
        </div>
      </div>

      <template v-if="step === 'credentials'">
        <button type="button" class="login-card__link" @click="showDiscovery = !showDiscovery">
          {{ $t('login.find_org') }}
        </button>
        <div v-if="showDiscovery" class="login-card__discovery">
          <PublicInput id="discover-email" v-model="discoverEmail" type="email" :label="$t('login.email')" />
          <PublicButton variant="ghost" class="login-card__submit" @click="requestDiscovery">
            {{ $t('login.send_link') }}
          </PublicButton>
          <p v-if="discoveryInfo" class="login-card__info" role="status">{{ discoveryInfo }}</p>
          <p v-if="discoveryInfo && showMailhogHint" class="login-card__info login-card__info--dev" role="note">
            {{ $t('login.discovery_dev_hint') }}
            <a :href="mailhogUiUrl" target="_blank" rel="noopener noreferrer" class="login-card__link-inline">MailHog</a>
          </p>
        </div>
        <div v-if="showSso" class="login-card__sso">
          <p class="login-card__divider" aria-hidden="true">{{ $t('login.or_divider') }}</p>
          <PublicButton variant="secondary" class="login-card__submit" @click="startSSO">{{ ssoButtonLabel }}</PublicButton>
        </div>
      </template>
      <p v-if="error" class="login-card__error" role="alert">{{ error }}</p>
    </PublicCard>
  </div>
</template>

<script setup lang="ts">
definePageMeta({ layout: 'public' })

const { t } = useI18n()
const config = useRuntimeConfig()
const mailhogUiUrl = config.public.mailhogUiUrl as string
const showMailhogHint = config.public.showMailhogHint as boolean
const login = ref('ADM_admin')
const password = ref('Admin123!')
const tenantId = ref('')
const showSso = ref(false)
const ssoProviderName = ref('')
const error = ref('')
const showDiscovery = ref(false)
const discoverEmail = ref('')
const discoveryInfo = ref('')

type LoginStep = 'credentials' | 'totp' | 'enrollment'
const step = ref<LoginStep>('credentials')
const challengeToken = ref('')
const enrollmentToken = ref('')

const {
  manualSecret,
  totpCode,
  backupCodes,
  qrCodeDataUrl,
  loadSetup,
  reset: resetTotpSetup
} = useTwoFactorSetup()

const stepTitle = computed(() => {
  switch (step.value) {
    case 'totp':
      return t('login.2fa.title')
    case 'enrollment':
      return t('login.2fa.enrollment_title')
    default:
      return t('login.title')
  }
})

const stepSubtitle = computed(() => {
  switch (step.value) {
    case 'totp':
      return t('login.2fa.subtitle')
    case 'enrollment':
      return t('login.2fa.enrollment_subtitle')
    default:
      return t('login.subtitle')
  }
})

type LoginData = {
  requires2FA?: boolean
  Requires2FA?: boolean
  requires2FAEnrollment?: boolean
  Requires2FAEnrollment?: boolean
  challengeToken?: string
  ChallengeToken?: string
  enrollmentToken?: string
  EnrollmentToken?: string
}

const ssoButtonLabel = computed(() => {
  if (ssoProviderName.value) {
    return t('login.sso_continue', { provider: ssoProviderName.value })
  }
  return t('login.sso_continue_default')
})

async function checkSsoAvailability(tenant: string) {
  try {
    const res = await $fetch<{ data?: { enabled?: boolean; providerName?: string } }>('/api/auth/oidc/status', {
      query: { tenant }
    })
    const enabled = Boolean(res?.data?.enabled)
    showSso.value = enabled
    ssoProviderName.value = res?.data?.providerName?.trim() ?? ''
  } catch {
    showSso.value = false
    ssoProviderName.value = ''
  }
}

async function resolveTenant(tenant: string) {
  tenantId.value = tenant
  await checkSsoAvailability(tenant)
}

const submit = async () => {
  error.value = ''
  try {
    const res = await $fetch<{ data?: LoginData }>('/api/auth/login', {
      method: 'POST',
      body: { login: login.value, password: password.value }
    })
    const data = res?.data
    if (data?.requires2FA || data?.Requires2FA) {
      step.value = 'totp'
      challengeToken.value = data.challengeToken ?? data.ChallengeToken ?? ''
      return
    }
    if (data?.requires2FAEnrollment || data?.Requires2FAEnrollment) {
      step.value = 'enrollment'
      enrollmentToken.value = data.enrollmentToken ?? data.EnrollmentToken ?? ''
      await startEnrollmentSetup()
      return
    }
    await navigateTo('/dashboard')
  } catch (e: unknown) {
    const err = e as { data?: { error?: { message?: string } } }
    error.value = err?.data?.error?.message || t('login.error')
  }
}

const submitTotp = async () => {
  error.value = ''
  try {
    await $fetch('/api/auth/2fa/verify', {
      method: 'POST',
      body: { challengeToken: challengeToken.value, code: totpCode.value }
    })
    await navigateTo('/dashboard')
  } catch (e: unknown) {
    const err = e as { data?: { error?: { message?: string } } }
    error.value = err?.data?.error?.message || t('login.2fa.error')
  }
}

const startEnrollmentSetup = async () => {
  resetTotpSetup()
  try {
    await loadSetup('/api/auth/2fa/enrollment/setup', {
      body: { enrollmentToken: enrollmentToken.value }
    })
  } catch {
    error.value = t('login.2fa.setup_error')
  }
}

const submitEnrollment = async () => {
  error.value = ''
  try {
    const res = await $fetch<{ data?: { backupCodes?: string[] } }>('/api/auth/2fa/enrollment/confirm', {
      method: 'POST',
      body: { enrollmentToken: enrollmentToken.value, code: totpCode.value }
    })
    backupCodes.value = res?.data?.backupCodes ?? []
    if (backupCodes.value.length === 0) {
      await navigateTo('/dashboard')
    }
  } catch (e: unknown) {
    const err = e as { data?: { error?: { message?: string } } }
    error.value = err?.data?.error?.message || t('login.2fa.error')
  }
}

const finishEnrollment = async () => {
  await navigateTo('/dashboard')
}

function randomVerifier(): string {
  const arr = new Uint8Array(32)
  crypto.getRandomValues(arr)
  return btoa(String.fromCharCode(...arr)).replace(/\+/g, '-').replace(/\//g, '_').replace(/=+$/, '')
}

async function sha256Base64Url(input: string): Promise<string> {
  const data = new TextEncoder().encode(input)
  const hash = await crypto.subtle.digest('SHA-256', data)
  return btoa(String.fromCharCode(...new Uint8Array(hash))).replace(/\+/g, '-').replace(/\//g, '_').replace(/=+$/, '')
}

const startSSO = async () => {
  error.value = ''
  if (!tenantId.value) {
    error.value = t('login.sso_unavailable')
    return
  }
  try {
    const verifier = randomVerifier()
    const challenge = await sha256Base64Url(verifier)
    const redirectUri = `${window.location.origin}/login`
    sessionStorage.setItem('oidc_verifier', verifier)
    sessionStorage.setItem('oidc_tenant', tenantId.value)
    sessionStorage.setItem('oidc_redirect', redirectUri)
    const res = await $fetch<{ data?: { authorizeUrl?: string } }>('/api/auth/oidc/authorize', {
      query: {
        tenant: tenantId.value,
        redirect_uri: redirectUri,
        code_challenge: challenge
      }
    })
    const url = res?.data?.authorizeUrl
    if (url) window.location.href = url
  } catch (e: unknown) {
    const err = e as { data?: { error?: { message?: string } } }
    error.value = err?.data?.error?.message || t('login.error')
  }
}

const requestDiscovery = async () => {
  error.value = ''
  discoveryInfo.value = ''
  try {
    await $fetch('/api/auth/tenant-discovery/request', { method: 'POST', body: { email: discoverEmail.value } })
    discoveryInfo.value = t('login.discovery_sent')
  } catch (e: unknown) {
    const err = e as { data?: { error?: { message?: string } } }
    error.value = err?.data?.error?.message || t('login.error')
  }
}

onMounted(async () => {
  const params = new URLSearchParams(window.location.search)
  const inviteToken = params.get('invite')
  const discoverToken = params.get('discover')
  if (inviteToken) {
    try {
      const res = await $fetch<{ data?: { tenantId?: string } }>('/api/public/invitations/resolve', {
        query: { token: inviteToken }
      })
      const resolved = res?.data?.tenantId
      if (resolved) await resolveTenant(resolved)
    } catch (e: unknown) {
      // ignore
    }
  } else if (discoverToken) {
    try {
      const res = await $fetch<{ data?: { tenantId?: string } }>('/api/auth/tenant-discovery/resolve', {
        query: { token: discoverToken }
      })
      const resolved = res?.data?.tenantId
      if (resolved) await resolveTenant(resolved)
    } catch (e: unknown) {
      // ignore
    }
  }
  const code = params.get('code')
  const state = params.get('state')
  if (!code || !state) return
  const verifier = sessionStorage.getItem('oidc_verifier')
  const tenant = sessionStorage.getItem('oidc_tenant')
  const redirectUri = sessionStorage.getItem('oidc_redirect')
  if (!verifier || !tenant || !redirectUri) return
  const handledKey = `oidc_callback_handled:${state}`
  if (sessionStorage.getItem(handledKey) === '1') {
    return
  }
  sessionStorage.setItem(handledKey, '1')
  try {
    await $fetch('/api/auth/oidc/callback', {
      method: 'POST',
      body: { tenantId: tenant, code, state, codeVerifier: verifier, redirectUri }
    })
    sessionStorage.removeItem('oidc_verifier')
    sessionStorage.removeItem('oidc_tenant')
    sessionStorage.removeItem('oidc_redirect')
    await navigateTo('/dashboard')
  } catch (e: unknown) {
    const err = e as { data?: { error?: { message?: string } } }
    error.value = err?.data?.error?.message || t('login.error')
  }
})
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
    grid-template-columns: 1fr var(--kore-form-max);
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

.login-card__link {
  border: 0;
  background: transparent;
  padding: var(--kore-space-md) 0 0;
  font: inherit;
  color: var(--kore-link);
  text-align: center;
  width: 100%;
  cursor: pointer;
}

.login-card__sso {
  margin-top: var(--kore-space-md);
  display: flex;
  flex-direction: column;
  gap: var(--kore-space-sm);
}

.login-card__divider {
  margin: 0;
  color: var(--kore-text-muted);
  font-size: var(--kore-text-small);
  text-align: center;
}

.login-card__discovery {
  margin-top: var(--kore-space-md);
  display: grid;
  gap: var(--kore-space-sm);
}

.login-card__info {
  margin: 0;
  color: var(--kore-text-muted);
  font-size: var(--kore-text-small);
  text-align: center;
}

.login-card__info--dev {
  color: var(--kore-brand-gold);
}

.login-card__link-inline {
  color: var(--kore-link);
  font-weight: 500;
}

.login-enrollment {
  display: grid;
  gap: var(--kore-space-md);
}

.login-enrollment__qr {
  margin: 0 auto;
  border-radius: var(--kore-radius-md);
}

.login-enrollment__secret {
  margin: 0;
  font-size: var(--kore-text-small);
  text-align: center;
  word-break: break-all;
}

.login-enrollment__secret-label {
  display: block;
  color: var(--kore-text-muted);
  margin-bottom: var(--kore-space-2xs);
}

.login-enrollment__backup ul {
  list-style: none;
  margin: 0;
  padding: 0;
  display: grid;
  gap: var(--kore-space-xs);
}

.login-enrollment__backup code {
  font-family: var(--kore-font-mono, monospace);
}
</style>
