<script setup lang="ts">
defineProps<{
  modelValue: string
  label?: string
  type?: string
  placeholder?: string
  required?: boolean
  error?: string
  id?: string
  list?: string
  tooltip?: string
  min?: string | number
  max?: string | number
  step?: string | number
}>()

defineEmits<{ 'update:modelValue': [value: string] }>()
</script>

<template>
  <div class="app-input">
    <label v-if="label" :for="id" class="app-input__label">{{ label }}</label>
    <input
      :id="id"
      class="app-input__field"
      :class="{ 'app-input__field--error': !!error }"
      :type="type || 'text'"
      :value="modelValue"
      :placeholder="placeholder"
      :required="required"
      :list="list"
      :min="min"
      :max="max"
      :step="step"
      :title="tooltip || undefined"
      @input="$emit('update:modelValue', ($event.target as HTMLInputElement).value)"
    />
    <p v-if="error" class="app-input__error" role="alert">{{ error }}</p>
  </div>
</template>

<style scoped>
.app-input { display: flex; flex-direction: column; gap: var(--kore-space-xs); }
.app-input__label { font-size: var(--kore-text-small); color: var(--kore-text-muted); font-weight: 500; }
.app-input__field {
  padding: 0.75rem 1rem;
  font-family: var(--kore-font);
  font-size: var(--kore-text-body);
  color: var(--kore-text);
  background: var(--kore-bg-elevated);
  border: 1px solid var(--kore-border);
  border-radius: var(--kore-radius-md);
}
.app-input__field--error { border-color: var(--kore-error); }
.app-input__error { margin: 0; font-size: var(--kore-text-caption); color: var(--kore-error); }
</style>
