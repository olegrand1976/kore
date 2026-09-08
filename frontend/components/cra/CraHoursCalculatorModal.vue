<template>
  <AppModal
    :open="open"
    width="md"
    :title-id="titleId"
    :aria-label="$t('cra.hours_calc.title')"
    @update:open="$emit('update:open', $event)"
  >
    <form class="hours-calc" @submit.prevent="onApply">
      <h2 :id="titleId" class="hours-calc__title">{{ $t('cra.hours_calc.title') }}</h2>
      <p class="hours-calc__hint">{{ $t('cra.hours_calc.hint') }}</p>

      <div class="hours-calc__fields">
        <AppInput
          id="hours-calc-start"
          v-model="start"
          :label="$t('cra.hours_calc.start')"
          placeholder="08:30"
          autocomplete="off"
        />
        <AppInput
          id="hours-calc-end"
          v-model="end"
          :label="$t('cra.hours_calc.end')"
          placeholder="12:30"
          autocomplete="off"
        />
        <AppInput
          id="hours-calc-break"
          v-model="breakMinutes"
          type="number"
          min="0"
          step="5"
          :label="$t('cra.hours_calc.break')"
        />
      </div>

      <p v-if="preview.ok" class="hours-calc__result" role="status">
        {{ $t('cra.hours_calc.result', { hours: preview.hoursLabel }) }}
      </p>
      <p v-if="preview.ok && willClamp" class="hours-calc__warn" role="status">
        {{ $t('cra.hours_calc.clamp_warn', { hours: clampedLabel }) }}
      </p>
      <p v-if="!preview.ok && touched" class="hours-calc__error" role="alert">
        {{ errorLabel }}
      </p>

      <div class="hours-calc__actions">
        <AppButton
          variant="ghost"
          size="sm"
          type="button"
          @click="$emit('update:open', false)"
        >
          {{ $t('common.cancel') }}
        </AppButton>
        <AppButton variant="primary" size="sm" type="submit" :disabled="!preview.ok">
          {{ $t('cra.hours_calc.apply') }}
        </AppButton>
      </div>
    </form>
  </AppModal>
</template>

<script setup lang="ts">
import {
  calculateHoursFromRange,
  clampHoursLabel,
  type HoursCalculatorResult
} from '~/utils/craHoursCalculator'

const props = withDefaults(defineProps<{
  open: boolean
  maxHours?: number
}>(), {
  maxHours: 8
})

const emit = defineEmits<{
  'update:open': [value: boolean]
  apply: [hoursLabel: string]
}>()

const { t } = useI18n()
const titleId = useId()
const start = ref('08:30')
const end = ref('12:30')
const breakMinutes = ref('0')
const touched = ref(false)

const preview = computed<HoursCalculatorResult>(() =>
  calculateHoursFromRange({
    start: start.value,
    end: end.value,
    breakMinutes: Number(breakMinutes.value)
  })
)

const willClamp = computed(() =>
  preview.value.ok && preview.value.hours > props.maxHours
)

const clampedLabel = computed(() =>
  preview.value.ok
    ? clampHoursLabel(preview.value.hours, props.maxHours)
    : clampHoursLabel(0, props.maxHours)
)

const errorLabel = computed(() => {
  if (preview.value.ok) return ''
  switch (preview.value.reason) {
    case 'invalid_time':
      return t('cra.hours_calc.error_invalid')
    case 'end_before_start':
      return t('cra.hours_calc.error_order')
    case 'break_exceeds':
      return t('cra.hours_calc.error_break')
    case 'zero_or_negative':
      return t('cra.hours_calc.error_zero')
    default: {
      const _exhaustive: never = preview.value.reason
      return _exhaustive
    }
  }
})

watch(
  () => props.open,
  (open) => {
    if (open) touched.value = false
  }
)

const onApply = () => {
  touched.value = true
  if (!preview.value.ok) return
  emit('apply', clampedLabel.value)
  emit('update:open', false)
}
</script>

<style scoped>
.hours-calc {
  display: grid;
  gap: var(--kore-space-md);
}

.hours-calc__title {
  margin: 0;
  font-size: var(--kore-text-h3);
}

.hours-calc__hint {
  margin: 0;
  font-size: var(--kore-text-small);
  color: var(--kore-text-muted);
}

.hours-calc__fields {
  display: grid;
  gap: var(--kore-space-sm);
  grid-template-columns: 1fr 1fr;
}

.hours-calc__fields > :last-child {
  grid-column: 1 / -1;
}

.hours-calc__result {
  margin: 0;
  font-weight: 600;
  color: var(--kore-text);
}

.hours-calc__warn {
  margin: 0;
  color: var(--kore-brand-gold);
  font-size: var(--kore-text-small);
}

.hours-calc__error {
  margin: 0;
  color: var(--kore-error);
  font-size: var(--kore-text-small);
}

.hours-calc__actions {
  display: flex;
  flex-wrap: wrap;
  gap: var(--kore-space-sm);
  justify-content: flex-end;
}

@media (max-width: 768px) {
  .hours-calc__fields {
    grid-template-columns: 1fr;
  }

  .hours-calc__actions {
    flex-direction: column-reverse;
  }

  .hours-calc__actions :deep(.app-btn) {
    width: 100%;
  }
}
</style>
