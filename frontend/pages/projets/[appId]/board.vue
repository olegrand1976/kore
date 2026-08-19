<template>
  <div>
    <AppPageHeader :title="terms.board" :subtitle="appLabel">
      <template #actions>
        <AppButton variant="ghost" size="sm" @click="navigateTo('/projets')">{{ $t('project.back') }}</AppButton>
        <AppButton v-if="canConfig" variant="ghost" size="sm" @click="navigateTo(`/projets/${appId}/kanban-config`)">
          {{ $t('project.nav_kanban_config') }}
        </AppButton>
      </template>
    </AppPageHeader>

    <p v-if="errorMsg" class="flash flash--error" role="alert">{{ errorMsg }}</p>
    <p v-if="pending" class="muted">{{ $t('project.loading') }}</p>

    <div v-else class="kanban-board">
      <section v-for="col in displayColumns" :key="col.stateCode" class="kanban-col">
        <header class="kanban-col__head">
          <h2 class="kanban-col__title">{{ col.label || col.stateCode }}</h2>
          <span class="kanban-col__count">
            {{ cardsByStatus[col.stateCode]?.length ?? 0 }}<template v-if="col.wipLimit != null && col.wipLimit > 0">/{{ col.wipLimit }}</template>
          </span>
        </header>
        <ul class="kanban-col__list" role="list">
          <li v-for="card in cardsByStatus[col.stateCode] ?? []" :key="card.id" class="kanban-card">
            <NuxtLink :to="`/tma/${card.id}`" class="kanban-card__link">{{ card.subject }}</NuxtLink>
            <div v-if="canWrite" class="kanban-card__actions">
              <AppButton
                v-if="canTakeOver(card.status)"
                variant="ghost"
                size="sm"
                @click="doTakeOver(card.id)"
              >
                {{ $t('tma.action_take_over') }}
              </AppButton>
              <AppButton
                v-if="canResolve(card.status)"
                variant="ghost"
                size="sm"
                @click="doResolve(card.id)"
              >
                {{ $t('tma.action_resolve') }}
              </AppButton>
            </div>
          </li>
        </ul>
      </section>
    </div>
  </div>
</template>

<script setup lang="ts">
import type { KanbanColumnConfig } from '~/composables/useProject'

definePageMeta({ layout: 'default' })

const route = useRoute()
const { t } = useI18n()
const { extractFetchError } = useApiError()
const appId = computed(() => String(route.params.appId ?? ''))
const { get, pickAppLabel, pickAppMethodologyProfile } = useApplications()
const { getKanbanConfig, pickKanbanColumns } = useProject()
const { list, takeOver, resolve, pickId, pickSubject } = useTma()
const { can } = usePermissions()

const canWrite = computed(() => can('tma', 'E'))
const canConfig = computed(() => can('project', 'E'))
const appLabel = ref('')
const profile = ref<'psa' | 'agile_scrum' | 'agile_kanban'>('agile_kanban')
const terms = useMethodologyTerms(profile)
const pending = ref(true)
const errorMsg = ref('')
const columns = ref<KanbanColumnConfig[]>([])
const cards = ref<{ id: string; subject: string; status: string }[]>([])

const displayColumns = computed((): KanbanColumnConfig[] => {
  const configured = columns.value
  const codes = new Set(configured.map((c) => c.stateCode))
  const orphanCodes = [...new Set(cards.value.map((c) => c.status).filter((s) => s && !codes.has(s)))]
  return [
    ...configured,
    ...orphanCodes.map((stateCode) => ({ stateCode, label: stateCode }))
  ]
})

const cardsByStatus = computed(() => {
  const map: Record<string, typeof cards.value> = {}
  for (const col of displayColumns.value) {
    map[col.stateCode] = []
  }
  for (const card of cards.value) {
    if (!map[card.status]) {
      map[card.status] = []
    }
    map[card.status].push(card)
  }
  return map
})

function canTakeOver(status: string) {
  return status === 'ouverte' || status === 'affectee' || status === 'rework'
}

function canResolve(status: string) {
  return status !== 'resolue' && status !== 'en_attente_creation'
}

async function load() {
  pending.value = true
  errorMsg.value = ''
  try {
    const app = await get(appId.value)
    appLabel.value = pickAppLabel(app)
    profile.value = pickAppMethodologyProfile(app)
    const cfg = await getKanbanConfig(appId.value)
    columns.value = pickKanbanColumns(cfg)
    if (!columns.value.length) {
      columns.value = [
        { stateCode: 'ouverte', label: t('project.kanban_col_open') },
        { stateCode: 'affectee', label: t('project.kanban_col_assigned') },
        { stateCode: 'en_cours', label: t('project.kanban_col_progress'), wipLimit: 3 },
        { stateCode: 'rework', label: t('project.kanban_col_rework') },
        { stateCode: 'resolue', label: t('project.kanban_col_done') }
      ]
    }
    const demands = (await list()).filter(
      (d) => String(d.applicationId ?? d.ApplicationID ?? '') === appId.value
    )
    cards.value = demands.map((d) => ({
      id: pickId(d),
      subject: pickSubject(d),
      status: d.status ?? d.Status ?? 'ouverte'
    }))
  } finally {
    pending.value = false
  }
}

async function doTakeOver(id: string) {
  errorMsg.value = ''
  try {
    await takeOver(id)
    await load()
  } catch (e) {
    errorMsg.value = extractFetchError(e, t('project.kanban_action_error'))
  }
}

async function doResolve(id: string) {
  errorMsg.value = ''
  try {
    await resolve(id)
    await load()
  } catch (e) {
    errorMsg.value = extractFetchError(e, t('project.kanban_action_error'))
  }
}

onMounted(load)
</script>

<style scoped>
.kanban-board {
  display: flex;
  gap: var(--kore-space-md);
  overflow-x: auto;
  padding-bottom: var(--kore-space-md);
}

.kanban-col {
  flex: 0 0 min(280px, 85vw);
  background: var(--kore-bg-subtle);
  border: 1px solid var(--kore-border);
  border-radius: var(--kore-radius-md);
  display: flex;
  flex-direction: column;
  max-height: 70vh;
}

.kanban-col__head {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: var(--kore-space-sm) var(--kore-space-md);
  border-bottom: 1px solid var(--kore-border);
}

.kanban-col__title {
  margin: 0;
  font-size: var(--kore-text-small);
  font-weight: 600;
}

.kanban-col__count {
  font-size: var(--kore-text-small);
  color: var(--kore-text-muted);
}

.kanban-col__list {
  list-style: none;
  margin: 0;
  padding: var(--kore-space-sm);
  overflow-y: auto;
  flex: 1;
  display: flex;
  flex-direction: column;
  gap: var(--kore-space-sm);
}

.kanban-card {
  padding: var(--kore-space-sm);
  background: var(--kore-bg);
  border: 1px solid var(--kore-border);
  border-radius: var(--kore-radius-sm);
}

.kanban-card__link {
  color: var(--kore-link);
  text-decoration: none;
  font-size: var(--kore-text-small);
  display: block;
  margin-bottom: var(--kore-space-xs);
}

.kanban-card__actions {
  display: flex;
  flex-wrap: wrap;
  gap: var(--kore-space-xs);
}
</style>
