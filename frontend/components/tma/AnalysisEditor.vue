<script setup lang="ts">
type AnalysisFields = {
  functional: string
  technical: string
  risks: string
  testScenario: string
}

type SectionKey = keyof AnalysisFields

const props = defineProps<{
  analysis: AnalysisFields
  disabled?: boolean
  demandId?: string
  subject?: string
  applicationId?: string
  hasIndexableDocs?: boolean
}>()

const emit = defineEmits<{ save: [AnalysisFields] }>()

const { t } = useI18n()
const { generateAnalysisDraft, generateAnalysisSection, extractFetchError } = useAi()

const local = reactive({ ...props.analysis })
const generating = ref(false)
const draftVisible = ref(false)
const errorMsg = ref('')
const sectionBusy = ref<Partial<Record<SectionKey, boolean>>>({})
const sectionError = ref<Partial<Record<SectionKey, string>>>({})
const sectionPrompt = reactive<Record<SectionKey, string>>({
  functional: '',
  technical: '',
  risks: '',
  testScenario: ''
})
const sectionUseRAG = reactive<Record<SectionKey, boolean>>({
  functional: false,
  technical: false,
  risks: false,
  testScenario: false
})
const sectionGenerated = reactive<Record<SectionKey, boolean>>({
  functional: false,
  technical: false,
  risks: false,
  testScenario: false
})

const sections: Array<{ key: SectionKey; labelKey: string; fieldId: string }> = [
  { key: 'functional', labelKey: 'tma.analysis_functional', fieldId: 'analysis-functional' },
  { key: 'technical', labelKey: 'tma.analysis_technical', fieldId: 'analysis-technical' },
  { key: 'risks', labelKey: 'tma.analysis_risks', fieldId: 'analysis-risks' },
  { key: 'testScenario', labelKey: 'tma.analysis_tests', fieldId: 'analysis-tests' }
]

watch(
  () => props.analysis,
  (value) => {
    Object.assign(local, value)
  },
  { deep: true }
)

watch(
  () => props.hasIndexableDocs,
  (ok) => {
    if (!ok) {
      ;(Object.keys(sectionUseRAG) as SectionKey[]).forEach((k) => {
        sectionUseRAG[k] = false
      })
    }
  }
)

const onSave = () => emit('save', { ...local })

const onGenerate = async () => {
  if (!props.demandId) return
  errorMsg.value = ''
  generating.value = true
  draftVisible.value = false
  try {
    const res = await generateAnalysisDraft({
      demandId: props.demandId,
      subject: props.subject,
      applicationId: props.applicationId
    })
    Object.assign(local, res.draft)
    draftVisible.value = true
  } catch (err) {
    errorMsg.value = extractFetchError(err)
  } finally {
    generating.value = false
  }
}

const dismissDraft = () => {
  draftVisible.value = false
}

const onGenerateSection = async (key: SectionKey) => {
  if (!props.demandId || props.disabled) return
  const prompt = sectionPrompt[key].trim()
  if (!prompt) {
    sectionError.value[key] = t('ai.section_prompt_required')
    return
  }
  sectionError.value[key] = ''
  sectionBusy.value[key] = true
  try {
    const res = await generateAnalysisSection({
      demandId: props.demandId,
      section: key,
      prompt,
      useRAG: sectionUseRAG[key] && !!props.hasIndexableDocs,
      subject: props.subject
    })
    local[key] = res.text
    sectionGenerated[key] = true
  } catch (err) {
    sectionError.value[key] = extractFetchError(err)
  } finally {
    sectionBusy.value[key] = false
  }
}
</script>

