<script setup lang="ts">
const props = defineProps<{
  guideKey: string
  defaultCollapsed?: boolean
}>()

const { t, tm } = useI18n()
const { guidesEnabled } = useRequestSettings()

const storageKey = computed(() => `kore.guide.dismissed.${props.guideKey}`)
const dismissed = ref(false)
const collapsed = ref(props.defaultCollapsed ?? false)

onMounted(() => {
  if (import.meta.client) {
    dismissed.value = localStorage.getItem(storageKey.value) === '1'
    if (window.matchMedia('(max-width: 768px)').matches) {
      collapsed.value = true
    }
  }
})

const visible = computed(() => guidesEnabled.value && !dismissed.value)

const title = computed(() => t(`guides.${props.guideKey}.title`))
const items = computed(() => {
  const raw = tm(`guides.${props.guideKey}.items`) as string[] | { body?: { static?: string } }
  if (Array.isArray(raw)) return raw
  return []
})

const dismiss = () => {
  dismissed.value = true
  if (import.meta.client) {
    localStorage.setItem(storageKey.value, '1')
  }
}

const showAgain = () => {
  dismissed.value = false
  if (import.meta.client) {
    localStorage.removeItem(storageKey.value)
  }
}

defineExpose({ showAgain, dismissed })
</script>

<template>
  <AppCard v-if="visible" padding="lg" class="section-guide">
    <div class="section-guide__head">
      <button
        type="button"
        class="section-guide__toggle"
        :aria-expanded="!collapsed"
        @click="collapsed = !collapsed"
      >
        <span class="material-symbols-outlined section-guide__icon" aria-hidden="true">info</span>
        <span class="section-guide__title">{{ title }}</span>
        <span class="material-symbols-outlined section-guide__chevron" aria-hidden="true">
          {{ collapsed ? 'expand_more' : 'expand_less' }}
        </span>
      </button>
      <AppButton variant="ghost" size="sm" type="button" @click="dismiss">
        {{ t('guides.hide') }}
      </AppButton>
    </div>
    <ul v-show="!collapsed" class="section-guide__list">
      <li v-for="(item, idx) in items" :key="idx">{{ item }}</li>
    </ul>
  </AppCard>
</template>

<style scoped>
.section-guide {
  margin-bottom: var(--kore-space-md);
}

.section-guide__head {
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  justify-content: space-between;
  gap: var(--kore-space-sm);
}

.section-guide__toggle {
  display: inline-flex;
  align-items: center;
  gap: 0.5rem;
  border: none;
  background: none;
  padding: 0;
  cursor: pointer;
  color: var(--kore-text);
  font: inherit;
  font-weight: 600;
  text-align: left;
}

.section-guide__icon {
  color: var(--kore-gold);
  font-size: 1.25rem;
}

.section-guide__chevron {
  color: var(--kore-text-muted);
}

.section-guide__list {
  margin: var(--kore-space-sm) 0 0;
  padding-left: 1.25rem;
  color: var(--kore-text-muted);
  font-size: var(--kore-text-small);
  display: flex;
  flex-direction: column;
  gap: 0.35rem;
}

@media (max-width: 768px) {
  .section-guide__head {
    flex-direction: column;
    align-items: flex-start;
  }
}
</style>
