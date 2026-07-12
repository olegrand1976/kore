<script setup lang="ts">
withDefaults(
  defineProps<{ variant?: 'icon' | 'chip' }>(),
  { variant: 'icon' }
)

const { theme, toggleTheme } = useTheme()
const { t } = useI18n()

const label = computed(() =>
  theme.value === 'dark' ? t('theme.aria_light') : t('theme.aria_dark')
)
</script>

<template>
  <button
    type="button"
    class="theme-toggle"
    :class="{ 'theme-toggle--chip': variant === 'chip' }"
    :aria-pressed="theme === 'dark'"
    :aria-label="label"
    :title="label"
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
  color: var(--kore-text-muted);
  cursor: pointer;
  transition: border-color 0.15s, color 0.15s, background 0.15s;
}

.theme-toggle--chip {
  width: auto;
  height: auto;
  padding: 0.375rem 0.625rem;
  border-radius: var(--kore-radius-full);
  background: var(--kore-bg);
}

.theme-toggle:hover {
  border-color: var(--kore-brand-gold);
  color: var(--kore-brand-gold);
}
</style>
