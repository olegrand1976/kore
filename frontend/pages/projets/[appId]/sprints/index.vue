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
              v-if="row.rawStatus === 'planned' && canWrite"
              variant="ghost"
              size="sm"
              @click="openPlan(row.id)"
            >
              {{ $t('project.plan_sprint') }}
            </AppButton>
            <AppButton
              v-if="row.rawStatus === 'planned' && canValidate"
              variant="ghost"
              size="sm"
              @click="start(row.id)"
            >
              {{ $t('project.sprint_start') }}
            </AppButton>
            <AppButton
              v-if="row.rawStatus === 'active' && canValidate"
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

    <AppModal v-model:open="showPlan" :title="$t('project.plan_sprint_title')">
      <p v-if="planPending" class="muted">{{ $t('project.loading') }}</p>
      <ul v-else class="plan-list">
        <li v-for="item in planItems" :key="item.id" class="plan-item">
          <label>
            <input v-model="selectedIds" type="checkbox" :value="item.id" />
            {{ item.subject }}
            <span class="plan-item__sp">{{ item.storyPoints ?? '—' }} SP</span>
          </label>
        </li>
      </ul>
      <p v-if="!planPending && !planItems.length" class="muted">{{ $t('project.backlog_empty') }}</p>
      <div class="projets-form__actions">
        <AppButton variant="ghost" type="button" @click="showPlan = false">{{ $t('common.cancel') }}</AppButton>
        <AppButton variant="primary" type="button" :disabled="planBusy || !selectedIds.length" @click="submitPlan">
          {{ $t('project.plan_sprint') }}
        </AppButton>
      </div>
    </AppModal>
  </div>
</template>

<script setup lang="ts">
definePageMeta({ layout: 'default' })

const route = useRoute()
const { t } = useI18n()
const appId = computed(() => String(route.params.appId ?? ''))
const { get, pickAppLabel } = useApplications()
const { listSprints, createSprint, startSprint, closeSprint, planSprint, listBacklog, pickSprintId, pickSprintName, pickDemandId } = useProject()
const { can } = usePermissions()

const canWrite = computed(() => can('project', 'E'))
const canValidate = computed(() => can('project', 'V'))

const appLabel = ref('')
const rows = ref<{ id: string; name: string; status: string; rawStatus: string; period: string }[]>([])
const showForm = ref(false)
const showPlan = ref(false)
const planPending = ref(false)
const planBusy = ref(false)
const planSprintId = ref('')
const planItems = ref<{ id: string; subject: string; storyPoints: number | null }[]>([])
const selectedIds = ref<string[]>([])
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
  rows.value = items.map((s) => {
    const raw = s.status ?? s.Status ?? ''
    return {
      id: pickSprintId(s),
      name: pickSprintName(s),
      rawStatus: raw,
      status: sprintStatusLabel(raw),
      period: `${s.startDate ?? s.StartDate ?? ''} → ${s.endDate ?? s.EndDate ?? ''}`
    }
  })
}

async function openPlan(sprintId: string) {
  planSprintId.value = sprintId
  showPlan.value = true
  planPending.value = true
  selectedIds.value = []
  try {
    const items = await listBacklog(appId.value, true)
    planItems.value = items.map((item) => ({
      id: pickDemandId(item),
      subject: item.subject ?? item.Subject ?? '',
      storyPoints: item.storyPoints ?? item.StoryPoints ?? null
    }))
  } finally {
    planPending.value = false
  }
}

async function submitPlan() {
  planBusy.value = true
  try {
    await planSprint(appId.value, planSprintId.value, selectedIds.value)
    showPlan.value = false
    await load()
  } finally {
    planBusy.value = false
  }
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
  margin-top: var(--kore-space-md);
}

.plan-list {
  list-style: none;
  margin: 0;
  padding: 0;
  display: flex;
  flex-direction: column;
  gap: var(--kore-space-sm);
  max-height: 320px;
  overflow-y: auto;
}

.plan-item label {
  display: flex;
  align-items: center;
  gap: var(--kore-space-sm);
  cursor: pointer;
}

.plan-item__sp {
  margin-left: auto;
  color: var(--kore-text-muted);
  font-size: var(--kore-text-small);
}
</style>
