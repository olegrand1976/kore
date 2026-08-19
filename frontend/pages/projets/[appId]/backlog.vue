<template>
  <div>
    <AppPageHeader :title="terms.backlog" :subtitle="appLabel">
      <template #actions>
        <AppButton variant="ghost" size="sm" @click="navigateTo('/projets')">{{ $t('project.back') }}</AppButton>
      </template>
    </AppPageHeader>

    <p v-if="errorMsg" class="flash flash--error" role="alert">{{ errorMsg }}</p>

    <AppCard padding="lg">
      <p v-if="pending" class="muted">{{ $t('project.loading') }}</p>
      <p v-else-if="!rows.length" class="muted">{{ $t('project.backlog_empty') }}</p>
      <ul v-else class="backlog-list" role="list">
        <li
          v-for="(row, index) in rows"
          :key="row.id"
          class="backlog-item"
          :class="{ 'backlog-item--dragging': dragIndex === index }"
          :draggable="canWrite"
          @dragstart="onDragStart(index)"
          @dragover.prevent
          @drop="onDrop(index)"
        >
          <span v-if="canWrite" class="backlog-item__handle" aria-hidden="true">⋮⋮</span>
          <span class="backlog-item__subject">{{ row.subject }}</span>
          <AppBadge variant="neutral">{{ row.status }}</AppBadge>
          <span class="backlog-item__sp">{{ row.storyPoints ?? '—' }} SP</span>
        </li>
      </ul>
      <p v-if="reordering" class="muted">{{ $t('project.reordering') }}</p>
    </AppCard>
  </div>
</template>

<script setup lang="ts">
definePageMeta({ layout: 'default' })

const route = useRoute()
const { t } = useI18n()
const { extractFetchError } = useApiError()
const appId = computed(() => String(route.params.appId ?? ''))
const { get, pickAppLabel, pickAppMethodologyProfile } = useApplications()
const { listBacklog, reorderBacklog, pickDemandId } = useProject()
const { can } = usePermissions()

const canWrite = computed(() => can('project', 'E'))
const appLabel = ref('')
const profile = ref<'psa' | 'agile_scrum' | 'agile_kanban'>('agile_scrum')
const terms = useMethodologyTerms(profile)
const pending = ref(true)
const reordering = ref(false)
const errorMsg = ref('')
const dragIndex = ref<number | null>(null)
const rows = ref<{ id: string; subject: string; status: string; storyPoints: number | null }[]>([])

async function load() {
  pending.value = true
  errorMsg.value = ''
  try {
    const app = await get(appId.value)
    appLabel.value = pickAppLabel(app)
    profile.value = pickAppMethodologyProfile(app)
    const items = await listBacklog(appId.value, true)
    rows.value = items.map((item) => ({
      id: pickDemandId(item),
      subject: item.subject ?? item.Subject ?? '',
      status: item.status ?? item.Status ?? '',
      storyPoints: item.storyPoints ?? item.StoryPoints ?? null
    }))
  } finally {
    pending.value = false
  }
}

function onDragStart(index: number) {
  dragIndex.value = index
}

async function onDrop(targetIndex: number) {
  const from = dragIndex.value
  dragIndex.value = null
  if (from == null || from === targetIndex || !canWrite.value) return
  const next = [...rows.value]
  const [moved] = next.splice(from, 1)
  next.splice(targetIndex, 0, moved)
  rows.value = next
  reordering.value = true
  errorMsg.value = ''
  try {
    await reorderBacklog(appId.value, next.map((r) => r.id))
  } catch (e) {
    errorMsg.value = extractFetchError(e, t('project.reorder_error'))
    await load()
  } finally {
    reordering.value = false
  }
}

onMounted(load)
</script>

<style scoped>
.backlog-list {
  list-style: none;
  margin: 0;
  padding: 0;
  display: flex;
  flex-direction: column;
  gap: var(--kore-space-sm);
}

.backlog-item {
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  gap: var(--kore-space-sm);
  padding: var(--kore-space-sm) var(--kore-space-md);
  border: 1px solid var(--kore-border);
  border-radius: var(--kore-radius-md);
  background: var(--kore-bg-subtle);
  cursor: grab;
}

.backlog-item--dragging {
  opacity: 0.6;
}

.backlog-item__handle {
  color: var(--kore-text-muted);
  font-size: var(--kore-text-small);
}

.backlog-item__subject {
  flex: 1;
  min-width: 0;
}

.backlog-item__sp {
  font-size: var(--kore-text-small);
  color: var(--kore-text-muted);
}

@media (max-width: 768px) {
  .backlog-item {
    flex-direction: column;
    align-items: flex-start;
  }
}
</style>
