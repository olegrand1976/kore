<script setup lang="ts">
const open = defineModel<boolean>('open', { default: false })

watch(open, (isOpen) => {
  if (import.meta.client) {
    document.body.style.overflow = isOpen ? 'hidden' : ''
  }
})

onUnmounted(() => {
  if (import.meta.client) {
    document.body.style.overflow = ''
  }
})
</script>

<template>
  <Teleport to="body">
    <Transition name="drawer-fade">
      <button
        v-if="open"
        type="button"
        class="drawer-backdrop"
        aria-label="Fermer le menu"
        @click="open = false"
      />
    </Transition>
    <Transition name="drawer-slide">
      <nav v-if="open" class="drawer" aria-label="Menu mobile">
        <slot />
      </nav>
    </Transition>
  </Teleport>
</template>

<style scoped>
.drawer-backdrop {
  position: fixed;
  inset: 0;
  z-index: 80;
  border: none;
  background: rgba(0, 0, 0, 0.45);
  cursor: pointer;
}

.drawer {
  position: fixed;
  top: 0;
  right: 0;
  z-index: 90;
  display: flex;
  flex-direction: column;
  gap: var(--kore-space-xs);
  width: min(300px, 88vw);
  height: 100dvh;
  padding: var(--kore-space-xl) var(--kore-space-lg);
  padding-top: calc(var(--kore-space-xl) + env(safe-area-inset-top, 0px));
  background: var(--kore-bg-elevated);
  border-left: 1px solid var(--kore-border);
  box-shadow: var(--kore-shadow-md);
  overflow-y: auto;
}

.drawer-fade-enter-active,
.drawer-fade-leave-active { transition: opacity 0.2s; }
.drawer-fade-enter-from,
.drawer-fade-leave-to { opacity: 0; }

.drawer-slide-enter-active,
.drawer-slide-leave-active { transition: transform 0.22s ease; }
.drawer-slide-enter-from,
.drawer-slide-leave-to { transform: translateX(100%); }
</style>
