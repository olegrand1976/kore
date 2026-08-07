<script setup lang="ts">
export type ClientContactOption = {
  id: string
  nom?: string
  prenom?: string
  email?: string
  role?: string
  telephone?: string
}

const props = defineProps<{
  modelValue: string[]
  contacts: ClientContactOption[]
  label?: string
  id?: string
  emptyHint?: string
}>()

const emit = defineEmits<{ 'update:modelValue': [value: string[]] }>()

const { t } = useI18n()

const displayName = (c: ClientContactOption) => {
  const name = [c.prenom, c.nom].filter(Boolean).join(' ').trim()
  if (name) return name
  if (c.email?.trim()) return c.email.trim()
  return c.id
}

const isChecked = (contactId: string) => props.modelValue.includes(contactId)

const toggle = (contactId: string) => {
  const next = new Set(props.modelValue)
  if (next.has(contactId)) next.delete(contactId)
  else next.add(contactId)
  emit('update:modelValue', [...next])
}
</script>

<template>
  <div class="client-contact-multi-select">
    <p v-if="label" :id="`${id}-label`" class="client-contact-multi-select__label">{{ label }}</p>
    <p v-if="!contacts.length" class="client-contact-multi-select__hint">
      {{ emptyHint || t('missions.contacts_empty_hint') }}
    </p>
    <ul
      v-else
      class="client-contact-multi-select__list"
      role="group"
      :aria-labelledby="label ? `${id}-label` : undefined"
    >
      <li v-for="contact in contacts" :key="contact.id">
        <label class="client-contact-multi-select__item">
          <input
            type="checkbox"
            :checked="isChecked(contact.id)"
            @change="toggle(contact.id)"
          />
          <span>
            <strong>{{ displayName(contact) }}</strong>
            <span v-if="contact.role" class="client-contact-multi-select__meta"> — {{ contact.role }}</span>
            <span v-if="contact.email" class="client-contact-multi-select__meta"> · {{ contact.email }}</span>
          </span>
        </label>
      </li>
    </ul>
  </div>
</template>

<style scoped>
.client-contact-multi-select {
  display: flex;
  flex-direction: column;
  gap: var(--kore-space-xs);
}

.client-contact-multi-select__label {
  margin: 0;
  font-size: var(--kore-text-small);
  color: var(--kore-text-muted);
  font-weight: 500;
}

.client-contact-multi-select__hint {
  margin: 0;
  font-size: var(--kore-text-caption);
  color: var(--kore-text-muted);
}

.client-contact-multi-select__list {
  margin: 0;
  padding: 0;
  list-style: none;
  display: grid;
  gap: var(--kore-space-xs);
}

.client-contact-multi-select__item {
  display: flex;
  align-items: flex-start;
  gap: var(--kore-space-sm);
  cursor: pointer;
  font-size: var(--kore-text-small);
  color: var(--kore-text);
}

.client-contact-multi-select__meta {
  color: var(--kore-text-muted);
  font-weight: 400;
}
</style>
