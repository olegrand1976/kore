<script setup lang="ts">
defineProps<{
  icon?: string
  title: string
  description?: string
}>()
</script>

<template>
  <div class="empty-state">
    <div v-if="icon" class="empty-state__icon" aria-hidden="true">
      <AppIcon :name="icon" />
    </div>
    <h3 class="empty-state__title">{{ title }}</h3>
    <p v-if="description" class="empty-state__desc">{{ description }}</p>
    <div v-if="$slots.default" class="empty-state__action">
      <slot />
    </div>
  </div>
</template>

<style scoped>
.empty-state {
  display: flex;
  flex-direction: column;
  align-items: center;
  text-align: center;
  padding: var(--kore-space-2xl) var(--kore-space-lg);
}

.empty-state__icon {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 4rem;
  height: 4rem;
  margin-bottom: var(--kore-space-lg);
  border-radius: var(--kore-radius-full);
  background: rgba(201, 162, 39, 0.1);
  color: var(--kore-brand-gold);
}

.empty-state__icon :deep(.material-symbols-outlined) {
  font-size: 2rem !important;
}

.empty-state__title {
  margin: 0 0 var(--kore-space-sm);
  font-size: var(--kore-text-h3);
  font-weight: 600;
}

.empty-state__desc {
  margin: 0;
  max-width: 360px;
  font-size: var(--kore-text-small);
  color: var(--kore-text-muted);
  line-height: 1.55;
}

.empty-state__action {
  margin-top: var(--kore-space-lg);
}
</style>