<template>
  <div class="analysis-editor">
    <div v-if="!disabled && demandId" class="analysis-editor__toolbar">
      <AppButton variant="secondary" size="sm" type="button" :disabled="generating || disabled" @click="onGenerate">
        {{ generating ? $t('ai.generating') : $t('ai.generate_draft') }}
      </AppButton>
    </div>
    <p v-if="errorMsg" class="flash flash--error" role="alert">{{ errorMsg }}</p>
    <div v-if="draftVisible" class="analysis-editor__preview">
      <AppAiBadge variant="generated" />
      <p class="analysis-editor__disclaimer">{{ $t('ai.disclaimer') }}</p>
      <div class="analysis-editor__actions">
        <AppButton variant="ghost" size="sm" type="button" @click="dismissDraft">{{ $t('ai.reject') }}</AppButton>
      </div>
    </div>

    <section
      v-for="sec in sections"
      :key="sec.key"
      class="analysis-editor__section"
    >
      <label :for="sec.fieldId" class="analysis-editor__label">{{ $t(sec.labelKey) }}</label>
      <textarea
        :id="sec.fieldId"
        v-model="local[sec.key]"
        class="analysis-editor__textarea"
        rows="5"
        :disabled="disabled"
      />
      <div v-if="!disabled && demandId" class="analysis-editor__section-ai">
        <AppInput
          :id="`${sec.fieldId}-prompt`"
          v-model="sectionPrompt[sec.key]"
          :label="$t('ai.section_prompt')"
          :placeholder="$t('ai.section_prompt_placeholder')"
          :disabled="disabled || !!sectionBusy[sec.key]"
        />
        <label class="analysis-editor__rag">
          <input
            v-model="sectionUseRAG[sec.key]"
            type="checkbox"
            :disabled="disabled || !hasIndexableDocs || !!sectionBusy[sec.key]"
          >
          <span>{{ $t('ai.use_rag') }}</span>
        </label>
        <p v-if="!hasIndexableDocs" class="analysis-editor__hint">{{ $t('ai.rag_no_docs') }}</p>
        <div class="analysis-editor__section-actions">
          <AppButton
            variant="secondary"
            size="sm"
            type="button"
            :disabled="disabled || !!sectionBusy[sec.key]"
            @click="onGenerateSection(sec.key)"
          >
            {{ sectionBusy[sec.key] ? $t('ai.generating') : $t('ai.section_generate') }}
          </AppButton>
          <AppAiBadge v-if="sectionGenerated[sec.key]" variant="generated" />
        </div>
        <p v-if="sectionError[sec.key]" class="flash flash--error" role="alert">{{ sectionError[sec.key] }}</p>
        <p v-if="sectionGenerated[sec.key]" class="analysis-editor__disclaimer">{{ $t('ai.disclaimer') }}</p>
      </div>
    </section>

    <AppButton v-if="!disabled" variant="secondary" size="sm" type="button" @click="onSave">
      {{ $t('tma.analysis_save') }}
    </AppButton>
  </div>
</template>

<style scoped>
.analysis-editor {
  display: grid;
  gap: var(--kore-space-md);
}

.analysis-editor__toolbar {
  display: flex;
  flex-wrap: wrap;
  gap: var(--kore-space-sm);
}

.analysis-editor__preview {
  display: grid;
  gap: var(--kore-space-xs);
  padding: var(--kore-space-sm);
  border-radius: var(--kore-radius-md);
  background: var(--kore-bg-subtle);
}

.analysis-editor__disclaimer {
  margin: 0;
  font-size: var(--kore-text-caption);
  color: var(--kore-text-muted);
}

.analysis-editor__actions,
.analysis-editor__section-actions {
  display: flex;
  flex-wrap: wrap;
  gap: var(--kore-space-sm);
  align-items: center;
}

.analysis-editor__section {
  display: grid;
  gap: var(--kore-space-sm);
  padding-bottom: var(--kore-space-md);
  border-bottom: 1px solid var(--kore-border);
}

.analysis-editor__label {
  font-size: var(--kore-text-small);
  color: var(--kore-text-muted);
  font-weight: 500;
}

.analysis-editor__textarea {
  width: 100%;
  min-height: 6rem;
  padding: 0.75rem 1rem;
  font-family: var(--kore-font);
  font-size: var(--kore-text-body);
  color: var(--kore-text);
  background: var(--kore-bg-elevated);
  border: 1px solid var(--kore-border);
  border-radius: var(--kore-radius-md);
  resize: vertical;
}

.analysis-editor__textarea:disabled {
  opacity: 0.65;
  cursor: not-allowed;
}

.analysis-editor__section-ai {
  display: grid;
  gap: var(--kore-space-sm);
  max-width: var(--kore-form-wide-max);
}

.analysis-editor__rag {
  display: flex;
  align-items: center;
  gap: var(--kore-space-xs);
  font-size: var(--kore-text-small);
  color: var(--kore-text);
}

.analysis-editor__hint {
  margin: 0;
  font-size: var(--kore-text-caption);
  color: var(--kore-text-muted);
}

.flash--error {
  margin: 0;
  color: var(--kore-error);
  font-size: var(--kore-text-small);
}

@media (max-width: 768px) {
  .analysis-editor__toolbar :deep(.app-button),
  .analysis-editor__section-actions :deep(.app-button) {
    width: 100%;
  }
}
</style>
