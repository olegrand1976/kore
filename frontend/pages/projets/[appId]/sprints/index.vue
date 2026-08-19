<template>
  <div>
    <AppPageHeader :title="$t('project.sprints_title')" :subtitle="appLabel">
      <template #actions>
        <AppButton v-if="canWrite" variant="primary" size="sm" @click="showForm = true">{{ $t('project.sprint_new') }}</AppButton>
      </template>
    </AppPageHeader>

    <AppCard padding="lg">
      <AppTable :columns="columns" :rows="rows" :empty-title="$t('project.sprints_empty')" row-key="id">
        <template #cell-actions="{ row }">
          <div class="sprints-actions">
            <AppButton
              v-if="row.status === 'planned' && canValidate"
              variant="ghost"
              size="sm"
              @click="start(row.id)"
            >
              {{ $t('project.sprint_start') }}
            </AppButton>
            <AppButton
              v-if="row.status === 'active' && canValidate"
              variant="ghost"
              size="sm"
              @click="close(row.id)"
            >
              {{ $t('project.sprint_close') }}
            </AppButton>
            <NuxtLink :to="`/projets/${appId}/sprints/${row.id}`" class="sprints-link">{{ $t('project.sprint_board') }}</NuxtLink>
          </div>
        </template>
      </AppTable>
    </AppCard>

    <AppModal v-model:open="showForm" :title="$t('project.sprint_new')">
      <form class="projets-form" @submit.prevent="submit">
        <AppInput v-model="name" :label="$t('project.field_name')" required />
        <AppInput v-model="goal" :label="$t('project.field_goal')" />
        <AppInput v-model="startDate" type="date" :label="$t('project.field_start')" required />
        <AppInput v-model="endDate" type="date" :label="$t('project.field_end')" required />
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
const { listSprints, createSprint, startSprint, closeSprint, pickSprintId, pickSprintName } = useProject()
const { can } = usePermissions()

const canWrite = computed(() => can('project', 'E'))
const canValidate = computed(() => can('project', 'V'))

const appLabel = ref('')
const rows = ref<{ id: string; name: string; status: string; period: string }[]>([])
const showForm = ref(false)
const name = ref('')
const goal = ref('')
const startDate = ref('')
const endDate = ref('')

const columns = computed(() => [
  { key: 'name', label: t('project.field_name') },
  { key: 'period', label: t('project.col_period') },
  { key: 'status', label: t('project.col_status') },
  { key: 'actions', label: t('common.actions') }
])

function sprintStatusLabel(status: string) {
  switch (status) {
    case 'planned':
      return t('project.status.sprint_planned')
    case 'active':
      return t('project.status.sprint_active')
    case 'closed':
      return t('project.status.sprint_closed')
    default:
      return status
  }
}

async function load() {
  const app = await get(appId.value)
  appLabel.value = pickAppLabel(app)
  const items = await listSprints(appId.value)
  rows.value = items.map((s) => ({
    id: pickSprintId(s),
    name: pickSprintName(s),
    status: sprintStatusLabel(s.status ?? s.Status ?? ''),
    period: `${s.startDate ?? s.StartDate ?? ''} → ${s.endDate ?? s.EndDate ?? ''}`
  }))
}

async function submit() {
  await createSprint(appId.value, {
    name: name.value,
    goal: goal.value,
    startDate: startDate.value,
    endDate: endDate.value
  })
  showForm.value = false
  await load()
}

async function start(id: string) {
  await startSprint(appId.value, id)
  await load()
}

async function close(id: string) {
  await closeSprint(appId.value, id)
  await load()
}

onMounted(load)
</script>

<style scoped>
.sprints-actions {
  display: flex;
  flex-wrap: wrap;
  gap: var(--kore-space-sm);
  align-items: center;
}

.sprints-link {
  color: var(--kore-link);
  font-size: var(--kore-text-small);
}

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
