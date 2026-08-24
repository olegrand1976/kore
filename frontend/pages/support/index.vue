<template>
  <div>
    <AppPageHeader :title="$t('support.title')" :subtitle="$t('support.subtitle')">
      <template #actions>
        <AppButton v-if="guideRef?.dismissed" variant="ghost" size="sm" type="button" @click="guideRef?.showAgain()">
          {{ $t('guides.show') }}
        </AppButton>
        <AppButton variant="primary" size="sm" @click="showForm = !showForm">
          {{ $t('support.new') }}
        </AppButton>
      </template>
    </AppPageHeader>

    <AppSectionGuide ref="guideRef" guide-key="support" />

    <AppCard v-if="showForm" padding="lg" class="mb">
      <ServiceRequestForm :busy="busy" @submit="onCreate" />
    </AppCard>

    <AppCard v-if="pending" padding="lg"><p class="muted">{{ $t('support.loading') }}</p></AppCard>
    <AppCard v-else-if="!rows.length" padding="lg">
      <AppEmptyState icon="inbox" :title="$t('support.empty')" />
    </AppCard>
    <AppCard v-else padding="none">
      <AppTable :columns="columns" :rows="rows" row-key="id">
        <template #cell-priority="{ value }">
          <AppBadge variant="neutral">{{ priorityLabel(String(value)) }}</AppBadge>
        </template>
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
import type { ServiceRequestPayload } from '~/components/requests/ServiceRequestForm.vue'
import { REQUEST_RESOURCE, useRequestAttachments } from '~/composables/useRequestAttachments'
import { formatOrgClock } from '~/composables/useCountryTimezone'
import { useUsers } from '~/composables/useUsers'
import { useSupport } from '~/composables/useSupport'
import { ref } from 'vue'

definePageMeta({ layout: 'default' })

const route = useRoute()
const guideRef = ref<{ showAgain: () => void; dismissed: boolean } | null>(null)
const { t, locale } = useI18n()
const { extractFetchError } = useApiError()
const { list, create, pickId, pickSubject, pickState, pickPriority, pickDueAt, pickApplicationId,pickAssigneeId,
  pickTakenOverById } = useSupport()
const { uploadAll } = useRequestAttachments()
const { list: listApps, pickAppLabel, appById } = useApplications()
const { orgPays } = useOrgPays()
const { list: listUsers, pickUserId, pickUserLogin } = useUsers()

const pending = ref(true)
const busy = ref(false)
const showForm = ref(route.query.create === '1')
const errorMsg = ref('')
const tickets = ref<Awaited<ReturnType<typeof list>>>([])
const apps = ref<Awaited<ReturnType<typeof listApps>>>([])

const columns = computed(() => [
  { key: 'subject', label: t('support.col_subject') },
  { key: 'application', label: t('requests.col_application') },
  { key: 'assignee', label: 'Affecté à' },
  { key: 'takenOverBy', label: 'Pris en charge par' },
  { key: 'priority', label: t('requests.col_priority') },
  { key: 'dueAt', label: t('requests.col_due_at') },
  { key: 'state', label: t('support.col_state') },
  { key: 'actions', label: '' }
])

const appMap = computed(() => appById(apps.value))

const rows = computed(() =>
  tickets.value.map((ticket) => {
    const appId = pickApplicationId(ticket)
    return {
      id: pickId(ticket),
      subject: pickSubject(ticket),
      application: pickAppLabel(appMap.value.get(appId)),
      assignee: userMap.value.get(pickAssigneeId(ticket)) || '—',
      takenOverBy: userMap.value.get(pickTakenOverById(ticket)) || '—',
      priority: pickPriority(ticket),
      dueAt: formatDueAt(pickDueAt(ticket)),
      state: pickState(ticket)
    }
  })
)

const userMap = computed(() => {
  const map = new Map<string, string>()

  for (const user of users.value) {
    map.set(pickUserId(user), pickUserLogin(user))
  }

  return map
})

const formatDueAt = (raw: string) =>
  formatOrgClock(raw || null, orgPays.value, locale.value, 'datetime')

const priorityLabel = (priority: string) => t(`requests.priority_${priority}` as const, priority)

const supportStateLabel = (state: string) => {
  const key = `support.state_${state}` as const
  return t(key, state)
}
const users = ref<Awaited<ReturnType<typeof listUsers>>>([])
  
const load = async () => {
  pending.value = true
  errorMsg.value = ''
  try {
    const [ticketList, appList,userList] = await Promise.all([list(), listApps(), listUsers()])
    tickets.value = ticketList
    apps.value = appList
    users.value = userList
  } catch (e) {
    errorMsg.value = extractFetchError(e)
  } finally {
    pending.value = false
  }
}

const onCreate = async (payload: ServiceRequestPayload) => {
  busy.value = true
  errorMsg.value = ''
  try {
    const created = await create(payload)
    const id = pickId(created)
    if (id && payload.files.length) {
      await uploadAll(REQUEST_RESOURCE.support, id, payload.files)
    }
    showForm.value = false
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
.muted { color: var(--kore-text-muted); }
.flash { margin-top: var(--kore-space-md); font-size: var(--kore-text-small); }
.flash--error { color: var(--kore-status-danger); }
</style>
