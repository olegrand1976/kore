<script setup lang="ts">
export type ButtonVariant = 'primary' | 'secondary' | 'ghost' | 'danger'

withDefaults(defineProps<{
  variant?: ButtonVariant
  type?: 'button' | 'submit' | 'reset'
  disabled?: boolean
  to?: string
  size?: 'sm' | 'md'
}>(), {
  variant: 'primary',
  type: 'button',
  disabled: false,
  size: 'md'
})
</script>

<template>
  <NuxtLink v-if="to" :to="to" class="app-btn" :class="[`app-btn--${variant}`, `app-btn--${size}`]">
    <slot />
  </NuxtLink>
  <button
    v-else
    :type="type"
    class="app-btn"
    :class="[`app-btn--${variant}`, `app-btn--${size}`]"
    :disabled="disabled"
  >
    <slot />
  </button>
</template>

<style scoped>
.app-btn {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  gap: var(--kore-space-sm);
  font-family: var(--kore-font);
  font-weight: 600;
  text-decoration: none;
  border-radius: var(--kore-radius-md);
  border: 1px solid transparent;
  cursor: pointer;
  transition: background 0.15s, border-color 0.15s, color 0.15s;
}

.app-btn--md { padding: 0.625rem 1.25rem; font-size: var(--kore-text-small); }
.app-btn--sm { padding: 0.375rem 0.75rem; font-size: var(--kore-text-caption); }

.app-btn:disabled { opacity: 0.5; cursor: not-allowed; }

.app-btn--primary {
  background: var(--kore-brand-gold);
  color: var(--kore-text-inverse);
  border-color: var(--kore-brand-gold);
}

.app-btn--primary:hover:not(:disabled) {
  background: var(--kore-brand-gold-light);
}

.app-btn--secondary {
  background: transparent;
  color: var(--kore-brand-gold);
  border-color: var(--kore-brand-gold);
}

.app-btn--ghost {
  background: transparent;
  color: var(--kore-text-muted);
  border-color: transparent;
}

.app-btn--ghost:hover:not(:disabled) {
  color: var(--kore-text);
  background: var(--kore-bg-subtle);
}

.app-btn--danger {
  background: var(--kore-error);
  color: #fff;
}
</style>
