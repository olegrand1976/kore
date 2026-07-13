<script setup lang="ts">
type Props = {
  buttonLabel?: string
}

const props = defineProps<Props>()

const isOpen = ref(false)
const rootEl = ref<HTMLElement | null>(null)
const tooltipId = ref<string>('')

let onDocPointerDown: ((e: PointerEvent) => void) | null = null
let onDocKeyDown: ((e: KeyboardEvent) => void) | null = null

const close = () => {
  isOpen.value = false
}

const toggle = () => {
  isOpen.value = !isOpen.value
}

onMounted(() => {
  if (!tooltipId.value) {
    const uuid =
      typeof crypto !== 'undefined' && typeof crypto.randomUUID === 'function'
        ? crypto.randomUUID()
        : `${Date.now()}-${Math.random().toString(16).slice(2)}`
    tooltipId.value = `app-tooltip-${uuid}`
  }

  onDocPointerDown = (e: PointerEvent) => {
    if (!isOpen.value) return
    const target = e.target as Node | null
    if (!target) return
    if (rootEl.value?.contains(target)) return
    close()
  }

  onDocKeyDown = (e: KeyboardEvent) => {
    if (!isOpen.value) return
    if (e.key === 'Escape') close()
  }

  document.addEventListener('pointerdown', onDocPointerDown, true)
  document.addEventListener('keydown', onDocKeyDown)
})

onBeforeUnmount(() => {
  if (onDocPointerDown) document.removeEventListener('pointerdown', onDocPointerDown, true)
  if (onDocKeyDown) document.removeEventListener('keydown', onDocKeyDown)
})
</script>

<template>
  <span ref="rootEl" class="app-tooltip">
    <button
      type="button"
      class="app-tooltip__button"
      :aria-expanded="isOpen"
      :aria-controls="isOpen ? tooltipId : undefined"
      :aria-describedby="isOpen ? tooltipId : undefined"
      :aria-label="props.buttonLabel || 'Info'"
      @click="toggle"
    >
      <span aria-hidden="true" class="app-tooltip__icon">i</span>
    </button>
    <span
      v-if="isOpen && tooltipId"
      :id="tooltipId"
      class="app-tooltip__panel"
      role="tooltip"
    >
      <slot />
    </span>
  </span>
</template>

<style scoped>
.app-tooltip { position: relative; display: inline-flex; align-items: center; }

.app-tooltip__button {
  width: 2rem;
  height: 2rem;
  display: inline-flex;
  align-items: center;
  justify-content: center;
  border-radius: 999px;
  border: 1px solid var(--kore-border);
  background: var(--kore-bg-elevated);
  color: var(--kore-text);
  cursor: pointer;
}

.app-tooltip__button:focus-visible {
  outline: 2px solid var(--kore-primary);
  outline-offset: 2px;
}

.app-tooltip__icon {
  font-size: var(--kore-text-small);
  font-weight: 700;
  line-height: 1;
}

.app-tooltip__panel {
  position: absolute;
  top: calc(100% + 0.5rem);
  right: 0;
  z-index: 50;
  max-width: min(22rem, 80vw);
  padding: 0.75rem 0.9rem;
  border-radius: var(--kore-radius-md);
  border: 1px solid var(--kore-border);
  background: var(--kore-bg-elevated);
  color: var(--kore-text);
  font-size: var(--kore-text-small);
  line-height: 1.35;
  box-shadow: 0 12px 28px rgba(0, 0, 0, 0.25);
  overflow-wrap: anywhere;
  white-space: normal;
}
</style>
