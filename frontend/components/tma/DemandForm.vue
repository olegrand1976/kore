<script setup lang="ts">
const emit = defineEmits<{ submit: [payload: { applicationId: string; subject: string; requiresChefGate: boolean }] }>()

const { t } = useI18n()
const { list: listBudgets } = useBudget()

const subject = ref('')
const applicationId = ref('')
const requiresChefGate = ref(false)
const loadingApps = ref(true)

onMounted(async () => {
  try {
    const budgets = await listBudgets()
    const first = budgets[0]
    applicationId.value = String(first?.applicationId ?? first?.ApplicationID ?? '')
  } finally {
    loadingApps.value = false
  }
})

const onSubmit = () => {
  if (!subject.value.trim() || !applicationId.value) return
  emit('submit', {
    applicationId: applicationId.value,
    subject: subject.value.trim(),
    requiresChefGate: requiresChefGate.value
  })
}
</script>

<template>
  <form class="demand-form" @submit.prevent="onSubmit">
    <AppInput
      id="tma-subject"
      v-model="subject"
      :label="$t('tma.form_subject')"
      required
    />
    <AppInput
      id="tma-app"
      v-model="applicationId"
      :label="$t('tma.form_application')"
      :disabled="loadingApps"
      required
    />
    <label class="demand-form__check">
      <input v-model="requiresChefGate" type="checkbox" />
      {{ $t('tma.form_chef_gate') }}
    </label>
    <AppButton variant="primary" size="sm" type="submit" :disabled="!subject.trim() || !applicationId">
      {{ $t('tma.form_submit') }}
    </AppButton>
  </form>
</template>

<style scoped>
.demand-form {
  display: grid;
  gap: var(--kore-space-md);
  max-width: var(--kore-form-wide-max);
}

.demand-form__check {
  display: flex;
  align-items: center;
  gap: var(--kore-space-sm);
  font-size: var(--kore-text-small);
}
</style>
