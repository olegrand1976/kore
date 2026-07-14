<script setup lang="ts">
import type { RequestResourceKey } from '~/composables/useRequestAttachments'
import { REQUEST_RESOURCE, useRequestAttachments } from '~/composables/useRequestAttachments'

const props = withDefaults(defineProps<{
  resource: RequestResourceKey
  resourceId: string
  canUpload?: boolean
}>(), {
  canUpload: true
})

const { t } = useI18n()
const { extractFetchError } = useApiError()
const { list, uploadAll, downloadUrl, pickId, pickFileName } = useRequestAttachments()

const resourceType = computed(() => REQUEST_RESOURCE[props.resource])

const attachments = ref<Awaited<ReturnType<typeof list>>>([])
const pending = ref(true)
const uploading = ref(false)
const errorMsg = ref('')
const files = ref<File[]>([])

const load = async () => {
  if (!props.resourceId) {
    attachments.value = []
    pending.value = false
    return
  }
  pending.value = true
  errorMsg.value = ''
  try {
    attachments.value = await list(resourceType.value, props.resourceId)
  } catch (e) {
    errorMsg.value = extractFetchError(e)
  } finally {
    pending.value = false
  }
}

watch(() => props.resourceId, () => load(), { immediate: true })

const onUpload = async () => {
  if (!files.value.length || !props.resourceId) return
  uploading.value = true
  errorMsg.value = ''
  try {
    await uploadAll(resourceType.value, props.resourceId, files.value)
    files.value = []
    await load()
  } catch (e) {
    errorMsg.value = extractFetchError(e)
  } finally {
    uploading.value = false
  }
}
</script>

<template>
  <AppCard padding="lg" class="request-attachments">
    <h2 class="request-attachments__title">{{ t('requests.form_attachments') }}</h2>

    <p v-if="pending" class="request-attachments__muted">{{ t('common.loading') }}</p>
    <p v-else-if="errorMsg" class="request-attachments__error" role="alert">{{ errorMsg }}</p>

    <ul v-if="!pending && attachments.length" class="request-attachments__list">
      <li v-for="att in attachments" :key="pickId(att)">
        <a :href="downloadUrl(pickId(att))" target="_blank" rel="noopener">
          {{ pickFileName(att) }}
        </a>
      </li>
    </ul>
    <p v-else-if="!pending" class="request-attachments__muted">{{ t('requests.attachments_empty') }}</p>

    <form v-if="canUpload" class="request-attachments__upload" @submit.prevent="onUpload">
      <AppFileUpload id="request-attachments-upload" v-model="files" :label="t('requests.attachments_add')" />
      <AppButton
        variant="primary"
        size="sm"
        type="submit"
        :disabled="uploading || !files.length"
      >
        {{ uploading ? t('common.loading') : t('requests.attachments_upload') }}
      </AppButton>
    </form>
  </AppCard>
</template>

<style scoped>
.request-attachments {
  display: grid;
  gap: var(--kore-space-md);
  margin-bottom: var(--kore-space-lg);
}

.request-attachments__title {
  margin: 0;
  font-size: var(--kore-text-h3);
}

.request-attachments__list {
  margin: 0;
  padding: 0;
  list-style: none;
  display: grid;
  gap: var(--kore-space-xs);
}

.request-attachments__list a {
  color: var(--kore-brand-blue);
  text-decoration: none;
  font-size: var(--kore-text-small);
}

.request-attachments__list a:hover {
  text-decoration: underline;
}

.request-attachments__upload {
  display: grid;
  gap: var(--kore-space-md);
  max-width: var(--kore-form-wide-max);
}

.request-attachments__muted {
  margin: 0;
  color: var(--kore-text-muted);
  font-size: var(--kore-text-small);
}

.request-attachments__error {
  margin: 0;
  color: var(--kore-status-danger);
  font-size: var(--kore-text-small);
}

@media (max-width: 768px) {
  .request-attachments__upload :deep(.app-button) {
    width: 100%;
  }
}
</style>
