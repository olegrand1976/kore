<script setup lang="ts">
import type { PreviewKind, RequestResourceKey } from '~/composables/useRequestAttachments'
import {
  REQUEST_RESOURCE,
  isIndexableAttachment,
  isPreviewable,
  previewKindFrom,
  useRequestAttachments
} from '~/composables/useRequestAttachments'
import AppCard from '~/components/ui/AppCard.vue'

const props = withDefaults(defineProps<{
  resource: RequestResourceKey
  resourceId: string
  canUpload?: boolean
  /** Nest inside another card without wrapping AppCard. */
  embedded?: boolean
  title?: string
  inputId?: string
}>(), {
  canUpload: true,
  embedded: false,
  title: '',
  inputId: ''
})

const emit = defineEmits<{ indexableDocsChanged: [boolean] }>()

const { t } = useI18n()
const { extractFetchError } = useApiError()
const {
  list,
  uploadAll,
  downloadUrl,
  fetchContent,
  pickId,
  pickFileName,
  pickMimeType
} = useRequestAttachments()

const resourceType = computed(() => REQUEST_RESOURCE[props.resource])
const heading = computed(() => props.title || t('requests.form_attachments'))
const uploadInputId = computed(
  () => props.inputId || `request-attachments-upload-${props.resource}-${props.resourceId || 'new'}`
)
const rootTag = computed(() => (props.embedded ? 'div' : AppCard))
const rootBind = computed(() =>
  props.embedded
    ? {
        class: 'request-attachments request-attachments--embedded'
      }
    : {
        padding: 'lg' as const,
        class: 'request-attachments'
      }
)

const attachments = ref<Awaited<ReturnType<typeof list>>>([])
const pending = ref(true)
const uploading = ref(false)
const errorMsg = ref('')
const files = ref<File[]>([])

const selectedId = ref('')
const previewUrl = ref('')
const previewKind = ref<PreviewKind>('none')
const previewLoading = ref(false)
const previewError = ref('')
const previewSeq = ref(0)

const emitIndexableDocs = () => {
  const ok = attachments.value.some(a => isIndexableAttachment(pickFileName(a)))
  emit('indexableDocsChanged', ok)
}

const revokePreview = () => {
  if (previewUrl.value) {
    URL.revokeObjectURL(previewUrl.value)
    previewUrl.value = ''
  }
}

const loadPreview = async (id: string) => {
  const seq = ++previewSeq.value
  revokePreview()
  previewError.value = ''
  previewKind.value = 'none'
  if (!id) return
  const att = attachments.value.find(a => pickId(a) === id)
  if (!att) return
  const name = pickFileName(att)
  const mime = pickMimeType(att)
  const kind = previewKindFrom(mime, name)
  if (seq !== previewSeq.value) return
  previewKind.value = kind
  if (kind !== 'image' && kind !== 'pdf') return
  previewLoading.value = true
  try {
    const blob = await fetchContent(id)
    if (seq !== previewSeq.value) return
    previewUrl.value = URL.createObjectURL(blob)
  } catch (e) {
    if (seq !== previewSeq.value) return
    previewError.value = extractFetchError(e) || t('requests.attachments_preview_error')
  } finally {
    if (seq === previewSeq.value) {
      previewLoading.value = false
    }
  }
}

const selectAttachment = (id: string) => {
  selectedId.value = id
  void loadPreview(id)
}

const pickDefaultSelection = () => {
  if (!attachments.value.length) {
    selectedId.value = ''
    revokePreview()
    previewKind.value = 'none'
    return
  }
  const previewable = attachments.value.find(a =>
    isPreviewable(pickMimeType(a), pickFileName(a))
  )
  const first = previewable ?? attachments.value[0]
  selectAttachment(pickId(first))
}

const load = async () => {
  if (!props.resourceId) {
    attachments.value = []
    pending.value = false
    emitIndexableDocs()
    return
  }
  pending.value = true
  errorMsg.value = ''
  try {
    attachments.value = await list(resourceType.value, props.resourceId)
    pickDefaultSelection()
    emitIndexableDocs()
  } catch (e) {
    errorMsg.value = extractFetchError(e)
  } finally {
    pending.value = false
  }
}

