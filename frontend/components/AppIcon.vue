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
  font-size: 1.25rem;
  vertical-align: middle;
  line-height: 1;
}
</style>
