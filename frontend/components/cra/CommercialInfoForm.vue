<template>
  <AppCard padding="lg">
    <h3 class="section-title">{{ $t('cra.commercial_title') }}</h3>
    <p class="section-hint">{{ $t('cra.commercial_hint') }}</p>
    <form class="commercial-form" @submit.prevent="$emit('submit')">
      <AppInput id="client" v-model="local.client" :label="$t('cra.client')" />
      <AppInput id="mission" v-model="local.mission" :label="$t('cra.mission')" />
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
  disabled?: boolean
  saving?: boolean
  message?: string
  isError?: boolean
}>()

defineEmits<{ submit: [] }>()

const local = reactive({ client: props.client ?? '', mission: props.mission ?? '' })

watch(
  () => [props.client, props.mission],
  () => {
    local.client = props.client ?? ''
    local.mission = props.mission ?? ''
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
</style>
