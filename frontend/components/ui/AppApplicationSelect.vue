<script setup lang="ts">
const props = defineProps<{
  modelValue: string
  label?: string
  id?: string
  required?: boolean
  error?: string
}>()

const emit = defineEmits<{ 'update:modelValue': [value: string] }>()

const { t } = useI18n()
const { list, pickAppId, pickAppLabel } = useApplications()

const apps = ref<Awaited<ReturnType<typeof list>>>([])
const loading = ref(true)

onMounted(async () => {
  try {
    apps.value = await list()
    if (!props.modelValue) {
      const first = apps.value[0]
      const id = first ? pickAppId(first) : ''
      if (id) emit('update:modelValue', id)
    }
  } finally {
    loading.value = false
  }
})
</script>

<template>
  <div class="app-application-select">
    <label v-if="label" :for="id" class="app-application-select__label">{{ label }}</label>
    <select
      :id="id"
      class="app-application-select__field"
      :class="{ 'app-application-select__field--error': !!error }"
      :value="modelValue"
      :required="required"
      :disabled="loading"
      @change="emit('update:modelValue', ($event.target as HTMLSelectElement).value)"
    >
      <option value="">{{ t('requests.form_application_placeholder') }}</option>
      <option v-for="app in apps" :key="pickAppId(app)" :value="pickAppId(app)">
        {{ pickAppLabel(app) }}
      </option>
    </select>
    <p v-if="error" class="app-application-select__error" role="alert">{{ error }}</p>
  </div>
</template>

<style scoped>
.app-application-select {
  display: flex;
  flex-direction: column;
  gap: var(--kore-space-xs);
}

.app-application-select__label {
  font-size: var(--kore-text-small);
  color: var(--kore-text-muted);
  font-weight: 500;
}

.app-application-select__field {
  padding: 0.75rem 1rem;
  font-family: var(--kore-font);
  font-size: var(--kore-text-body);
  color: var(--kore-text);
  background: var(--kore-bg-elevated);
  border: 1px solid var(--kore-border);
  border-radius: var(--kore-radius-md);
}

.app-application-select__field--error {
  border-color: var(--kore-error);
}

.app-application-select__error {
  margin: 0;
  font-size: var(--kore-text-caption);
  color: var(--kore-error);
}
</style>
