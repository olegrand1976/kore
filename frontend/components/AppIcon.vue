<template>
  <span
    class="material-symbols-outlined app-icon"
    :aria-hidden="label ? undefined : 'true'"
    :aria-label="label || undefined"
    role="img"
  >{{ name }}</span>
</template>

<script setup lang="ts">
defineProps<{ name: string; label?: string }>()
</script>

<style scoped>
.app-icon {
  font-family: 'Material Symbols Outlined', sans-serif;
  font-size: 1.25rem;
  vertical-align: middle;
  line-height: 1;
  font-variation-settings: 'FILL' 0, 'wght' 400, 'GRAD' 0, 'opsz' 24;
}
</style>
