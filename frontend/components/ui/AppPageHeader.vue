<script setup lang="ts">
defineProps<{
  title: string
  subtitle?: string
}>()
</script>

<template>
  <header class="app-page-header">
    <div>
      <h1 class="app-page-header__title">{{ title }}</h1>
      <p v-if="subtitle" class="app-page-header__subtitle">{{ subtitle }}</p>
    </div>
    <div v-if="$slots.actions" class="app-page-header__actions">
      <slot name="actions" />
    </div>
  </header>
</template>

<style scoped>
.app-page-header {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  gap: var(--kore-space-lg);
  margin-bottom: var(--kore-space-xl);
}

.app-page-header__title {
  font-size: var(--kore-text-h1);
  margin: 0;
}

.app-page-header__subtitle {
  margin: var(--kore-space-sm) 0 0;
  color: var(--kore-text-muted);
}

.app-page-header__actions {
  display: flex;
  flex-wrap: wrap;
  gap: var(--kore-space-sm);
  flex-shrink: 0;
}

@media (max-width: 640px) {
  .app-page-header {
    flex-direction: column;
    align-items: stretch;
  }

  .app-page-header__title {
    font-size: var(--kore-text-h2);
  }

  .app-page-header__actions {
    width: 100%;
  }

  .app-page-header__actions :deep(.app-btn) {
    flex: 1;
    justify-content: center;
  }
}
</style>
