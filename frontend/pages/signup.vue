<template>
  <div class="signup-page">
    <PublicPageHero :title="$t('signup.title')" :subtitle="$t('signup.subtitle')" />
    <PublicSection>
      <PublicCard padding="lg" class="signup-card">
        <form class="signup-card__form" @submit.prevent="submit">
          <PublicInput
            id="raison"
            v-model="form.raisonSociale"
            :label="$t('signup.raison_sociale')"
            required
          />
          <div class="signup-card__field">
            <label for="pays" class="signup-card__label">{{ $t('org.country') }}</label>
            <select id="pays" v-model="form.pays" class="signup-card__select" required>
              <option value="FR">{{ $t('org.country_fr') }}</option>
              <option value="BE">{{ $t('org.country_be') }}</option>
              <option value="MG">{{ $t('org.country_mg') }}</option>
              <option value="MA">{{ $t('org.country_ma') }}</option>
              <option value="TN">{{ $t('org.country_tn') }}</option>
              <option value="CA">{{ $t('org.country_ca') }}</option>
            </select>
          </div>
          <PublicInput id="login" v-model="form.adminLogin" :label="$t('signup.admin_login')" placeholder="ADM_monorg" required />
          <PublicInput id="email" v-model="form.adminEmail" type="email" :label="$t('signup.admin_email')" required />
          <PublicInput id="password" v-model="form.adminPassword" type="password" :label="$t('signup.admin_password')" required />
          <p class="signup-card__hint">{{ $t('signup.trial_hint') }}</p>
          <p v-if="error" class="signup-card__error" role="alert">{{ error }}</p>
          <PublicButton variant="primary" type="submit" class="signup-card__submit" :disabled="saving">
            {{ saving ? $t('common.loading') : $t('signup.submit') }}
          </PublicButton>
        </form>
        <p class="signup-card__login">
          {{ $t('signup.have_account') }}
          <NuxtLink to="/login" class="signup-card__link">{{ $t('signup.login_link') }}</NuxtLink>
        </p>
      </PublicCard>
    </PublicSection>
  </div>
</template>

<script setup lang="ts">
definePageMeta({ layout: 'public' })

const { t } = useI18n()
const router = useRouter()
const { extractFetchError } = useApiError()

const form = reactive({
  raisonSociale: '',
  pays: 'FR',
  adminLogin: '',
  adminEmail: '',
  adminPassword: ''
})
const saving = ref(false)
const error = ref('')

async function submit() {
  saving.value = true
  error.value = ''
  try {
    await $fetch('/api/public/signup', {
      method: 'POST',
      body: {
        raisonSociale: form.raisonSociale,
        pays: form.pays,
        adminLogin: form.adminLogin,
        adminEmail: form.adminEmail,
        adminPassword: form.adminPassword
      }
    })
    await router.push({ path: '/login', query: { signup: '1', login: form.adminLogin } })
  } catch (err) {
    error.value = extractFetchError(err)
  } finally {
    saving.value = false
  }
}
</script>

<style scoped>
.signup-page {
  display: flex;
  flex-direction: column;
  gap: var(--kore-space-lg);
}

.signup-card {
  max-width: var(--kore-form-max);
  margin: 0 auto;
  width: 100%;
}

.signup-card__form {
  display: grid;
  gap: var(--kore-space-md);
}

.signup-card__field {
  display: flex;
  flex-direction: column;
  gap: var(--kore-space-xs);
}

.signup-card__label {
  font-size: var(--kore-text-sm);
  color: var(--kore-text-muted);
}

.signup-card__select {
  padding: var(--kore-space-sm) var(--kore-space-md);
  border: 1px solid var(--kore-border);
  border-radius: var(--kore-radius-sm);
  background: var(--kore-surface);
  color: var(--kore-text);
  width: 100%;
}

.signup-card__hint {
  margin: 0;
  color: var(--kore-text-muted);
  font-size: var(--kore-text-sm);
}

.signup-card__error {
  margin: 0;
  color: var(--kore-danger);
  font-size: var(--kore-text-sm);
}

.signup-card__submit {
  width: 100%;
}

.signup-card__login {
  margin: var(--kore-space-md) 0 0;
  text-align: center;
  color: var(--kore-text-muted);
  font-size: var(--kore-text-sm);
}

.signup-card__link {
  color: var(--kore-accent);
  text-decoration: underline;
}

@media (max-width: 768px) {
  .signup-card__submit {
    width: 100%;
  }
}
</style>
