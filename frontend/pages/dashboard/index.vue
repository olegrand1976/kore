<template>
  <div>
    <h1>Dashboard</h1>
    <p>Bienvenue dans l'application Kore.</p>
  </div>
</template>

<script setup lang="ts">
definePageMeta({ layout: 'default', middleware: ['auth'] })
</script>
