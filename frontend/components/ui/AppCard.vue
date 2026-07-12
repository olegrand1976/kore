<script setup lang="ts">
withDefaults(defineProps<{ hoverable?: boolean; padding?: 'none' | 'sm' | 'md' | 'lg' }>(), {
  hoverable: false,
  padding: 'md'
})
</script>

<template>
  <div class="app-card" :class="[hoverable && 'app-card--hoverable', `app-card--${padding}`]">
    <slot />
  </div>
</template>

<style scoped>
.app-card {
  background: var(--kore-bg-elevated);
  border: 1px solid var(--kore-border);
  border-radius: var(--kore-radius-lg);
  box-shadow: var(--kore-shadow-sm);
}

.app-card--none { padding: 0; }
.app-card--sm { padding: var(--kore-space-md); }
.app-card--md { padding: var(--kore-space-lg); }
.app-card--lg { padding: var(--kore-space-xl); }

.app-card--hoverable {
  transition: box-shadow 0.15s, border-color 0.15s;
}

.app-card--hoverable:hover {
  box-shadow: var(--kore-shadow-md);
  border-color: var(--kore-brand-gold);
}
</style>
