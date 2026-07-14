<template>
  <div class="activity-line" :class="absenceClass">
    <div class="activity-line__head">
      <AppIcon :name="icon" class="activity-line__icon" />
      <span class="activity-line__label">{{ label }}</span>
      <AppBadge v-if="origin === 'prefill'" variant="info">{{ $t('cra.origin_prefill') }}</AppBadge>
      <span v-if="absence && !hasHours" class="activity-line__full-day">{{ $t('cra.full_day_absence') }}</span>
    </div>
    <div class="activity-line__fields">
      <div v-if="!absence || hasHours" class="activity-line__hours">
        <button type="button" class="stepper-btn" :disabled="disabled" :aria-label="$t('cra.decrease_hours')" @click="step(-0.5)">−</button>
        <AppInput
          :id="inputId"
          v-model="localHours"
          type="number"
          min="0"
          :max="maxHours"
          step="0.5"
          :label="$t('cra.hours')"
          :disabled="disabled"
        />
        <button type="button" class="stepper-btn" :disabled="disabled" :aria-label="$t('cra.increase_hours')" @click="step(0.5)">+</button>
      </div>
      <div v-else-if="allowPartialAbsence" class="activity-line__hours-placeholder">
        <span class="activity-line__hours-placeholder-label">{{ $t('cra.hours') }}</span>
        <span class="activity-line__hours-placeholder-value">{{ $t('cra.full_day_absence') }}</span>
        <button
          type="button"
          class="activity-line__hours-edit"
          :disabled="disabled"
          @click.stop="startPartialAbsence"
        >
          {{ $t('cra.enter_partial_hours') }}
        </button>
      </div>
      <div v-else class="activity-line__hours-placeholder">
        <span class="activity-line__hours-placeholder-label">{{ $t('cra.hours') }}</span>
        <span class="activity-line__hours-placeholder-value">{{ $t('cra.full_day_absence') }}</span>
      </div>
      <AppInput v-model="localComment" :label="$t('cra.comment')" :disabled="disabled" />
      <label v-if="!absence" class="activity-line__billable">
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
import { partialAbsenceHoursLabel } from '~/utils/craDayState'

const props = withDefaults(defineProps<{
  inputId: string
  label: string
  icon: string
  sourceType: string
  hours: string
  comment: string
  origin: string
  billable: boolean
  absence?: boolean
  dayCapacityMinutes?: number
  disabled?: boolean
  canRemove?: boolean
}>(), {
  dayCapacityMinutes: 8 * 60
})

const emit = defineEmits<{
  'update:hours': [value: string]
  'update:comment': [value: string]
  'update:billable': [value: boolean]
  remove: []
}>()

const hasHours = computed(() => {
  const value = Number(props.hours)
  return Number.isFinite(value) && value > 0
})

const allowPartialAbsence = computed(() => Boolean(props.absence))

const maxHours = computed(() => Math.max(0.5, props.dayCapacityMinutes / 60))

const absenceClass = computed(() => {
  if (!props.absence) return ''
  switch (props.sourceType) {
    case 'holiday':
      return 'activity-line--absence activity-line--absence-holiday'
    case 'leave':
    case 'conge':
      return 'activity-line--absence activity-line--absence-leave'
    default:
      return 'activity-line--absence'
  }
})

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
  const next = Math.max(0, Math.min(maxHours.value, current + delta))
  emit('update:hours', Number.isInteger(next) ? String(next) : next.toFixed(1))
}

const startPartialAbsence = () => {
  emit('update:hours', partialAbsenceHoursLabel(props.dayCapacityMinutes))
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

.activity-line--absence {
  padding: var(--kore-space-sm);
  border: 1px solid var(--kore-border);
  border-radius: var(--kore-radius-md);
  background: var(--kore-bg-elevated);
  border-bottom: 1px solid var(--kore-border);
}

.activity-line--absence-holiday .activity-line__icon {
  color: var(--kore-brand-gold);
}

.activity-line--absence-leave .activity-line__icon {
  color: var(--kore-brand-blue);
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

.activity-line__full-day {
  font-size: var(--kore-text-caption);
  color: var(--kore-text-muted);
  font-style: italic;
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

.activity-line__hours-placeholder {
  display: grid;
  gap: var(--kore-space-xs);
  padding: var(--kore-space-sm);
  border: 1px dashed var(--kore-border);
  border-radius: var(--kore-radius-md);
  background: var(--kore-bg-subtle);
}

.activity-line__hours-placeholder-label {
  font-size: var(--kore-text-caption);
  color: var(--kore-text-muted);
}

.activity-line__hours-placeholder-value {
  font-size: var(--kore-text-small);
  font-weight: 600;
}

.activity-line__hours-edit {
  justify-self: start;
  border: none;
  background: none;
  padding: 0;
  color: var(--kore-link);
  font-size: var(--kore-text-caption);
  cursor: pointer;
  text-decoration: underline;
}

.activity-line__hours-edit:disabled {
  opacity: 0.5;
  cursor: not-allowed;
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
