<script setup lang="ts">
const props = withDefaults(defineProps<{
  modelValue: File[]
  label?: string
  id?: string
  accept?: string
  multiple?: boolean
}>(), {
  multiple: true,
  accept: '.pdf,.png,.jpg,.jpeg,.gif,.webp,.txt,.csv,.doc,.docx,.xls,.xlsx,.zip,.log,.md'
})

const emit = defineEmits<{ 'update:modelValue': [files: File[]] }>()

const { t } = useI18n()

const onChange = (event: Event) => {
  const input = event.target as HTMLInputElement
  const files = input.files ? Array.from(input.files) : []
  emit('update:modelValue', props.multiple ? [...props.modelValue, ...files] : files)
  input.value = ''
}

const removeFile = (index: number) => {
  const next = [...props.modelValue]
  next.splice(index, 1)
  emit('update:modelValue', next)
}
</script>

<template>
  <div class="app-file-upload">
    <label v-if="label" :for="id" class="app-file-upload__label">{{ label }}</label>
    <input
      :id="id"
      class="app-file-upload__input"
      type="file"
      :accept="accept"
      :multiple="multiple"
      @change="onChange"
    />
    <p class="app-file-upload__hint">{{ t('requests.attachments_hint') }}</p>
    <ul v-if="modelValue.length" class="app-file-upload__list">
      <li v-for="(file, index) in modelValue" :key="`${file.name}-${index}`">
        <span>{{ file.name }}</span>
        <button type="button" class="app-file-upload__remove" @click="removeFile(index)">
          {{ t('common.delete') }}
        </button>
      </li>
    </ul>
  </div>
</template>

<style scoped>
.app-file-upload {
  display: flex;
  flex-direction: column;
  gap: var(--kore-space-xs);
}

.app-file-upload__label {
  font-size: var(--kore-text-small);
  color: var(--kore-text-muted);
  font-weight: 500;
}

.app-file-upload__input {
  font-size: var(--kore-text-small);
  color: var(--kore-text);
}

.app-file-upload__hint {
  margin: 0;
  font-size: var(--kore-text-caption);
  color: var(--kore-text-muted);
}

.app-file-upload__list {
  margin: 0;
  padding: 0;
  list-style: none;
  display: flex;
  flex-direction: column;
  gap: var(--kore-space-xs);
}

.app-file-upload__list li {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: var(--kore-space-sm);
  font-size: var(--kore-text-small);
}

.app-file-upload__remove {
  border: none;
  background: none;
  color: var(--kore-status-danger);
  cursor: pointer;
  font-size: var(--kore-text-caption);
}
</style>
