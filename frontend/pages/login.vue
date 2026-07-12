<template>
  <div class="login">
    <h1>Connexion</h1>
    <form @submit.prevent="submit">
      <input v-model="login" placeholder="Login (ex. ADM_admin)" required />
      <input v-model="password" type="password" placeholder="Mot de passe" required />
      <button type="submit">Se connecter</button>
    </form>
    <p v-if="error" class="error">{{ error }}</p>
  </div>
</template>

<script setup lang="ts">
definePageMeta({ layout: 'public' })
const login = ref('ADM_admin')
const password = ref('Admin123!')
const error = ref('')

const submit = async () => {
  error.value = ''
  try {
    await $fetch('/api/auth/login', { method: 'POST', body: { login: login.value, password: password.value } })
    await navigateTo('/dashboard')
  } catch (e: any) {
    error.value = e?.data?.error?.message || 'Échec de connexion'
  }
}
</script>

<style scoped>
.login { max-width: 400px; margin: 2rem auto; }
form { display: flex; flex-direction: column; gap: 0.75rem; }
input, button { padding: 0.75rem; }
.error { color: #dc2626; }
</style>
