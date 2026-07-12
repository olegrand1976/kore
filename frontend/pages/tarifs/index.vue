<template>
  <div>
    <h1>Tarifs</h1>
    <pre>{{ pricing }}</pre>
  </div>
</template>

<script setup lang="ts">
definePageMeta({ layout: 'public' })
const { data } = await useFetch('/api/public/pricing')
const pricing = computed(() => data.value?.data ?? {})
</script>
