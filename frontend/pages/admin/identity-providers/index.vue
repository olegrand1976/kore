<template>
  <div class="idp-page">
    <AppPageHeader :title="$t('idp.title')" :subtitle="$t('idp.subtitle')" />
    <AppCard padding="lg">
      <form class="idp-form" @submit.prevent="save">
        <AppInput id="idp-name" v-model="form.name" :label="$t('idp.name')" required />
        <AppInput id="idp-issuer" v-model="form.issuer" :label="$t('idp.issuer')" placeholder="https://login.microsoftonline.com/..." required />
        <AppInput id="idp-client-id" v-model="form.clientId" :label="$t('idp.client_id')" required />
        <AppInput id="idp-client-secret" v-model="form.clientSecret" type="password" :label="$t('idp.client_secret')" />
        <AppInput id="idp-jwks" v-model="form.jwksUri" :label="$t('idp.jwks_uri')" />
        <AppInput id="idp-scopes" v-model="form.scopes" :label="$t('idp.scopes')" />
        <label class="idp-form__toggle">
          <input v-model="form.enabled" type="checkbox" />
          {{ $t('idp.enabled') }}
        </label>
        <AppButton type="submit" variant="primary" :loading="pending">{{ $t('common.save') }}</AppButton>
        <p v-if="message" class="idp-form__msg">{{ message }}</p>
      </form>
    </AppCard>
  </div>
</template>

<script setup lang="ts">
definePageMeta({ layout: 'default', middleware: 'admin' })

const { t } = useI18n()
const pending = ref(false)
const message = ref('')
const idpId = ref(crypto.randomUUID())

const form = reactive({
  name: 'Azure AD',
  issuer: '',
  clientId: '',
  clientSecret: '',
  jwksUri: '',
  scopes: 'openid profile email',
  defaultProfile: 'Collaborateur',
  enabled: false
})

const { data } = await useFetch<Array<Record<string, unknown>>>('/api/admin/identity-providers')
watch(data, (items) => {
  const first = items?.[0]
  if (!first) return
  idpId.value = String(first.id)
  form.name = String(first.name ?? form.name)
  form.issuer = String(first.issuer ?? '')
  form.clientId = String(first.clientId ?? '')
  form.jwksUri = String(first.jwksUri ?? '')
  form.scopes = String(first.scopes ?? form.scopes)
  form.enabled = Boolean(first.enabled)
}, { immediate: true })

const save = async () => {
  pending.value = true
  message.value = ''
  try {
    await $fetch(`/api/admin/identity-providers/${idpId.value}`, {
      method: 'PUT',
      body: { ...form }
    })
    message.value = t('idp.saved')
  } catch {
    message.value = t('common.error')
  } finally {
    pending.value = false
  }
}
</script>

<style scoped>
.idp-form {
  display: grid;
  gap: var(--kore-space-md);
  max-width: var(--kore-form-wide-max);
}

.idp-form__toggle {
  display: flex;
  align-items: center;
  gap: var(--kore-space-sm);
}

.idp-form__msg {
  color: var(--kore-text-muted);
  font-size: var(--kore-text-small);
}
</style>
