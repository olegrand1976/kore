<template>
  <div>
    <AppPageHeader :title="$t('support.title')">
      <template #actions>
        <AppButton variant="primary" size="sm" @click="showForm = !showForm">
          {{ $t('support.new') }}
        </AppButton>
      </template>
    </AppPageHeader>

    <AppCard v-if="showForm" padding="lg" class="mb">
      <form class="support-form" @submit.prevent="onCreate">
        <AppInput v-model="form.applicationId" :label="$t('support.application_id')" required />
        <AppInput v-model="form.subject" :label="$t('support.col_subject')" required />
        <AppInput v-model="form.description" :label="$t('support.col_description')" multiline />
        <AppButton variant="primary" size="sm" type="submit" :disabled="busy">
          {{ $t('support.create') }}
        </AppButton>
      </form>
    </AppCard>

    <AppCard v-if="pending" padding="lg"><p class="muted">{{ $t('support.loading') }}</p></AppCard>
    <AppCard v-else-if="!rows.length" padding="lg">
      <AppEmptyState icon="inbox" :title="$t('support.empty')" />
    </AppCard>
    <AppCard v-else padding="none">
      <AppTable :columns="columns" :rows="rows" row-key="id">
        <template #cell-state="{ value }">
          <AppBadge variant="neutral">{{ supportStateLabel(String(value)) }}</AppBadge>
        </template>
        <template #cell-actions="{ row }">
          <AppButton variant="ghost" size="sm" @click="navigateTo(`/support/${row.id}`)">
            {{ $t('support.open') }}
          </AppButton>
        </template>
      </AppTable>
    </AppCard>
    <p v-if="errorMsg" class="flash flash--error" role="alert">{{ errorMsg }}</p>
  </div>
</template>

<script setup lang="ts">
definePageMeta({ layout: 'default' })

const { t } = useI18n()
const { extractFetchError } = useApiError()
const { list, create, pickId, pickSubject, pickState } = useSupport()

const pending = ref(true)
const busy = ref(false)
const showForm = ref(false)
const errorMsg = ref('')
const tickets = ref<Awaited<ReturnType<typeof list>>>([])
const form = reactive({ applicationId: '', subject: '', description: '' })

const columns = computed(() => [
  { key: 'subject', label: t('support.col_subject') },
  { key: 'state', label: t('support.col_state') },
  { key: 'actions', label: '' }
])

const rows = computed(() =>
  tickets.value.map((ticket) => ({
    id: pickId(ticket),
    subject: pickSubject(ticket),
    state: pickState(ticket)
  }))
)

const supportStateLabel = (state: string) => {
  const key = `support.state_${state}` as const
  return t(key, state)
}

const load = async () => {
  pending.value = true
  errorMsg.value = ''
  try {
    tickets.value = await list()
  } catch (e) {
    errorMsg.value = extractFetchError(e)
  } finally {
    pending.value = false
  }
}

const onCreate = async () => {
  busy.value = true
  errorMsg.value = ''
  try {
    await create(form)
    showForm.value = false
    form.subject = ''
    form.description = ''
    await load()
  } catch (e) {
    errorMsg.value = extractFetchError(e)
  } finally {
    busy.value = false
  }
}

await load()
</script>

<style scoped>
.mb { margin-bottom: var(--kore-space-lg); }
.support-form {
  display: flex;
  flex-direction: column;
  gap: var(--kore-space-md);
  max-width: var(--kore-form-max);
}
.muted { color: var(--kore-text-muted); }
.flash { margin-top: var(--kore-space-md); font-size: var(--kore-text-small); }
.flash--error { color: var(--kore-status-danger); }
@media (max-width: 768px) {
  .support-form :deep(.app-button) { width: 100%; }
}
</style>
