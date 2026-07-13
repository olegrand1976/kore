<script setup lang="ts">
type KpiTone = 'default' | 'gold' | 'success' | 'warn' | 'blue'

const props = withDefaults(
  defineProps<{
    icon: string
    label: string
    value?: string | number
    hint?: string
    tone?: KpiTone
    loading?: boolean
    error?: boolean
    to?: string
    clickable?: boolean
  }>(),
  { tone: 'default', loading: false, error: false, clickable: false }
)

const emit = defineEmits<{ click: [] }>()

const { t } = useI18n()

const toneClass = computed(() => {
  switch (props.tone) {
    case 'default':
      return ''
    case 'gold':
      return 'feature-card__icon--gold'
    case 'success':
      return 'feature-card__icon--success'
    case 'warn':
      return 'feature-card__icon--warn'
    case 'blue':
      return 'feature-card__icon--blue'
    default: {
      const _exhaustive: never = props.tone
      return _exhaustive
    }
  }
})

const displayValue = computed(() => {
  if (props.loading) return '…'
  if (props.error) return t('common.unavailable')
  return props.value ?? '—'
})

const valueClass = computed(() => ({
  'kpi-card__value--muted': props.loading || props.error
}))
</script>

<template>
  <NuxtLink v-if="to" :to="to" class="kpi-card-link">
    <AppCard padding="lg" hoverable class="kpi-card kpi-card--clickable" role="group" :aria-label="label">
      <div class="feature-card__icon" :class="toneClass">
        <AppIcon :name="icon" />
      </div>
      <p class="kpi-card__value" :class="valueClass">{{ displayValue }}</p>
      <p class="kpi-card__label">{{ label }}</p>
      <p v-if="hint" class="kpi-card__hint">{{ hint }}</p>
      <slot />
    </AppCard>
  </NuxtLink>
  <AppCard
    v-else
    padding="lg"
    hoverable
    class="kpi-card"
    :class="{ 'kpi-card--clickable': clickable }"
    :role="clickable ? 'button' : 'group'"
    :tabindex="clickable ? 0 : undefined"
    :aria-label="label"
    @click="clickable ? emit('click') : undefined"
    @keydown.enter.prevent="clickable ? emit('click') : undefined"
    @keydown.space.prevent="clickable ? emit('click') : undefined"
  >
    <div class="feature-card__icon" :class="toneClass">
      <AppIcon :name="icon" />
    </div>
    <p class="kpi-card__value" :class="valueClass">{{ displayValue }}</p>
    <p class="kpi-card__label">{{ label }}</p>
    <p v-if="hint" class="kpi-card__hint">{{ hint }}</p>
    <slot />
  </AppCard>
</template>

<style scoped>
.kpi-card-link {
  display: block;
  color: inherit;
  text-decoration: none;
}

.kpi-card--clickable {
  cursor: pointer;
  height: 100%;
}

.kpi-card__value--muted {
  color: var(--kore-text-muted);
  font-size: var(--kore-text-h3);
}

.kpi-card__hint {
  margin: 0;
  font-size: var(--kore-text-caption);
  color: var(--kore-text-muted);
  line-height: 1.35;
}
</style>
