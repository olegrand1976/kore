<script setup lang="ts">
const props = defineProps<{
  open: boolean
  titleId?: string
  ariaLabel?: string
  closeLabel?: string
  width?: 'sm' | 'md' | 'lg'
}>()

const emit = defineEmits<{ 'update:open': [open: boolean] }>()

const onClose = () => emit('update:open', false)

const onKeydown = (e: KeyboardEvent) => {
  if (e.key === 'Escape') onClose()
}

const panelWidthClass = computed(() => {
  switch (props.width ?? 'md') {
    case 'sm':
      return 'app-modal__panel--sm'
    case 'md':
      return 'app-modal__panel--md'
    case 'lg':
      return 'app-modal__panel--lg'
    default: {
      const _exhaustive: never = props.width as never
      return _exhaustive
    }
  }
})

watch(
  () => props.open,
  (open) => {
    if (!open || !import.meta.client) return
    // Let the DOM render, then focus the first focusable element.
    requestAnimationFrame(() => {
      const el = document.querySelector<HTMLElement>('[data-modal-panel] button, [data-modal-panel] [href], [data-modal-panel] input, [data-modal-panel] select, [data-modal-panel] textarea, [data-modal-panel] [tabindex]:not([tabindex="-1"])')
      el?.focus()
    })
  }
)
</script>

<template>
  <Teleport to="body">
    <div v-if="open" class="app-modal" @keydown="onKeydown">
      <button
        type="button"
        class="app-modal__backdrop"
        :aria-label="closeLabel || ariaLabel || 'Close dialog'"
        @click="onClose"
      />
      <section
        class="app-modal__panel"
        :class="panelWidthClass"
        data-modal-panel
        role="dialog"
        aria-modal="true"
        :aria-label="titleId ? undefined : ariaLabel"
        :aria-labelledby="titleId || undefined"
      >
        <slot />
      </section>
    </div>
  </Teleport>
</template>

<style scoped>
.app-modal {
  position: fixed;
  inset: 0;
  z-index: 50;
  display: grid;
  place-items: center;
  padding: var(--kore-space-md);
}

.app-modal__backdrop {
  position: absolute;
  inset: 0;
  background: rgba(0, 0, 0, 0.45);
  border: 0;
  cursor: pointer;
}

.app-modal__panel {
  position: relative;
  width: 100%;
  max-height: min(80vh, 56rem);
  overflow: auto;
  background: var(--kore-bg-elevated);
  border: 1px solid var(--kore-border);
  border-radius: var(--kore-radius-lg);
  box-shadow: var(--kore-shadow-lg);
  padding: var(--kore-space-xl);
}

.app-modal__panel--sm { max-width: 34rem; }
.app-modal__panel--md { max-width: 48rem; }
.app-modal__panel--lg { max-width: 64rem; }

@media (max-width: 768px) {
  .app-modal { padding: var(--kore-space-sm); }
  .app-modal__panel { padding: var(--kore-space-lg); max-height: 86vh; }
}
</style>
