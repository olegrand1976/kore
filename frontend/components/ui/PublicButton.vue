<script setup lang="ts">
export type ButtonVariant = 'primary' | 'secondary' | 'ghost' | 'danger'

withDefaults(defineProps<{
  variant?: ButtonVariant
  type?: 'button' | 'submit' | 'reset'
  disabled?: boolean
  to?: string
}>(), {
  variant: 'primary',
  type: 'button',
  disabled: false
})
</script>

<template>
  <NuxtLink
    v-if="to"
    :to="to"
    class="pub-btn"
    :class="[`pub-btn--${variant}`]"
  >
    <slot />
  </NuxtLink>
  <button
    v-else
    :type="type"
    class="pub-btn"
    :class="[`pub-btn--${variant}`]"
    :disabled="disabled"
  >
    <slot />
  </button>
</template>

<style scoped>
.pub-btn {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  gap: var(--kore-space-sm);
  padding: 0.75rem 1.5rem;
  font-family: var(--kore-font);
  font-size: var(--kore-text-small);
  font-weight: 600;
  text-decoration: none;
  border-radius: var(--kore-radius-md);
  border: 1px solid transparent;
  cursor: pointer;
  transition: background 0.15s, border-color 0.15s, color 0.15s, box-shadow 0.15s;
}

.pub-btn:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}

.pub-btn--primary {
  background: var(--kore-brand-gold);
  color: var(--kore-text-inverse);
  border-color: var(--kore-brand-gold);
}

.pub-btn--primary:hover:not(:disabled) {
  background: var(--kore-brand-gold-light);
  border-color: var(--kore-brand-gold-light);
  box-shadow: var(--kore-gold-glow);
}

.pub-btn--secondary {
  background: transparent;
  color: var(--kore-brand-gold);
  border-color: var(--kore-brand-gold);
}

.pub-btn--secondary:hover:not(:disabled) {
  background: rgba(201, 162, 39, 0.1);
}

.pub-btn--ghost {
  background: transparent;
  color: var(--kore-text-muted);
  border-color: transparent;
}

.pub-btn--ghost:hover:not(:disabled) {
  color: var(--kore-text);
  background: var(--kore-bg-subtle);
}

.pub-btn--danger {
  background: var(--kore-error);
  color: #fff;
  border-color: var(--kore-error);
}
</style>
