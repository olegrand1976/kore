<template>
  <AppCard padding="lg">
    <h3 class="section-title">{{ $t('cra.commercial_title') }}</h3>
    <p class="section-hint">{{ $t('cra.commercial_hint') }}</p>
    <form class="commercial-form" @submit.prevent="$emit('submit')">
      <AppInput id="client" v-model="local.client" :label="$t('cra.client')" />
      <AppInput id="mission" v-model="local.mission" :label="$t('cra.mission')" />
      <AppInput id="description" v-model="local.description" :label="$t('cra.commercial_description')" />
      <AppInput id="technologies" v-model="technologiesText" :label="$t('cra.commercial_technologies')" />
      <AppInput id="lieu" v-model="local.lieu" :label="$t('cra.commercial_lieu')" />
      <AppInput id="responsable" v-model="local.responsableClient" :label="$t('cra.commercial_responsable')" />
      <AppButton variant="primary" size="sm" type="submit" :disabled="disabled || saving">
        {{ $t('cra.save_commercial') }}
      </AppButton>
    </form>
    <p v-if="message" class="flash" :class="{ 'flash--error': isError }" role="status">{{ message }}</p>
  </AppCard>
</template>

<script setup lang="ts">
const props = defineProps<{
  client?: string
  mission?: string
  description?: string
  technologies?: string[]
  lieu?: string
  responsableClient?: string
  disabled?: boolean
  saving?: boolean
  message?: string
  isError?: boolean
}>()

defineEmits<{ submit: [] }>()

const local = reactive({
  client: props.client ?? '',
  mission: props.mission ?? '',
  description: props.description ?? '',
  technologies: [...(props.technologies ?? [])],
  lieu: props.lieu ?? '',
  responsableClient: props.responsableClient ?? ''
})

const technologiesText = computed({
  get: () => local.technologies.join(', '),
  set: (value: string) => {
    local.technologies = value.split(',').map((s) => s.trim()).filter(Boolean)
  }
})

watch(
  () => [props.client, props.mission, props.description, props.technologies, props.lieu, props.responsableClient],
  () => {
    local.client = props.client ?? ''
    local.mission = props.mission ?? ''
    local.description = props.description ?? ''
    local.technologies = [...(props.technologies ?? [])]
    local.lieu = props.lieu ?? ''
    local.responsableClient = props.responsableClient ?? ''
  }
)

defineExpose({ local })
</script>

<style scoped>
.section-title {
  margin: 0 0 var(--kore-space-xs);
  font-size: var(--kore-text-h3);
}

.section-hint {
  margin: 0 0 var(--kore-space-lg);
  font-size: var(--kore-text-small);
  color: var(--kore-text-muted);
}

.commercial-form {
  display: flex;
  flex-direction: column;
  gap: var(--kore-space-md);
  max-width: var(--kore-form-max);
}

.flash {
  margin-top: var(--kore-space-md);
  font-size: var(--kore-text-small);
  color: var(--kore-success);
}

.flash--error { color: var(--kore-error); }

@media (max-width: 768px) {
  .commercial-form {
    max-width: none;
  }
}
</style>
