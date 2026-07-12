<script setup lang="ts">
const { theme, toggleTheme } = useTheme()

const label = computed(() =>
  theme.value === 'dark' ? 'Activer le mode clair' : 'Activer le mode sombre'
)
</script>

<template>
  <button
    type="button"
    class="theme-toggle"
    :aria-pressed="theme === 'dark'"
    :aria-label="label"
    @click="toggleTheme"
  >
    <AppIcon :name="theme === 'dark' ? 'light_mode' : 'dark_mode'" />
  </button>
</template>

<style scoped>
.theme-toggle {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  width: 2.25rem;
  height: 2.25rem;
  padding: 0;
  border: 1px solid var(--kore-border);
  border-radius: var(--kore-radius-md);
  background: var(--kore-bg-elevated);
  color: var(--kore-text);
  cursor: pointer;
  transition: border-color 0.15s, color 0.15s;
}

.theme-toggle:hover {
  border-color: var(--kore-brand-gold);
  color: var(--kore-brand-gold);
}
</style>