watch(() => props.resourceId, () => load(), { immediate: true })
onBeforeUnmount(() => revokePreview())

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
  <component :is="rootTag" v-bind="rootBind">
    <h2 class="request-attachments__title">{{ heading }}</h2>

    <p v-if="pending" class="request-attachments__muted">{{ t('common.loading') }}</p>
    <p v-else-if="errorMsg" class="request-attachments__error" role="alert">{{ errorMsg }}</p>

    <ul v-if="!pending && attachments.length" class="request-attachments__list" role="listbox">
      <li
        v-for="att in attachments"
        :key="pickId(att)"
        class="request-attachments__item"
        :class="{ 'request-attachments__item--active': selectedId === pickId(att) }"
      >
        <button
          type="button"
          class="request-attachments__select"
          role="option"
          :aria-selected="selectedId === pickId(att)"
          @click="selectAttachment(pickId(att))"
        >
          {{ pickFileName(att) }}
        </button>
        <a
          class="request-attachments__download"
          :href="downloadUrl(pickId(att))"
          target="_blank"
          rel="noopener"
        >
          {{ t('requests.attachments_download') }}
        </a>
      </li>
    </ul>
    <p v-else-if="!pending" class="request-attachments__muted">{{ t('requests.attachments_empty') }}</p>

    <section
      v-if="!pending && attachments.length"
      class="request-attachments__preview"
      :aria-label="t('requests.attachments_preview')"
    >
      <h3 class="request-attachments__preview-title">{{ t('requests.attachments_preview') }}</h3>
      <p v-if="previewLoading" class="request-attachments__muted">{{ t('common.loading') }}</p>
      <p v-else-if="previewError" class="request-attachments__error" role="alert">{{ previewError }}</p>
      <img
        v-else-if="previewKind === 'image' && previewUrl"
        :src="previewUrl"
        class="request-attachments__image"
        :alt="t('requests.attachments_preview')"
      >
      <iframe
        v-else-if="previewKind === 'pdf' && previewUrl"
        :src="previewUrl"
        class="request-attachments__frame"
        :title="t('requests.attachments_preview')"
      />
      <p v-else-if="previewKind === 'unsupported'" class="request-attachments__muted">
        {{ t('requests.attachments_preview_unsupported') }}
      </p>
    </section>

    <form v-if="canUpload" class="request-attachments__upload" @submit.prevent="onUpload">
      <AppFileUpload :id="uploadInputId" v-model="files" :label="t('requests.attachments_add')" />
      <AppButton
        variant="primary"
        size="sm"
        type="submit"
        :disabled="uploading || !files.length"
      >
        {{ uploading ? t('common.loading') : t('requests.attachments_upload') }}
      </AppButton>
    </form>
  </component>
</template>

<style scoped>
.request-attachments {
  display: grid;
  gap: var(--kore-space-md);
  margin-bottom: var(--kore-space-lg);
}

.request-attachments--embedded {
  margin-bottom: 0;
  margin-top: var(--kore-space-lg);
  padding-top: var(--kore-space-lg);
  border-top: 1px solid var(--kore-border);
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

.request-attachments__item {
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  gap: var(--kore-space-sm);
  padding: var(--kore-space-xs) var(--kore-space-sm);
  border-radius: var(--kore-radius-md);
  border: 1px solid transparent;
}

.request-attachments__item--active {
  border-color: var(--kore-brand-blue);
  background: var(--kore-bg-subtle);
}

.request-attachments__select {
  flex: 1;
  min-width: 0;
  text-align: left;
  border: 0;
  background: transparent;
  color: var(--kore-text);
  font: inherit;
  font-size: var(--kore-text-small);
  cursor: pointer;
  padding: 0;
}

.request-attachments__download {
  color: var(--kore-brand-blue);
  text-decoration: none;
  font-size: var(--kore-text-caption);
}

.request-attachments__download:hover {
  text-decoration: underline;
}

.request-attachments__preview {
  display: grid;
  gap: var(--kore-space-sm);
  padding: var(--kore-space-md);
  border: 1px solid var(--kore-border);
  border-radius: var(--kore-radius-md);
  background: var(--kore-bg-subtle);
}

.request-attachments__preview-title {
  margin: 0;
  font-size: var(--kore-text-small);
  font-weight: 600;
}

.request-attachments__image {
  max-width: 100%;
  max-height: 60vh;
  object-fit: contain;
  border-radius: var(--kore-radius-md);
  background: var(--kore-surface);
}

.request-attachments__frame {
  width: 100%;
  min-height: 60vh;
  border: 1px solid var(--kore-border);
  border-radius: var(--kore-radius-md);
  background: var(--kore-surface);
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
  color: var(--kore-error);
  font-size: var(--kore-text-small);
}

@media (max-width: 768px) {
  .request-attachments__upload :deep(.app-btn) {
    width: 100%;
  }

  .request-attachments__frame {
    min-height: 50vh;
  }

  .request-attachments__item {
    flex-direction: column;
    align-items: stretch;
  }

  .request-attachments__download {
    width: 100%;
    text-align: center;
  }
}
</style>
