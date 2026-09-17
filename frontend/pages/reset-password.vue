<template>
  <div class="reset-page">
    <PublicCard padding="lg" class="reset-card">
      <h1>{{ $t('reset_password.title') }}</h1>
      <p class="reset-card__subtitle">{{ stepSubtitle }}</p>

      <form v-if="step === 'request'" @submit.prevent="requestReset">
        <PublicInput
          id="reset-email"
          v-model="email"
          type="email"
          :label="$t('reset_password.email')"
          required
        />
        <PublicButton variant="primary" type="submit" class="reset-card__submit" :disabled="busy">
          {{ $t('reset_password.send_link') }}
        </PublicButton>
      </form>

      <form v-else-if="step === 'confirm'" @submit.prevent="confirmReset">
        <PublicInput
          id="new-password"
          v-model="newPassword"
          type="password"
          :label="$t('reset_password.new_password')"
          required
        />
        <PublicInput
          id="confirm-password"
          v-model="confirmPassword"
          type="password"
          :label="$t('reset_password.confirm_password')"
          required
        />
        <p class="reset-card__hint">{{ $t('reset_password.password_hint') }}</p>
        <PublicButton variant="primary" type="submit" class="reset-card__submit" :disabled="busy">
          {{ $t('reset_password.submit') }}
        </PublicButton>
      </form>

      <p v-if="info" class="reset-card__info" role="status">{{ info }}</p>
      <p v-if="error" class="reset-card__error" role="alert">{{ error }}</p>

      <p class="reset-card__back">
        <NuxtLink to="/login" class="reset-card__link">{{ $t('reset_password.back_login') }}</NuxtLink>
      </p>
    </PublicCard>
  </div>
</template>

<script setup lang="ts">
definePageMeta({ layout: 'public' })

const { t } = useI18n()
const route = useRoute()
const config = useRuntimeConfig()
const showMailhogHint = config.public.showMailhogHint as boolean
const mailhogUiUrl = config.public.mailhogUiUrl as string

type ResetStep = 'request' | 'confirm'
const token = computed(() => String(route.query.token || '').trim())
const step = ref<ResetStep>(token.value ? 'confirm' : 'request')

const email = ref('')
const newPassword = ref('')
const confirmPassword = ref('')
const busy = ref(false)
const info = ref('')
const error = ref('')

const stepSubtitle = computed(() => {
  switch (step.value) {
    case 'confirm':
      return t('reset_password.confirm_subtitle')
    case 'request':
      return t('reset_password.subtitle')
    default: {
      const _exhaustive: never = step.value
      return _exhaustive
    }
  }
})

watch(token, (value) => {
  step.value = value ? 'confirm' : 'request'
})

async function requestReset() {
  busy.value = true
  error.value = ''
  info.value = ''
  try {
    await $fetch('/api/auth/password-reset/request', {
      method: 'POST',
      body: { email: email.value }
    })
    info.value = t('reset_password.sent')
    if (showMailhogHint) {
      info.value = `${info.value} ${t('login.discovery_dev_hint')} MailHog (${mailhogUiUrl}).`
    }
  } catch {
    error.value = t('reset_password.error_generic')
  } finally {
    busy.value = false
  }
}

async function confirmReset() {
  busy.value = true
  error.value = ''
  info.value = ''
  if (newPassword.value !== confirmPassword.value) {
    error.value = t('reset_password.error_mismatch')
    busy.value = false
    return
  }
  try {
    await $fetch('/api/auth/password-reset/confirm', {
      method: 'POST',
      body: { token: token.value, newPassword: newPassword.value }
    })
    info.value = t('reset_password.success')
    await navigateTo('/login?reset=ok')
  } catch (e: unknown) {
    const status = (e as { statusCode?: number })?.statusCode
    if (status === 422) {
      error.value = t('reset_password.error_weak')
    } else if (status === 401) {
      error.value = t('reset_password.error_token')
    } else {
      error.value = t('reset_password.error_generic')
    }
  } finally {
    busy.value = false
  }
}
</script>

<style scoped>
.reset-page {
  min-height: 70vh;
  display: flex;
  align-items: center;
  justify-content: center;
  padding: var(--kore-space-xl) var(--kore-space-md);
}

.reset-card {
  width: 100%;
  max-width: var(--kore-form-max);
}

.reset-card h1 {
  margin: 0 0 var(--kore-space-xs);
  font-size: var(--kore-text-h2);
}

.reset-card__subtitle {
  margin: 0 0 var(--kore-space-xl);
  color: var(--kore-text-muted);
  font-size: var(--kore-text-small);
}

form {
  display: flex;
  flex-direction: column;
  gap: var(--kore-space-md);
}

.reset-card__submit {
  width: 100%;
  margin-top: var(--kore-space-sm);
}

.reset-card__hint {
  margin: 0;
  color: var(--kore-text-muted);
  font-size: var(--kore-text-small);
}

.reset-card__info {
  margin: var(--kore-space-md) 0 0;
  color: var(--kore-text-muted);
  font-size: var(--kore-text-small);
  text-align: center;
}

.reset-card__error {
  margin: var(--kore-space-md) 0 0;
  padding: var(--kore-space-sm) var(--kore-space-md);
  color: var(--kore-error);
  font-size: var(--kore-text-small);
  text-align: center;
  background: rgba(248, 113, 113, 0.08);
  border-radius: var(--kore-radius-md);
}

.reset-card__back {
  margin: var(--kore-space-lg) 0 0;
  text-align: center;
  font-size: var(--kore-text-small);
}

.reset-card__link {
  color: var(--kore-link);
  text-decoration: none;
}

@media (max-width: 768px) {
  .reset-page {
    padding: var(--kore-space-lg) var(--kore-space-sm);
    align-items: flex-start;
  }
}
</style>
