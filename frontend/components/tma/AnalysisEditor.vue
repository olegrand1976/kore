<script setup lang="ts">
const props = defineProps<{
  analysis: {
    functional: string
    technical: string
    risks: string
    testScenario: string
  }
  disabled?: boolean
  demandId?: string
  subject?: string
  applicationId?: string
}>()

const emit = defineEmits<{ save: [typeof props.analysis] }>()

const { t } = useI18n()
const { generateAnalysisDraft, extractFetchError } = useAi()

const local = reactive({ ...props.analysis })
const generating = ref(false)
const draftVisible = ref(false)
const errorMsg = ref('')

watch(
  () => props.analysis,
  (value) => {
    Object.assign(local, value)
  },
  { deep: true }
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
    <AppInput id="analysis-functional" v-model="local.functional" :label="$t('tma.analysis_functional')" :disabled="disabled" />
    <AppInput id="analysis-technical" v-model="local.technical" :label="$t('tma.analysis_technical')" :disabled="disabled" />
    <AppInput id="analysis-risks" v-model="local.risks" :label="$t('tma.analysis_risks')" :disabled="disabled" />
    <AppInput id="analysis-tests" v-model="local.testScenario" :label="$t('tma.analysis_tests')" :disabled="disabled" />
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

.analysis-editor__actions {
  display: flex;
  gap: var(--kore-space-sm);
}

@media (max-width: 768px) {
  .analysis-editor__toolbar :deep(.app-button) {
    width: 100%;
  }
}
</style>
