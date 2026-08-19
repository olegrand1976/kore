<template>
  <div>
    <AppPageHeader :title="$t('project.kanban_config_title')" :subtitle="appLabel">
      <template #actions>
        <AppButton variant="ghost" size="sm" @click="navigateTo(`/projets/${appId}/board`)">{{ $t('project.nav_board') }}</AppButton>
      </template>
    </AppPageHeader>

    <p v-if="errorMsg" class="flash flash--error" role="alert">{{ errorMsg }}</p>

    <AppCard v-if="pending" padding="lg">
      <p class="muted">{{ $t('project.loading') }}</p>
    </AppCard>

    <AppCard v-else-if="!canWrite" padding="lg">
      <p class="muted">{{ $t('project.kanban_config_forbidden') }}</p>
    </AppCard>

    <AppCard v-else padding="lg">
      <form class="kanban-form" @submit.prevent="save">
        <div v-for="(col, index) in columns" :key="index" class="kanban-form__row">
          <AppInput v-model="col.label" :label="$t('project.field_column_label')" />
          <AppInput v-model="col.stateCode" :label="$t('project.field_state_code')" required />
          <AppInput v-model.number="col.wipLimit" type="number" min="0" :label="$t('project.field_wip_limit')" />
          <AppButton variant="ghost" size="sm" type="button" @click="removeColumn(index)">{{ $t('common.delete') }}</AppButton>
        </div>
        <div class="kanban-form__actions">
          <AppButton variant="ghost" type="button" @click="addColumn">{{ $t('project.column_add') }}</AppButton>
          <AppButton variant="primary" type="submit" :disabled="busy">{{ $t('common.save') }}</AppButton>
        </div>
      </form>
    </AppCard>
  </div>
</template>

<script setup lang="ts">
import type { KanbanColumnConfig } from '~/composables/useProject'

definePageMeta({ layout: 'default' })

const route = useRoute()
const { t } = useI18n()
const { extractFetchError } = useApiError()
const appId = computed(() => String(route.params.appId ?? ''))
const { get, pickAppLabel } = useApplications()
const { getKanbanConfig, saveKanbanConfig, pickKanbanColumns } = useProject()
const { can } = usePermissions()

const canWrite = computed(() => can('project', 'E'))

const appLabel = ref('')
const pending = ref(true)
const busy = ref(false)
const errorMsg = ref('')
const columns = ref<KanbanColumnConfig[]>([])

async function load() {
  pending.value = true
  try {
    const app = await get(appId.value)
    appLabel.value = pickAppLabel(app)
    const cfg = await getKanbanConfig(appId.value)
    columns.value = pickKanbanColumns(cfg).map((c) => ({ ...c }))
    if (!columns.value.length) {
      columns.value = [
        { stateCode: 'ouverte', label: t('project.kanban_col_open') },
        { stateCode: 'en_cours', label: t('project.kanban_col_progress'), wipLimit: 3 },
        { stateCode: 'resolue', label: t('project.kanban_col_done') }
      ]
    }
  } finally {
    pending.value = false
  }
}

function addColumn() {
  columns.value.push({ stateCode: '', label: '', wipLimit: undefined })
}

function removeColumn(index: number) {
  columns.value.splice(index, 1)
}

async function save() {
  busy.value = true
  errorMsg.value = ''
  try {
    await saveKanbanConfig(appId.value, columns.value)
    navigateTo(`/projets/${appId.value}/board`)
  } catch (e) {
    errorMsg.value = extractFetchError(e, t('project.kanban_config_error'))
  } finally {
    busy.value = false
  }
}

onMounted(load)
</script>

<style scoped>
.kanban-form {
  display: flex;
  flex-direction: column;
  gap: var(--kore-space-md);
}

.kanban-form__row {
  display: grid;
  gap: var(--kore-space-sm);
  grid-template-columns: repeat(auto-fit, minmax(140px, 1fr));
  align-items: end;
  padding-bottom: var(--kore-space-md);
  border-bottom: 1px solid var(--kore-border);
}

.kanban-form__actions {
  display: flex;
  flex-wrap: wrap;
  justify-content: flex-end;
  gap: var(--kore-space-sm);
}

@media (max-width: 768px) {
  .kanban-form__row {
    grid-template-columns: 1fr;
  }
}
</style>
