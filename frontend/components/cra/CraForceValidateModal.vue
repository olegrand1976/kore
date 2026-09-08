<template>
  <AppModal
    :open="open"
    width="md"
    :title-id="titleId"
    :aria-label="$t('cra.force_validate_title')"
    @update:open="$emit('update:open', $event)"
  >
    <div class="force-validate">
      <h2 :id="titleId" class="force-validate__title">{{ $t('cra.force_validate_title') }}</h2>
      <p>{{ $t('cra.force_validate_body') }}</p>
      <p class="force-validate__hint muted">{{ $t('cra.force_validate_hint') }}</p>
      <div class="force-validate__actions">
        <AppButton
          variant="ghost"
          size="sm"
          type="button"
          :disabled="validating"
          @click="$emit('update:open', false)"
        >
          {{ $t('common.cancel') }}
        </AppButton>
        <AppButton variant="primary" size="sm" :disabled="validating" @click="$emit('confirm')">
          {{ $t('cra.force_validate_confirm') }}
        </AppButton>
      </div>
    </div>
  </AppModal>
</template>

<script setup lang="ts">
defineProps<{
  open: boolean
  validating?: boolean
}>()

defineEmits<{
  confirm: []
  'update:open': [value: boolean]
}>()

const titleId = 'cra-force-validate-title'
</script>

<style scoped>
.force-validate {
  display: grid;
  gap: var(--kore-space-md);
}

.force-validate__title {
  margin: 0;
  font-size: var(--kore-text-h3);
}

.force-validate__hint {
  margin: 0;
  font-size: var(--kore-text-small);
  color: var(--kore-text-muted);
}

.force-validate__actions {
  display: flex;
  flex-wrap: wrap;
  gap: var(--kore-space-sm);
  justify-content: flex-end;
}

@media (max-width: 768px) {
  .force-validate__actions {
    flex-direction: column-reverse;
  }

  .force-validate__actions :deep(.app-btn) {
    width: 100%;
  }
}
</style>
