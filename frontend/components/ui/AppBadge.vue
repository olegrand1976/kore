<script setup lang="ts">
export type BadgeVariant = 'default' | 'success' | 'warning' | 'error' | 'gold'

defineProps<{ variant?: BadgeVariant }>()
</script>

<template>
  <span class="app-badge" :class="[`app-badge--${variant || 'default'}`]">
    <slot />
  </span>
</template>

<style scoped>
.app-badge {
  display: inline-flex;
  align-items: center;
  padding: 0.125rem 0.5rem;
  font-size: var(--kore-text-caption);
  font-weight: 600;
  border-radius: var(--kore-radius-full);
  text-transform: uppercase;
  letter-spacing: 0.04em;
}

.app-badge--default {
  background: var(--kore-bg-subtle);
  color: var(--kore-text-muted);
}

.app-badge--success {
  background: rgba(74, 222, 128, 0.15);
  color: var(--kore-success);
}

.app-badge--warning {
  background: rgba(201, 162, 39, 0.15);
  color: var(--kore-brand-gold);
}

.app-badge--error {
  background: rgba(248, 113, 113, 0.15);
  color: var(--kore-error);
}

.app-badge--gold {
  background: var(--kore-brand-gold);
  color: var(--kore-text-inverse);
}
</style>
