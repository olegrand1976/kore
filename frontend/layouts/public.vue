<template>
  <div class="layout-public">
    <header class="header">
      <NuxtLink to="/" class="brand">Kore</NuxtLink>
      <nav>
        <NuxtLink to="/modules">Modules</NuxtLink>
        <NuxtLink to="/tarifs">Tarifs</NuxtLink>
        <NuxtLink to="/reserver">Réserver</NuxtLink>
        <NuxtLink to="/login">Connexion</NuxtLink>
      </nav>
    </header>
    <main>
      <slot />
    </main>
  </div>
</template>

<style scoped>
.layout-public { min-height: 100vh; display: flex; flex-direction: column; }
.header { display: flex; justify-content: space-between; align-items: center; padding: 1rem 2rem; border-bottom: 1px solid #e5e7eb; }
.brand { font-weight: 700; font-size: 1.25rem; text-decoration: none; color: #111827; }
nav { display: flex; gap: 1rem; }
nav a { text-decoration: none; color: #374151; }
main { flex: 1; padding: 2rem; max-width: 1200px; margin: 0 auto; width: 100%; }
</style>
