<template>
  <div class="kpi-grid" :class="{ 'kpi-grid--compact': compact }">
    <slot />
  </div>
</template>

<script setup lang="ts">
withDefaults(defineProps<{ compact?: boolean }>(), { compact: false })
</script>

<style scoped>
.kpi-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(160px, 1fr));
  gap: var(--kore-space-md);
  margin-bottom: var(--kore-space-lg);
}

.kpi-grid--compact {
  grid-template-columns: repeat(auto-fill, minmax(140px, 1fr));
  gap: var(--kore-space-sm);
  margin-bottom: var(--kore-space-md);
}

@media (max-width: 640px) {
  .kpi-grid {
    grid-template-columns: repeat(2, minmax(0, 1fr));
  }
}
</style>
