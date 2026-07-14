<script setup lang="ts">
export type ServiceRequestPayload = {
  applicationId: string
  subject: string
  description: string
  priority: string
  dueAt: string
  requiresChefGate?: boolean
  files: File[]
}

const props = withDefaults(defineProps<{
  showChefGate?: boolean
  submitLabel?: string
  busy?: boolean
}>(), {
  showChefGate: false,
  busy: false
})

const emit = defineEmits<{ submit: [payload: ServiceRequestPayload] }>()

const { t } = useI18n()

const applicationId = ref('')
const subject = ref('')
const description = ref('')
const priority = ref('normal')
const dueAt = ref('')
const requiresChefGate = ref(false)
const files = ref<File[]>([])

const priorityOptions = [
  { value: 'low', labelKey: 'requests.priority_low' },
  { value: 'normal', labelKey: 'requests.priority_normal' },
  { value: 'high', labelKey: 'requests.priority_high' },
  { value: 'urgent', labelKey: 'requests.priority_urgent' }
] as const

const onSubmit = () => {
  if (!subject.value.trim() || !applicationId.value) return
  emit('submit', {
    applicationId: applicationId.value,
    subject: subject.value.trim(),
    description: description.value.trim(),
    priority: priority.value,
    dueAt: dueAt.value,
    requiresChefGate: props.showChefGate ? requiresChefGate.value : undefined,
    files: files.value
  })
}
</script>

<template>
  <form class="service-request-form" @submit.prevent="onSubmit">
    <AppApplicationSelect
      id="request-application"
      v-model="applicationId"
      :label="t('requests.form_application')"
      required
    />
    <AppInput
      id="request-subject"
      v-model="subject"
      :label="t('requests.form_subject')"
      required
    />
    <div class="service-request-form__field">
      <label for="request-description" class="service-request-form__label">
        {{ t('requests.form_description') }}
      </label>
      <textarea
        id="request-description"
        v-model="description"
        class="service-request-form__textarea"
        rows="4"
      />
    </div>
    <div class="service-request-form__row">
      <div class="service-request-form__field">
        <label for="request-priority" class="service-request-form__label">
          {{ t('requests.form_priority') }}
        </label>
        <select id="request-priority" v-model="priority" class="service-request-form__select">
          <option v-for="opt in priorityOptions" :key="opt.value" :value="opt.value">
            {{ t(opt.labelKey) }}
          </option>
        </select>
      </div>
      <AppInput
        id="request-due-at"
        v-model="dueAt"
        type="datetime-local"
        :label="t('requests.form_due_at')"
      />
    </div>
    <AppFileUpload id="request-files" v-model="files" :label="t('requests.form_attachments')" />
    <label v-if="showChefGate" class="service-request-form__check">
      <input v-model="requiresChefGate" type="checkbox" />
      {{ t('requests.form_chef_gate') }}
    </label>
    <AppButton
      variant="primary"
      size="sm"
      type="submit"
      :disabled="busy || !subject.trim() || !applicationId"
    >
      {{ submitLabel || t('requests.form_submit') }}
    </AppButton>
  </form>
</template>

<style scoped>
.service-request-form {
  display: grid;
  gap: var(--kore-space-md);
  max-width: var(--kore-form-wide-max);
}

.service-request-form__row {
  display: grid;
  gap: var(--kore-space-md);
}

@media (min-width: 640px) {
  .service-request-form__row {
    grid-template-columns: 1fr 1fr;
  }
}

.service-request-form__field {
  display: flex;
  flex-direction: column;
  gap: var(--kore-space-xs);
}

.service-request-form__label {
  font-size: var(--kore-text-small);
  color: var(--kore-text-muted);
  font-weight: 500;
}

.service-request-form__textarea,
.service-request-form__select {
  padding: 0.75rem 1rem;
  font-family: var(--kore-font);
  font-size: var(--kore-text-body);
  color: var(--kore-text);
  background: var(--kore-bg-elevated);
  border: 1px solid var(--kore-border);
  border-radius: var(--kore-radius-md);
}

.service-request-form__textarea {
  resize: vertical;
  min-height: 6rem;
}

.service-request-form__check {
  display: flex;
  align-items: center;
  gap: var(--kore-space-sm);
  font-size: var(--kore-text-small);
}

@media (max-width: 768px) {
  .service-request-form :deep(.app-button) {
    width: 100%;
  }
}
</style>
