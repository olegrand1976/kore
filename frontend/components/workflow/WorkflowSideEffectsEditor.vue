<script setup lang="ts">
import type { WorkflowSideEffect } from '~/composables/useWorkflowDefinition'
import {
  MAX_SIDE_EFFECTS_PER_HOOK,
  WORKFLOW_RECIPIENT_SCOPES,
  createEmptySideEffect
} from '~/composables/useWorkflowDefinition'

const props = defineProps<{
  modelValue: WorkflowSideEffect[]
  idPrefix: string
}>()

const emit = defineEmits<{
  'update:modelValue': [effects: WorkflowSideEffect[]]
}>()

const { t } = useI18n()

type UserRow = { id: string; login: string; prenom?: string; nom?: string }
type EquipeRow = { id: string; libelle: string }
type ServiceRow = { id: string; siteLabel?: string }
type ApplicationRow = { id: string; libelle: string }

const { data: usersData } = useFetch<{ data?: UserRow[] }>('/api/org/users')
const { data: equipesData } = useFetch<{ data?: EquipeRow[] }>('/api/org/equipes')
const { data: servicesData } = useFetch<{ data?: ServiceRow[] }>('/api/org/services')
const { data: applicationsData } = useFetch<{ data?: ApplicationRow[] }>('/api/org/applications')

const users = computed(() => usersData.value?.data ?? [])
const equipes = computed(() => equipesData.value?.data ?? [])
const services = computed(() => servicesData.value?.data ?? [])
const applications = computed(() => applicationsData.value?.data ?? [])

const effects = computed({
  get: () => props.modelValue ?? [],
  set: (value: WorkflowSideEffect[]) => emit('update:modelValue', value)
})

const canAdd = computed(() => effects.value.length < MAX_SIDE_EFFECTS_PER_HOOK)

const updateEffect = (index: number, patch: Partial<WorkflowSideEffect>) => {
  effects.value = effects.value.map((effect, i) => (i === index ? { ...effect, ...patch } : effect))
}

const updateRecipients = (index: number, patch: Partial<WorkflowSideEffect['recipients']>) => {
  const current = effects.value[index]
  if (!current) return
  updateEffect(index, {
    recipients: { ...current.recipients, ...patch }
  })
}

const onScopeChange = (index: number, scope: WorkflowSideEffect['recipients']['scope']) => {
  updateEffect(index, {
    recipients: {
      scope,
      userIds: scope === 'user' ? [] : undefined,
      equipeId: scope === 'equipe' ? '' : undefined,
      serviceId: scope === 'service' ? '' : undefined,
      applicationId: scope === 'application' ? '' : undefined
    }
  })
}

const toggleUser = (index: number, userId: string, checked: boolean) => {
  const current = effects.value[index]?.recipients.userIds ?? []
  const next = checked ? [...current, userId] : current.filter((id) => id !== userId)
  updateRecipients(index, { userIds: next })
}

const addEffect = () => {
  if (!canAdd.value) return
  effects.value = [...effects.value, createEmptySideEffect()]
}

const removeEffect = (index: number) => {
  effects.value = effects.value.filter((_, i) => i !== index)
}

const userLabel = (user: UserRow) => {
  const name = [user.prenom, user.nom].filter(Boolean).join(' ').trim()
  return name ? `${name} (${user.login})` : user.login
}

const serviceLabel = (service: ServiceRow) =>
  service.siteLabel ? `${service.siteLabel} — ${service.id.slice(0, 8)}` : service.id.slice(0, 8)
</script>

