<template>
  <div class="hero">
    <h1>Kore — PSA/ESN modulaire</h1>
    <p>CRA pivot, TMA, congés, budget UO, abonnement SaaS et conformité européenne.</p>
    <div class="actions">
      <NuxtLink class="btn primary" to="/tarifs">Voir les tarifs</NuxtLink>
      <NuxtLink class="btn" to="/reserver">Réserver un entretien</NuxtLink>
    </div>
    <p v-if="health" class="status">API : {{ health.status }}</p>
  </div>
</template>

<script setup lang="ts">
definePageMeta({ layout: 'public' })

const { data: health } = await useFetch('/api/health')
</script>

<style scoped>
.hero { text-align: center; padding: 4rem 1rem; }
h1 { font-size: 2.5rem; margin-bottom: 1rem; }
.actions { display: flex; gap: 1rem; justify-content: center; margin-top: 2rem; }
.btn { padding: 0.75rem 1.5rem; border-radius: 0.5rem; text-decoration: none; border: 1px solid #d1d5db; color: #111827; }
.btn.primary { background: #2563eb; color: white; border-color: #2563eb; }
.status { margin-top: 2rem; color: #6b7280; font-size: 0.875rem; }
</style>
