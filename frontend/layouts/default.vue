<template>
  <div class="layout-app">
    <aside class="sidebar">
      <div class="brand">Kore</div>
      <nav>
        <NuxtLink to="/dashboard"><AppIcon name="dashboard" /> Dashboard</NuxtLink>
        <NuxtLink to="/cra"><AppIcon name="schedule" /> CRA</NuxtLink>
        <NuxtLink to="/admin/organisation"><AppIcon name="corporate_fare" /> Organisation</NuxtLink>
      </nav>
    </aside>
    <div class="content">
      <header class="topbar">
        <span>Application Kore</span>
        <button type="button" @click="logout">Déconnexion</button>
      </header>
      <main><slot /></main>
    </div>
  </div>
</template>

<script setup lang="ts">
const logout = async () => {
  await $fetch('/api/auth/logout', { method: 'POST' })
  await navigateTo('/login')
}
</script>

<style scoped>
.layout-app { display: grid; grid-template-columns: 240px 1fr; min-height: 100vh; }
.sidebar { background: #111827; color: white; padding: 1rem; }
.brand { font-weight: 700; margin-bottom: 1.5rem; }
.sidebar nav { display: flex; flex-direction: column; gap: 0.5rem; }
.sidebar a { color: #d1d5db; text-decoration: none; display: flex; align-items: center; gap: 0.5rem; }
.content { display: flex; flex-direction: column; }
.topbar { display: flex; justify-content: space-between; padding: 1rem 1.5rem; border-bottom: 1px solid #e5e7eb; }
main { padding: 1.5rem; }
</style>