<template>
  <div class="wf-effects">
    <p class="settings-hint">{{ $t('workflows.effects.hint') }}</p>

    <div
      v-for="(effect, index) in effects"
      :key="`${idPrefix}-effect-${index}`"
      class="wf-effects__row"
    >
      <div class="wf-effects__header">
        <span class="wf-effects__title">{{ $t('workflows.effects.email_title', { n: index + 1 }) }}</span>
        <AppButton variant="ghost" size="sm" type="button" @click="removeEffect(index)">
          {{ $t('workflows.effects.remove') }}
        </AppButton>
      </div>

      <div class="settings-field">
        <label :for="`${idPrefix}-scope-${index}`">{{ $t('workflows.effects.recipient_scope') }}</label>
        <select
          :id="`${idPrefix}-scope-${index}`"
          :value="effect.recipients.scope"
          @change="onScopeChange(index, ($event.target as HTMLSelectElement).value as WorkflowSideEffect['recipients']['scope'])"
        >
          <option v-for="scope in WORKFLOW_RECIPIENT_SCOPES" :key="scope" :value="scope">
            {{ $t(`workflows.effects.scope_${scope}`) }}
          </option>
        </select>
      </div>

      <div v-if="effect.recipients.scope === 'user'" class="wf-effects__picker">
        <span class="wf-effects__picker-label">{{ $t('workflows.effects.pick_users') }}</span>
        <label
          v-for="user in users"
          :key="user.id"
          class="wf-effects__check"
        >
          <input
            type="checkbox"
            :checked="effect.recipients.userIds?.includes(user.id)"
            @change="toggleUser(index, user.id, ($event.target as HTMLInputElement).checked)"
          />
          <span>{{ userLabel(user) }}</span>
        </label>
      </div>

      <div v-else-if="effect.recipients.scope === 'equipe'" class="settings-field">
        <label :for="`${idPrefix}-equipe-${index}`">{{ $t('workflows.effects.pick_equipe') }}</label>
        <select
          :id="`${idPrefix}-equipe-${index}`"
          :value="effect.recipients.equipeId ?? ''"
          @change="updateRecipients(index, { equipeId: ($event.target as HTMLSelectElement).value })"
        >
          <option value="" disabled>{{ $t('workflows.effects.select_option') }}</option>
          <option v-for="equipe in equipes" :key="equipe.id" :value="equipe.id">
            {{ equipe.libelle }}
          </option>
        </select>
      </div>

      <div v-else-if="effect.recipients.scope === 'service'" class="settings-field">
        <label :for="`${idPrefix}-service-${index}`">{{ $t('workflows.effects.pick_service') }}</label>
        <select
          :id="`${idPrefix}-service-${index}`"
          :value="effect.recipients.serviceId ?? ''"
          @change="updateRecipients(index, { serviceId: ($event.target as HTMLSelectElement).value })"
        >
          <option value="" disabled>{{ $t('workflows.effects.select_option') }}</option>
          <option v-for="service in services" :key="service.id" :value="service.id">
            {{ serviceLabel(service) }}
          </option>
        </select>
      </div>

      <div v-else-if="effect.recipients.scope === 'application'" class="settings-field">
        <label :for="`${idPrefix}-application-${index}`">{{ $t('workflows.effects.pick_application') }}</label>
        <select
          :id="`${idPrefix}-application-${index}`"
          :value="effect.recipients.applicationId ?? ''"
          @change="updateRecipients(index, { applicationId: ($event.target as HTMLSelectElement).value })"
        >
          <option value="" disabled>{{ $t('workflows.effects.select_option') }}</option>
          <option v-for="app in applications" :key="app.id" :value="app.id">
            {{ app.libelle }}
          </option>
        </select>
      </div>

      <AppInput
        :id="`${idPrefix}-subject-${index}`"
        :model-value="effect.subject"
        @update:model-value="updateEffect(index, { subject: $event })"
      >
        <template #label>{{ $t('workflows.effects.subject') }}</template>
      </AppInput>

      <div class="settings-field">
        <label :for="`${idPrefix}-body-${index}`">{{ $t('workflows.effects.body') }}</label>
        <textarea
          :id="`${idPrefix}-body-${index}`"
          class="wf-effects__textarea"
          rows="3"
          :value="effect.bodyTemplate"
          @input="updateEffect(index, { bodyTemplate: ($event.target as HTMLTextAreaElement).value })"
        />
        <p class="settings-hint">{{ $t('workflows.effects.template_vars') }}</p>
      </div>
    </div>

    <AppButton variant="ghost" size="sm" type="button" :disabled="!canAdd" @click="addEffect">
      {{ $t('workflows.effects.add_email') }}
    </AppButton>
  </div>
</template>

<style scoped>
.wf-effects {
  display: grid;
  gap: var(--kore-space-md);
}

.wf-effects__row {
  display: grid;
  gap: var(--kore-space-sm);
  padding: var(--kore-space-md);
  border: 1px dashed var(--kore-border);
  border-radius: var(--kore-radius-md);
  background: var(--kore-bg-elevated);
}

.wf-effects__header {
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  justify-content: space-between;
  gap: var(--kore-space-sm);
}

.wf-effects__title {
  font-size: var(--kore-text-small);
  font-weight: 600;
}

.settings-field {
  display: flex;
  flex-direction: column;
  gap: 0.35rem;
}

.settings-field label,
.settings-field select,
.wf-effects__picker-label {
  font-size: var(--kore-text-small);
}

.settings-field select {
  padding: 0.5rem 0.75rem;
  border: 1px solid var(--kore-border);
  border-radius: var(--kore-radius-sm);
  background: var(--kore-bg);
  color: var(--kore-text);
}

.wf-effects__picker {
  display: grid;
  gap: 0.35rem;
  max-height: 12rem;
  overflow: auto;
  padding: var(--kore-space-sm);
  border: 1px solid var(--kore-border);
  border-radius: var(--kore-radius-sm);
}

.wf-effects__check {
  display: inline-flex;
  align-items: center;
  gap: 0.35rem;
  font-size: var(--kore-text-small);
}

.wf-effects__textarea {
  width: 100%;
  max-width: 100%;
  padding: 0.5rem 0.75rem;
  border: 1px solid var(--kore-border);
  border-radius: var(--kore-radius-sm);
  background: var(--kore-bg);
  color: var(--kore-text);
  font-family: inherit;
  font-size: var(--kore-text-small);
  resize: vertical;
}

.settings-hint {
  margin: 0;
  color: var(--kore-text-muted);
  font-size: var(--kore-text-small);
}

@media (max-width: 768px) {
  .wf-effects :deep(.app-button) {
    width: 100%;
  }
}
</style>
