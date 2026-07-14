<template>
  <div class="activity-line">
    <div class="activity-line__head">
      <AppIcon :name="icon" class="activity-line__icon" />
      <span class="activity-line__label">{{ label }}</span>
      <AppBadge v-if="origin === 'prefill'" variant="info">{{ $t('cra.origin_prefill') }}</AppBadge>
    </div>
    <div class="activity-line__fields">
      <div class="activity-line__hours">
        <button type="button" class="stepper-btn" :disabled="disabled" :aria-label="$t('cra.decrease_hours')" @click="step(-0.5)">−</button>
        <AppInput
          :id="inputId"
          v-model="localHours"
          type="number"
          min="0"
          max="8"
          step="0.5"
          :label="$t('cra.hours')"
          :disabled="disabled"
        />
        <button type="button" class="stepper-btn" :disabled="disabled" :aria-label="$t('cra.increase_hours')" @click="step(0.5)">+</button>
      </div>
      <AppInput v-model="localComment" :label="$t('cra.comment')" :disabled="disabled" />
      <label class="activity-line__billable">
        <input v-model="localBillable" type="checkbox" :disabled="disabled">
        {{ $t('cra.billable') }}
      </label>
      <AppButton
        v-if="canRemove"
        variant="ghost"
        size="sm"
        :disabled="disabled"
        :aria-label="$t('cra.remove_line')"
        @click="$emit('remove')"
      >
        <AppIcon name="delete" />
      </AppButton>
    </div>
  </div>
</template>

<script setup lang="ts">
const props = defineProps<{
  inputId: string
  label: string
  icon: string
  hours: string
  comment: string
  origin: string
  billable: boolean
  disabled?: boolean
  canRemove?: boolean
}>()

const emit = defineEmits<{
  'update:hours': [value: string]
  'update:comment': [value: string]
  'update:billable': [value: boolean]
  remove: []
}>()

const localHours = computed({
  get: () => props.hours,
  set: (v: string) => emit('update:hours', v)
})

const localComment = computed({
  get: () => props.comment,
  set: (v: string) => emit('update:comment', v)
})

const localBillable = computed({
  get: () => props.billable,
  set: (v: boolean) => emit('update:billable', v)
})

const step = (delta: number) => {
  const current = Number(localHours.value) || 0
  const next = Math.max(0, Math.min(8, current + delta))
  emit('update:hours', Number.isInteger(next) ? String(next) : next.toFixed(1))
}
</script>

<style scoped>
.activity-line {
  display: grid;
  gap: var(--kore-space-sm);
  padding: var(--kore-space-sm) 0;
  border-bottom: 1px solid var(--kore-border);
}

.activity-line:last-child {
  border-bottom: none;
}

.activity-line__head {
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  gap: var(--kore-space-xs);
}

.activity-line__icon {
  color: var(--kore-brand-gold);
}

.activity-line__label {
  font-weight: 600;
  font-size: var(--kore-text-small);
}

.activity-line__fields {
  display: grid;
  gap: var(--kore-space-sm);
  grid-template-columns: minmax(10rem, 14rem) 1fr auto;
  align-items: end;
}

.activity-line__hours {
  display: grid;
  grid-template-columns: 2.75rem 1fr 2.75rem;
  gap: var(--kore-space-xs);
  align-items: end;
}

.stepper-btn {
  min-height: 2.75rem;
  min-width: 2.75rem;
  border: 1px solid var(--kore-border);
  border-radius: var(--kore-radius-md);
  background: var(--kore-bg);
  color: var(--kore-text);
  cursor: pointer;
  font-size: 1.1rem;
}

.stepper-btn:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}

.activity-line__billable {
  display: flex;
  align-items: center;
  gap: var(--kore-space-xs);
  font-size: var(--kore-text-small);
  color: var(--kore-text-muted);
}

@media (max-width: 768px) {
  .activity-line__fields {
    grid-template-columns: 1fr;
  }
}
</style>
