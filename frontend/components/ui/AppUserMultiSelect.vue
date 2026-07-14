<script setup lang="ts">
const props = defineProps<{
  modelValue: string[]
  label?: string
  id?: string
  required?: boolean
}>()

const emit = defineEmits<{ 'update:modelValue': [value: string[]] }>()

const { list, pickUserId, pickUserLogin, pickUserActive } = useUsers()

const users = ref<Awaited<ReturnType<typeof list>>>([])
const loading = ref(true)

onMounted(async () => {
  try {
    users.value = (await list()).filter((user) => pickUserActive(user))
  } finally {
    loading.value = false
  }
})

const isChecked = (userId: string) => props.modelValue.includes(userId)

const toggle = (userId: string) => {
  const next = new Set(props.modelValue)
  if (next.has(userId)) next.delete(userId)
  else next.add(userId)
  emit('update:modelValue', [...next])
}
</script>

<template>
  <div class="app-user-multi-select">
    <p v-if="label" :id="`${id}-label`" class="app-user-multi-select__label">{{ label }}</p>
    <p v-if="loading" class="app-user-multi-select__hint">{{ $t('common.loading') }}</p>
    <ul
      v-else
      class="app-user-multi-select__list"
      role="group"
      :aria-labelledby="label ? `${id}-label` : undefined"
      :aria-required="required"
    >
      <li v-for="user in users" :key="pickUserId(user)">
        <label class="app-user-multi-select__item">
          <input
            type="checkbox"
            :checked="isChecked(pickUserId(user))"
            @change="toggle(pickUserId(user))"
          />
          <span>{{ pickUserLogin(user) }}</span>
        </label>
      </li>
    </ul>
  </div>
</template>

<style scoped>
.app-user-multi-select {
  display: flex;
  flex-direction: column;
  gap: var(--kore-space-xs);
}

.app-user-multi-select__label {
  margin: 0;
  font-size: var(--kore-text-small);
  color: var(--kore-text-muted);
  font-weight: 500;
}

.app-user-multi-select__hint {
  margin: 0;
  font-size: var(--kore-text-caption);
  color: var(--kore-text-muted);
}

.app-user-multi-select__list {
  margin: 0;
  padding: var(--kore-space-sm);
  list-style: none;
  display: flex;
  flex-direction: column;
  gap: var(--kore-space-xs);
  max-height: 14rem;
  overflow-y: auto;
  border: 1px solid var(--kore-border);
  border-radius: var(--kore-radius-md);
  background: var(--kore-bg-elevated);
}

.app-user-multi-select__item {
  display: flex;
  align-items: center;
  gap: var(--kore-space-sm);
  font-size: var(--kore-text-small);
  cursor: pointer;
}
</style>
