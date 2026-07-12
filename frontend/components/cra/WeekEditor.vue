<template>
  <div class="week-editor">
    <div v-for="(row, idx) in rows" :key="row.day" class="week-editor__row">
      <label :for="`day-${weekNumber}-${idx}`">{{ formatDay(row.day) }}</label>
      <AppInput
        :id="`day-${weekNumber}-${idx}`"
        v-model="row.hours"
        type="number"
        min="0"
        max="8"
        step="0.5"
        :label="$t('cra.hours')"
      />
      <AppInput v-model="row.comment" :label="$t('cra.comment')" />
    </div>
    <div class="week-editor__actions">
      <AppButton variant="primary" size="sm" :disabled="disabled || saving" @click="emitSave">
        {{ $t('cra.save_week') }}
      </AppButton>
      <AppButton variant="secondary" size="sm" :disabled="disabled || saving" @click="$emit('submit')">
        {{ $t('cra.submit_week') }}
      </AppButton>
    </div>
  </div>
</template>

<script setup lang="ts">
import type { CraLine, CraWeek } from '~/stores/cra'

const props = defineProps<{
  weekNumber: number
  week?: CraWeek
  month: string
  disabled?: boolean
  saving?: boolean
}>()

const emit = defineEmits<{
  save: [lines: CraLine[]]
  submit: []
}>()

const { locale } = useI18n()

type Row = { day: string; hours: string; comment: string }

const weekDays = computed(() => {
  const [y, m] = props.month.split('-').map(Number)
  const first = new Date(y, m - 1, 1)
  const startOffset = ((props.weekNumber - 1) * 7) + (1 - first.getDay() + 7) % 7
  const days: string[] = []
  for (let i = 0; i < 5; i++) {
    const d = new Date(y, m - 1, 1 + startOffset + i)
    if (d.getMonth() !== m - 1) break
    days.push(d.toISOString().slice(0, 10))
  }
  if (days.length === 0) {
    for (let w = 1; w <= 5; w++) {
      days.push(`${props.month}-${String(w + (props.weekNumber - 1) * 5).padStart(2, '0')}`)
    }
  }
  return days
})

const rows = ref<Row[]>([])

watch(
  () => [props.week, props.weekNumber, props.month] as const,
  () => {
    const existing = new Map<string, CraLine>()
    for (const line of props.week?.lines ?? []) {
      existing.set(line.day.slice(0, 10), line)
    }
    rows.value = weekDays.value.map((day) => {
      const line = existing.get(day)
      return {
        day,
        hours: line ? String(line.duration / 60) : '',
        comment: line?.comment ?? ''
      }
    })
  },
  { immediate: true }
)

const formatDay = (day: string) =>
  new Date(day).toLocaleDateString(locale.value === 'en' ? 'en-US' : 'fr-FR', {
    weekday: 'short',
    day: 'numeric',
    month: 'short'
  })

const emitSave = () => {
  const lines: CraLine[] = rows.value
    .filter((r) => r.hours && Number(r.hours) > 0)
    .map((r) => ({
      sourceType: 'manual',
      sourceId: 'default',
      day: r.day,
      duration: Math.round(Number(r.hours) * 60),
      comment: r.comment
    }))
  emit('save', lines)
}
</script>

<style scoped>
.week-editor {
  display: flex;
  flex-direction: column;
  gap: var(--kore-space-md);
}

.week-editor__row {
  display: grid;
  gap: var(--kore-space-sm);
  grid-template-columns: 120px 1fr 1fr;
  align-items: end;
}

.week-editor__actions {
  display: flex;
  gap: var(--kore-space-sm);
  flex-wrap: wrap;
}

@media (max-width: 768px) {
  .week-editor__row {
    grid-template-columns: 1fr;
  }
}
</style>
