<template>
  <AppModal v-model:open="openModel" width="xl" :aria-label="$t('cra.pdf_preview')">
    <div class="pdf-preview">
      <p v-if="loading" class="muted">{{ $t('cra.loading') }}</p>
      <p v-else-if="error" class="flash flash--error" role="alert">{{ error }}</p>
      <iframe v-else-if="previewUrl" :src="previewUrl" class="pdf-preview__frame" :title="$t('cra.pdf_preview')" />
    </div>
    <template #footer>
      <AppButton variant="ghost" size="sm" @click="openModel = false">{{ $t('common.close') }}</AppButton>
      <AppButton variant="primary" size="sm" :disabled="!previewUrl" @click="emit('download')">
        {{ $t('cra.download') }}
      </AppButton>
    </template>
  </AppModal>
</template>

<script setup lang="ts">
const props = defineProps<{
  open: boolean
  loading?: boolean
  error?: string
  previewUrl?: string
}>()

const emit = defineEmits<{ 'update:open': [value: boolean]; download: [] }>()

const openModel = computed({
  get: () => props.open,
  set: (value: boolean) => emit('update:open', value)
})
</script>

<style scoped>
.pdf-preview {
  min-height: 12rem;
}

.pdf-preview__frame {
  width: 100%;
  min-height: 60vh;
  border: 1px solid var(--kore-border);
  border-radius: var(--kore-radius-md);
  background: var(--kore-surface);
}

.muted {
  margin: 0;
  color: var(--kore-text-muted);
}

.flash--error {
  color: var(--kore-error);
}

@media (max-width: 768px) {
  .pdf-preview__frame {
    min-height: 50vh;
  }
}
</style>
