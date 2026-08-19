<template>
  <div>
    <AppPageHeader :title="terms.backlog" :subtitle="appLabel">
      <template #actions>
        <AppButton variant="ghost" size="sm" @click="navigateTo('/projets')">{{ $t('project.back') }}</AppButton>
      </template>
    </AppPageHeader>

    <AppCard padding="lg">
      <p v-if="pending" class="muted">{{ $t('project.loading') }}</p>
      <AppTable
        v-else
        :columns="columns"
        :rows="rows"
        :empty-title="$t('project.backlog_empty')"
        row-key="id"
      >
        <template #cell-storyPoints="{ value }">
          {{ value ?? '—' }}
        </template>
      </AppTable>
    </AppCard>
  </div>
</template>

<script setup lang="ts">
definePageMeta({ layout: 'default' })

const route = useRoute()
const { t } = useI18n()
const appId = computed(() => String(route.params.appId ?? ''))
const { get, pickAppLabel, pickAppMethodologyProfile } = useApplications()
const { listBacklog, pickDemandId } = useProject()

const appLabel = ref('')
const profile = ref<'psa' | 'agile_scrum' | 'agile_kanban'>('agile_scrum')
const terms = useMethodologyTerms(profile)
const pending = ref(true)
const rows = ref<{ id: string; subject: string; status: string; storyPoints: number | null }[]>([])

const columns = computed(() => [
  { key: 'subject', label: terms.value.workItem },
  { key: 'status', label: t('project.col_status') },
  { key: 'storyPoints', label: t('project.col_story_points') }
])

onMounted(async () => {
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
})
</script>
