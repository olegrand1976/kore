<template>
  <div>
    <AppPageHeader :title="terms.board" :subtitle="sprintName">
      <template #actions>
        <AppButton variant="ghost" size="sm" @click="navigateTo(`/projets/${appId}/sprints`)">
          {{ $t('project.back_sprints') }}
        </AppButton>
      </template>
    </AppPageHeader>

    <AppCard padding="lg">
      <p v-if="pending" class="muted">{{ $t('project.loading') }}</p>
      <AppKanbanBoard
        v-else
        :columns="columns"
        :items="items"
        :column-key="columnKey"
        :item-key="itemKey"
        :empty-label="$t('project.board_empty')"
      >
        <template #card="{ item }">
          <div class="board-card">
            <p class="board-card__title">{{ (item as BoardItem).subject }}</p>
            <p v-if="(item as BoardItem).storyPoints != null" class="board-card__points">
              {{ (item as BoardItem).storyPoints }} {{ $t('project.story_points_abbr') }}
            </p>
          </div>
        </template>
      </AppKanbanBoard>
    </AppCard>
  </div>
</template>

<script setup lang="ts">
import type { KanbanColumn } from '~/components/ui/AppKanbanBoard.vue'

type BoardItem = { id: string; subject: string; status: string; storyPoints: number | null }

definePageMeta({ layout: 'default' })

const route = useRoute()
const { t } = useI18n()
const appId = computed(() => String(route.params.appId ?? ''))
const sprintId = computed(() => String(route.params.id ?? ''))
const { apiFetch } = useApiFetch()
const { listSprints, pickSprintName } = useProject()

const terms = useMethodologyTerms('agile_scrum')
const pending = ref(true)
const sprintName = ref('')
const items = ref<BoardItem[]>([])

const columns: KanbanColumn[] = [
  { id: 'ouverte', label: t('tma.status_open'), tone: 'muted' },
  { id: 'affectee', label: t('tma.status_assigned'), tone: 'blue' },
  { id: 'en_cours', label: t('tma.status_in_progress'), tone: 'gold' },
  { id: 'resolue', label: t('tma.status_resolved'), tone: 'success' },
  { id: 'rework', label: t('tma.status_rework'), tone: 'warn' }
]

const columnKey = (item: unknown) => (item as BoardItem).status || 'ouverte'
const itemKey = (item: unknown) => (item as BoardItem).id

onMounted(async () => {
  try {
    const sprints = await listSprints(appId.value)
    const sprint = sprints.find((s) => (s.id ?? s.ID) === sprintId.value)
    sprintName.value = sprint ? pickSprintName(sprint) : sprintId.value
    const res = await apiFetch<{ data?: BoardItem[] }>(
      `/api/tma/demands?applicationId=${appId.value}&sprintId=${sprintId.value}`
    )
    const raw = res?.data ?? []
    items.value = raw.map((d: Record<string, unknown>) => ({
      id: String(d.id ?? d.ID ?? ''),
      subject: String(d.subject ?? d.Subject ?? ''),
      status: String(d.status ?? d.Status ?? 'ouverte'),
      storyPoints: (d.storyPoints ?? d.StoryPoints ?? null) as number | null
    }))
  } finally {
    pending.value = false
  }
})
</script>

<style scoped>
.board-card__title {
  margin: 0;
  font-size: var(--kore-text-small);
}

.board-card__points {
  margin: var(--kore-space-xs) 0 0;
  font-size: var(--kore-text-caption);
  color: var(--kore-text-muted);
}
</style>
