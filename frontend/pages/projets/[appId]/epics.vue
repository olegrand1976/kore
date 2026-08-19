<template>
  <div>
    <AppPageHeader :title="$t('project.epics_title')" :subtitle="appLabel">
      <template #actions>
        <AppButton v-if="canWrite" variant="primary" size="sm" @click="showForm = true">{{ $t('project.epic_new') }}</AppButton>
      </template>
    </AppPageHeader>

    <AppCard padding="lg">
      <AppTable :columns="columns" :rows="rows" :empty-title="$t('project.epics_empty')" row-key="id">
        <template #cell-actions="{ row }">
          <AppButton
            v-if="canWrite && row.rawStatus !== 'done'"
            variant="ghost"
            size="sm"
            @click="finishEpic(row.id)"
          >
            {{ $t('project.epic_finish') }}
          </AppButton>
        </template>
      </AppTable>
    </AppCard>

    <AppModal v-model:open="showForm" :title="$t('project.epic_new')">
      <form class="projets-form" @submit.prevent="submit">
        <AppInput v-model="title" :label="$t('project.field_title')" required />
        <AppInput v-model="description" :label="$t('project.field_description')" />
        <div class="projets-form__actions">
          <AppButton variant="ghost" type="button" @click="showForm = false">{{ $t('common.cancel') }}</AppButton>
          <AppButton variant="primary" type="submit">{{ $t('common.save') }}</AppButton>
        </div>
      </form>
    </AppModal>
  </div>
</template>

<script setup lang="ts">
definePageMeta({ layout: 'default' })

const route = useRoute()
const { t } = useI18n()
const appId = computed(() => String(route.params.appId ?? ''))
const { get, pickAppLabel } = useApplications()
const { listEpics, createEpic, updateEpic, pickEpicId, pickEpicTitle } = useProject()
const { can } = usePermissions()

const canWrite = computed(() => can('project', 'E'))

const appLabel = ref('')
const rows = ref<{ id: string; title: string; status: string; rawStatus: string }[]>([])
const showForm = ref(false)
const title = ref('')
const description = ref('')

const columns = computed(() => [
  { key: 'title', label: t('project.field_title') },
  { key: 'status', label: t('project.col_status') },
  { key: 'actions', label: t('common.actions') }
])

function epicStatusLabel(status: string) {
  switch (status) {
    case 'draft':
      return t('project.status.epic_draft')
    case 'active':
      return t('project.status.epic_active')
    case 'done':
      return t('project.status.epic_done')
    default:
      return status
  }
}

async function load() {
  const app = await get(appId.value)
  appLabel.value = pickAppLabel(app)
  const items = await listEpics(appId.value)
  rows.value = items.map((e) => {
    const raw = e.status ?? e.Status ?? ''
    return {
      id: pickEpicId(e),
      title: pickEpicTitle(e),
      rawStatus: raw,
      status: epicStatusLabel(raw)
    }
  })
}

async function finishEpic(id: string) {
  await updateEpic(appId.value, id, { status: 'done' })
  await load()
}

async function submit() {
  await createEpic(appId.value, { title: title.value, description: description.value })
  showForm.value = false
  title.value = ''
  description.value = ''
  await load()
}

onMounted(load)
</script>

<style scoped>
.projets-form {
  display: flex;
  flex-direction: column;
  gap: var(--kore-space-md);
}

.projets-form__actions {
  display: flex;
  justify-content: flex-end;
  gap: var(--kore-space-sm);
}
</style>
