<script setup lang="ts">
defineProps<{
  modelValue: string
  label?: string
  type?: string
  placeholder?: string
  required?: boolean
  error?: string
  id?: string
}>()

defineEmits<{ 'update:modelValue': [value: string] }>()
</script>

<template>
  <div class="pub-input">
    <label v-if="label" :for="id" class="pub-input__label">{{ label }}</label>
    <input
      :id="id"
      class="pub-input__field"
      :class="{ 'pub-input__field--error': !!error }"
      :type="type || 'text'"
      :value="modelValue"
      :placeholder="placeholder"
      :required="required"
      @input="$emit('update:modelValue', ($event.target as HTMLInputElement).value)"
    />
    <p v-if="error" class="pub-input__error" role="alert">{{ error }}</p>
  </div>
</template>

<style scoped>
.pub-input {
  display: flex;
  flex-direction: column;
  gap: var(--kore-space-xs);
}

.pub-input__label {
  font-size: var(--kore-text-small);
  color: var(--kore-text-muted);
  font-weight: 500;
}

.pub-input__field {
  padding: 0.75rem 1rem;
  font-family: var(--kore-font);
  font-size: var(--kore-text-body);
  color: var(--kore-text);
  background: var(--kore-bg-elevated);
  border: 1px solid var(--kore-border);
  border-radius: var(--kore-radius-md);
  transition: border-color 0.15s;
}

.pub-input__field:focus {
  border-color: var(--kore-brand-blue);
}

.pub-input__field--error {
  border-color: var(--kore-error);
}

.pub-input__error {
  margin: 0;
  font-size: var(--kore-text-caption);
  color: var(--kore-error);
}
</style>
