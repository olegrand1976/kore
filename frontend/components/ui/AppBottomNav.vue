<script setup lang="ts">
export type BottomNavItem = {
  to: string
  icon: string
  label: string
}

defineProps<{
  items: BottomNavItem[]
}>()
</script>

<template>
  <nav class="bottom-nav" aria-label="Navigation principale mobile">
    <NuxtLink
      v-for="item in items"
      :key="item.to"
      :to="item.to"
      class="bottom-nav__item"
    >
      <AppIcon :name="item.icon" />
      <span>{{ item.label }}</span>
    </NuxtLink>
  </nav>
</template>

<style scoped>
.bottom-nav {
  position: fixed;
  left: 0;
  right: 0;
  bottom: 0;
  z-index: 60;
  display: none;
  grid-auto-flow: column;
  grid-auto-columns: 1fr;
  gap: var(--kore-space-xs);
  padding: var(--kore-space-sm) var(--kore-space-md);
  padding-bottom: calc(var(--kore-space-sm) + env(safe-area-inset-bottom, 0px));
  background: color-mix(in srgb, var(--kore-bg-elevated) 92%, transparent);
  border-top: 1px solid var(--kore-border);
  backdrop-filter: blur(10px);
}

@media (max-width: 768px) {
  .bottom-nav { display: grid; }
}

.bottom-nav__item {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 0.125rem;
  padding: 0.375rem 0.25rem;
  color: var(--kore-text-muted);
  text-decoration: none;
  font-size: 0.625rem;
  font-weight: 600;
  border-radius: var(--kore-radius-md);
  min-width: 0;
}

.bottom-nav__item span {
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  max-width: 100%;
}

.bottom-nav__item :deep(.material-symbols-outlined) {
  font-size: 1.375rem !important;
}

.bottom-nav__item.router-link-active {
  color: var(--kore-brand-gold);
  background: rgba(201, 162, 39, 0.1);
}
</style>
