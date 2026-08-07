<script setup lang="ts">
export type ApplicationOption = {
  id: string
  libelle: string
  active?: boolean
}

const props = defineProps<{
  modelValue: string[]
  applications: ApplicationOption[]
  label?: string
  id?: string
  emptyHint?: string
}>()

const emit = defineEmits<{ 'update:modelValue': [value: string[]] }>()

const { t } = useI18n()

const isChecked = (applicationId: string) => props.modelValue.includes(applicationId)

const toggle = (applicationId: string) => {
  const app = props.applications.find((a) => a.id === applicationId)
  const checked = props.modelValue.includes(applicationId)
  // Inactive apps can only be removed (uncheck), not newly selected.
  if (app?.active === false && !checked) return
  const next = new Set(props.modelValue)
  if (checked) next.delete(applicationId)
  else next.add(applicationId)
  emit('update:modelValue', [...next])
}
</script>

<template>
  <div class="mission-app-multi-select">
    <p v-if="label" :id="`${id}-label`" class="mission-app-multi-select__label">{{ label }}</p>
    <p v-if="!applications.length" class="mission-app-multi-select__hint">
      {{ emptyHint || t('missions.apps_empty_hint') }}
    </p>
    <ul
      v-else
      class="mission-app-multi-select__list"
      role="group"
      :aria-labelledby="label ? `${id}-label` : undefined"
    >
      <li v-for="app in applications" :key="app.id">
        <label class="mission-app-multi-select__item">
          <input
            type="checkbox"
            :checked="isChecked(app.id)"
            :disabled="app.active === false && !isChecked(app.id)"
            @change="toggle(app.id)"
          />
          <span>
            <strong>{{ app.libelle }}</strong>
            <span v-if="app.active === false" class="mission-app-multi-select__meta">
              — {{ t('missions.app_inactive') }}
            </span>
          </span>
        </label>
      </li>
    </ul>
  </div>
</template>

<style scoped>
.mission-app-multi-select {
  display: flex;
  flex-direction: column;
  gap: var(--kore-space-xs);
}

.mission-app-multi-select__label {
  margin: 0;
  font-size: var(--kore-text-small);
  color: var(--kore-text-muted);
  font-weight: 500;
}

.mission-app-multi-select__hint {
  margin: 0;
  font-size: var(--kore-text-caption);
  color: var(--kore-text-muted);
}

.mission-app-multi-select__list {
  margin: 0;
  padding: 0;
  list-style: none;
  display: grid;
  gap: var(--kore-space-xs);
  max-height: 14rem;
  overflow: auto;
}

.mission-app-multi-select__item {
  display: flex;
  align-items: flex-start;
  gap: var(--kore-space-sm);
  cursor: pointer;
  font-size: var(--kore-text-small);
  color: var(--kore-text);
}

.mission-app-multi-select__meta {
  color: var(--kore-text-muted);
  font-weight: 400;
}
</style>
