<template>
  <div>
    <h1>Modules Kore</h1>
    <ul>
      <li v-for="mod in modules" :key="mod.code">{{ mod.name }} — {{ mod.description }}</li>
    </ul>
  </div>
</template>

<script setup lang="ts">
definePageMeta({ layout: 'public' })
const { data } = await useFetch('/api/public/modules')
const modules = computed(() => data.value?.data ?? [])
</script>
