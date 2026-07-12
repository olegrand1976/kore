<script setup lang="ts">
withDefaults(defineProps<{
  hoverable?: boolean
  padding?: 'sm' | 'md' | 'lg'
}>(), {
  hoverable: false,
  padding: 'md'
})
</script>

<template>
  <div
    class="pub-card"
    :class="[
      hoverable && 'pub-card--hoverable',
      `pub-card--${padding}`
    ]"
  >
    <slot />
  </div>
</template>

<style scoped>
.pub-card {
  background: var(--kore-bg-elevated);
  border: 1px solid var(--kore-border);
  border-radius: var(--kore-radius-lg);
  box-shadow: var(--kore-shadow-sm);
}

.pub-card--sm { padding: var(--kore-space-md); }
.pub-card--md { padding: var(--kore-space-lg); }
.pub-card--lg { padding: var(--kore-space-xl); }

.pub-card--hoverable {
  transition: box-shadow 0.15s, border-color 0.15s;
}

.pub-card--hoverable:hover {
  box-shadow: var(--kore-shadow-md);
  border-color: var(--kore-brand-gold);
}
</style>
