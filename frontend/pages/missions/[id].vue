<template>
  <div>
    <AppPageHeader :title="pageTitle" :subtitle="$t('fiche.mission_title')">
      <template #actions>
        <AppButton
          v-if="canEditBilling && mission && !editingBilling"
          variant="secondary"
          size="sm"
          type="button"
          class="mission-edit-btn"
          @click="startBillingEdit"
        >
          <AppIcon name="edit" /> {{ $t('missions.edit_billing') }}
        </AppButton>
        <AppButton variant="ghost" size="sm" to="/missions">
          <AppIcon name="arrow_back" /> {{ $t('missions.back_list') }}
        </AppButton>
      </template>
    </AppPageHeader>

    <AppCard v-if="pending" padding="lg">
      <p class="muted">{{ $t('fiche.loading') }}</p>
    </AppCard>

    <AppCard v-else-if="error" padding="lg">
      <AppEmptyState icon="error" :title="$t('fiche.not_found')" />
    </AppCard>

    <template v-else-if="mission">
      <AppKpiGrid compact>
        <AppKpiCard
          icon="flag"
          tone="gold"
          :value="missionStatusLabel(mission.status)"
          :label="$t('fiche.col_status')"
        />
        <AppKpiCard
          icon="payments"
          tone="blue"
          :value="rateLabel"
          :label="$t('missions.col_rate')"
        />
        <AppKpiCard
          icon="groups"
          tone="success"
          :value="collaborators.length"
          :label="$t('fiche.section_staffing')"
        />
        <AppKpiCard
          icon="apps"
          tone="default"
          :value="applications.length"
          :label="$t('fiche.section_applications')"
        />
        <AppKpiCard
          icon="event"
          tone="warn"
          :value="periodLabel"
          :label="$t('fiche.col_period')"
        />
      </AppKpiGrid>

      <div class="fiche-grid">
        <AppCard padding="lg">
          <h3 class="fiche-section-title">{{ $t('fiche.section_overview') }}</h3>

          <form v-if="editingBilling" class="mission-billing-form" @submit.prevent="saveBilling">
            <AppInput
              id="mission-edit-title"
              v-model="billingForm.title"
              :label="$t('missions.field_title')"
            />
            <div class="mission-billing-form__field">
              <label for="mission-edit-rate-unit">{{ $t('missions.field_rate_unit') }}</label>
              <select id="mission-edit-rate-unit" v-model="billingForm.rateUnit" class="mission-billing-form__select">
                <option value="tjm">{{ $t('missions.rate_unit_tjm') }}</option>
                <option value="hourly">{{ $t('missions.rate_unit_hourly') }}</option>
              </select>
            </div>
            <AppInput
              id="mission-edit-amount"
              v-model.number="billingForm.amountEuros"
              type="number"
              min="0"
              :step="billingForm.rateUnit === 'hourly' ? '0.01' : '1'"
              :label="editRateAmountLabel"
              required
            />
            <AppInput
              id="mission-edit-planned-hours"
              v-model="billingForm.plannedWeekHours"
              type="number"
              min="0"
              step="0.5"
              :label="$t('missions.field_planned_week_hours')"
              :tooltip="$t('missions.field_planned_week_hours_hint')"
            />
            <p class="muted-small">{{ $t('missions.field_planned_week_hours_hint') }}</p>
            <ClientContactMultiSelect
              id="mission-edit-contacts"
              v-model="billingForm.clientContactIds"
              :contacts="availableClientContacts"
              :label="$t('missions.field_contact')"
            />
            <AppButton
              variant="ghost"
              size="sm"
              type="button"
              :disabled="!mission.clientId"
              @click="openContactModal"
            >
              <AppIcon name="add" /> {{ $t('missions.contact_add_inline') }}
            </AppButton>
            <p v-if="billingError" class="flash flash--error" role="alert">{{ billingError }}</p>
            <div class="mission-billing-form__actions">
              <AppButton variant="ghost" type="button" :disabled="billingSaving" @click="cancelBillingEdit">
                {{ $t('common.cancel') }}
              </AppButton>
              <AppButton variant="primary" type="submit" :loading="billingSaving">
                {{ $t('missions.save_billing') }}
              </AppButton>
            </div>
          </form>

          <dl v-else class="fiche-dl">
            <div>
              <dt>{{ $t('fiche.col_status') }}</dt>
              <dd>
                <AppBadge :variant="missionStatusVariant(mission.status)">
                  {{ missionStatusLabel(mission.status) }}
                </AppBadge>
              </dd>
            </div>
            <div>
              <dt>{{ $t('missions.field_title') }}</dt>
              <dd>{{ mission.title?.trim() || $t('fiche.none') }}</dd>
            </div>
            <div>
              <dt>{{ $t('fiche.col_client') }}</dt>
              <dd>
                <NuxtLink
                  v-if="mission.clientId && mission.clientName"
                  :to="`/clients/${mission.clientId}`"
                  class="fiche-link"
                >
                  {{ mission.clientName }}
                </NuxtLink>
                <span v-else>{{ mission.clientName || $t('fiche.none') }}</span>
              </dd>
            </div>
            <div>
              <dt>{{ $t('fiche.col_start') }}</dt>
              <dd>{{ formatDate(mission.startDate) }}</dd>
            </div>
            <div>
              <dt>{{ $t('fiche.col_end') }}</dt>
              <dd>{{ mission.endDate ? formatDate(mission.endDate) : $t('fiche.none') }}</dd>
            </div>
            <div>
              <dt>{{ $t('missions.field_rate_unit') }}</dt>
              <dd>{{ rateUnitLabel }}</dd>
            </div>
            <div>
              <dt>{{ $t('missions.col_rate') }}</dt>
              <dd>{{ rateLabel }}</dd>
            </div>
            <div>
              <dt>{{ $t('missions.field_planned_week_hours') }}</dt>
              <dd>{{ plannedWeekHoursLabel }}</dd>
            </div>
            <div v-if="missionContacts.length">
              <dt>{{ $t('fiche.col_client_contact') }}</dt>
              <dd>
                <ul class="mission-contacts">
                  <li v-for="c in missionContacts" :key="c.id">
                    <span class="fiche-strong">{{ contactLabel(c) }}</span>
                    <span v-if="c.role" class="muted-small"> — {{ c.role }}</span>
                    <span v-if="c.email" class="muted-small"> · {{ c.email }}</span>
                  </li>
                </ul>
              </dd>
            </div>
            <div v-else-if="mission.clientContact">
              <dt>{{ $t('fiche.col_client_contact') }}</dt>
              <dd>{{ mission.clientContact }}</dd>
            </div>
            <div>
              <dt>{{ $t('fiche.col_created') }}</dt>
              <dd>{{ formatDate(mission.createdAt) }}</dd>
            </div>
          </dl>
        </AppCard>

        <AppCard padding="lg">
          <h3 class="fiche-section-title">{{ $t('fiche.col_technologies') }}</h3>
          <div v-if="mission.technologies.length" class="fiche-tags">
            <AppBadge v-for="tech in mission.technologies" :key="tech" variant="default">
              {{ tech }}
            </AppBadge>
          </div>
          <p v-else class="muted">{{ $t('fiche.none') }}</p>
        </AppCard>
      </div>

      <AppCard padding="none" class="fiche-table-wrap">
        <div class="fiche-table-head">
          <h3 class="fiche-section-title">{{ $t('fiche.section_applications') }}</h3>
        </div>
        <div v-if="applications.length" class="fiche-tags mission-apps">
          <AppBadge
            v-for="app in applications"
            :key="app.applicationId"
            :variant="app.active === false ? 'warning' : 'default'"
          >
            {{ app.libelle }}
          </AppBadge>
        </div>
        <p v-else class="muted mission-apps-empty">{{ $t('missions.apps_none') }}</p>
      </AppCard>

      <AppCard v-if="canEditApplications" padding="lg" class="fiche-apps-edit">
        <h3 class="fiche-section-title">{{ $t('missions.edit_applications') }}</h3>
        <MissionApplicationMultiSelect
          id="mission-apps-edit"
          :model-value="selectedApplicationIds"
          :applications="applicationOptions"
          :label="$t('missions.field_applications')"
          @update:model-value="onApplicationsModelUpdate"
        />
        <AppButton
          v-if="canCreateApp"
          variant="ghost"
          size="sm"
          type="button"
          @click="appModalOpen = true"
        >
          <AppIcon name="add" /> {{ $t('missions.app_add_inline') }}
        </AppButton>
        <p v-if="appsError" class="flash flash--error" role="alert">{{ appsError }}</p>
        <AppButton variant="primary" size="sm" :loading="appsSaving" @click="saveApplications">
          {{ $t('missions.save_applications') }}
        </AppButton>
      </AppCard>

      <AppCard padding="lg" class="mission-billing-history">
        <h3 class="fiche-section-title">{{ $t('missions.billing_history_title') }}</h3>
        <form
          v-if="canEditBilling"
          class="mission-billing-note"
          @submit.prevent="submitBillingNote"
        >
          <AppInput
            id="mission-billing-note"
            v-model="billingNote"
            :label="$t('missions.billing_note_label')"
            :placeholder="$t('missions.billing_note_placeholder')"
          />
          <AppButton variant="secondary" size="sm" type="submit" :loading="billingNoteSaving">
            {{ $t('missions.billing_note_add') }}
          </AppButton>
          <p v-if="billingNoteError" class="flash flash--error" role="alert">{{ billingNoteError }}</p>
        </form>
        <p v-if="billingEventsPending" class="muted">{{ $t('fiche.loading') }}</p>
        <ol v-else-if="billingEvents.length" class="mission-billing-timeline">
          <li v-for="ev in billingEvents" :key="ev.id" class="mission-billing-timeline__item">
            <div class="mission-billing-timeline__meta">
              <AppBadge variant="default">{{ billingEventTypeLabel(ev.eventType) }}</AppBadge>
              <time class="muted-small" :datetime="ev.createdAt">{{ formatDate(ev.createdAt) }}</time>
            </div>
            <p v-if="ev.message" class="mission-billing-timeline__message">{{ ev.message }}</p>
            <p v-else-if="billingEventSummary(ev)" class="mission-billing-timeline__message muted">
              {{ billingEventSummary(ev) }}
            </p>
          </li>
        </ol>
        <p v-else class="muted">{{ $t('missions.billing_history_empty') }}</p>
      </AppCard>

      <AppCard padding="none" class="fiche-table-wrap">
        <div class="fiche-table-head">
          <h3 class="fiche-section-title">{{ $t('fiche.section_staffing') }}</h3>
        </div>
        <AppTable
          :columns="staffColumns"
          :rows="staffRows"
          row-key="id"
          :empty-title="$t('fiche.staffing_empty')"
        >
          <template #cell-name="{ row }">
            <NuxtLink :to="`/collaborateurs/${row.id}`" class="fiche-link fiche-strong">
              {{ row.name }}
            </NuxtLink>
          </template>
          <template #cell-login="{ value }">
            <span class="muted-small">{{ value }}</span>
          </template>
        </AppTable>
      </AppCard>

      <AppCard v-if="canEditStaffing" padding="lg" class="fiche-staff-edit">
        <h3 class="fiche-section-title">{{ $t('missions.edit_collaborators') }}</h3>
        <AppUserMultiSelect
          id="mission-staff-edit"
          v-model="selectedCollaboratorIds"
          :label="$t('missions.field_collaborators')"
          required
        />
        <p v-if="staffError" class="flash flash--error" role="alert">{{ staffError }}</p>
        <AppButton variant="primary" size="sm" :loading="staffSaving" @click="saveCollaborators">
          {{ $t('missions.save_collaborators') }}
        </AppButton>
      </AppCard>
    </template>

    <AppModal v-model:open="contactModalOpen" width="sm" :aria-label="$t('missions.contact_add_inline')">
      <form class="mission-contact-form" @submit.prevent="submitContact">
        <h2 class="mission-contact-form__title">{{ $t('missions.contact_add_inline') }}</h2>
        <AppInput id="mission-detail-contact-prenom" v-model="contactForm.prenom" :label="$t('clients.contact_prenom')" />
        <AppInput id="mission-detail-contact-nom" v-model="contactForm.nom" :label="$t('clients.contact_nom')" />
        <AppInput id="mission-detail-contact-role" v-model="contactForm.role" :label="$t('clients.contact_role')" />
        <AppInput id="mission-detail-contact-email" v-model="contactForm.email" type="email" :label="$t('clients.contact_email')" />
        <AppInput id="mission-detail-contact-phone" v-model="contactForm.telephone" :label="$t('clients.contact_phone')" />
        <p v-if="contactError" class="flash flash--error" role="alert">{{ contactError }}</p>
        <div class="mission-contact-form__actions">
          <AppButton variant="ghost" type="button" @click="contactModalOpen = false">
            {{ $t('common.cancel') }}
          </AppButton>
          <AppButton variant="primary" type="submit" :loading="creatingContact">
            {{ $t('common.save') }}
          </AppButton>
        </div>
      </form>
    </AppModal>

    <MissionApplicationCreateModal
      v-model:open="appModalOpen"
      :default-libelle="mission?.title"
      :sites="sites"
      @created="onAppCreated"
    />
  </div>
</template>

<script setup lang="ts">
import { formatUserDisplayName } from '~/composables/useUserDisplay'
import { hoursInputToMinutes, minutesToHoursInput } from '~/utils/craWeekCapacity'

definePageMeta({ layout: 'default' })

const { apiFetch } = useApiFetch()
type MissionCollaborator = {
  userId?: string
  login?: string
  prenom?: string
  nom?: string
}

type MissionApplication = {
  applicationId: string
  libelle: string
  active?: boolean
}

type MissionClientContact = {
  id: string
  nom?: string
  prenom?: string
  email?: string
  role?: string
  telephone?: string
}

type MissionDetail = {
  id?: string
  clientId?: string
  clientName?: string
  status: string
  startDate?: string
  endDate?: string | null
  title?: string
  rateUnit?: string
  tjmAmount?: number
  plannedWeekMinutes?: number | null
  currency?: string
  technologies: string[]
  clientContact?: string
  clientContactIds?: string[]
  clientContacts?: MissionClientContact[]
  createdAt?: string
  collaborators?: MissionCollaborator[]
  applications?: MissionApplication[]
}

type MissionBillingEvent = {
  id: string
  missionId?: string
  actorUserId?: string
  eventType: string
  message?: string
  payload?: Record<string, unknown>
  createdAt: string
}

const route = useRoute()
const { t } = useI18n()
const { isAdmin } = useAuth()
const { can } = usePermissions()
const { extractFetchError } = useApiError()
const { replaceClientContacts, listApplications, listSites, orgId, orgLabel } = useOrganisation()
const { formatDate, formatMoney, missionStatusLabel, missionStatusVariant } = useFicheFormat()

const id = computed(() => String(route.params.id ?? ''))
const canEditStaffing = computed(() => isAdmin.value)
const canEditBilling = computed(() => can('ssii', 'E'))
const canEditApplications = computed(() => can('ssii', 'E'))
const canCreateApp = computed(() => can('org', 'E'))
const selectedCollaboratorIds = ref<string[]>([])
const staffSaving = ref(false)
const staffError = ref('')
const selectedApplicationIds = ref<string[]>([])
const appsDirty = ref(false)
const applicationOptions = ref<{ id: string; libelle: string; active?: boolean }[]>([])
const sites = ref<{ id: string; label: string }[]>([])
const appsSaving = ref(false)
const appsError = ref('')
const appModalOpen = ref(false)

type RateUnit = 'tjm' | 'hourly'
const editingBilling = ref(false)
const billingSaving = ref(false)
const billingError = ref('')
const billingForm = reactive({
  title: '',
  rateUnit: 'tjm' as RateUnit,
  amountEuros: 0,
  plannedWeekHours: '',
  clientContactIds: [] as string[]
})

const billingEvents = ref<MissionBillingEvent[]>([])
const billingEventsPending = ref(false)
const billingNote = ref('')
const billingNoteSaving = ref(false)
const billingNoteError = ref('')

const availableClientContacts = ref<MissionClientContact[]>([])
const contactModalOpen = ref(false)
const creatingContact = ref(false)
const contactError = ref('')
const contactForm = reactive({
  prenom: '',
  nom: '',
  role: '',
  email: '',
  telephone: ''
})

const contactLabel = (c: MissionClientContact) => {
  const name = [c.prenom, c.nom].filter(Boolean).join(' ').trim()
  if (name) return name
  if (c.email?.trim()) return c.email.trim()
  return c.id
}

const editRateAmountLabel = computed(() => {
  switch (billingForm.rateUnit) {
    case 'hourly':
      return t('missions.field_hourly_amount')
    case 'tjm':
      return t('missions.field_tjm_amount')
    default: {
      const _exhaustive: never = billingForm.rateUnit
      return _exhaustive
    }
  }
})

const normalizeRateUnit = (value: unknown): RateUnit => {
  switch (String(value ?? '').toLowerCase()) {
    case 'hourly':
      return 'hourly'
    case 'tjm':
    default:
      return 'tjm'
  }
}

const { data, pending, error, refresh } = await useFetch<MissionDetail>(() => `/api/ssii/missions/${id.value}`, {
  watch: [id]
})

const mission = computed(() => {
  const payload = (data.value as { data?: MissionDetail })?.data ?? data.value
  if (!payload || typeof payload !== 'object') return null
  return {
    ...payload,
    status: payload.status ?? 'active',
    technologies: payload.technologies ?? [],
    collaborators: payload.collaborators ?? [],
    applications: payload.applications ?? []
  }
})

const collaborators = computed(() => mission.value?.collaborators ?? [])
const applications = computed(() => mission.value?.applications ?? [])
const missionContacts = computed(() => mission.value?.clientContacts ?? [])

const loadClientContacts = async () => {
  const clientId = mission.value?.clientId
  if (!clientId) {
    availableClientContacts.value = []
    return
  }
  try {
    const res = await apiFetch<{ data?: { contacts?: MissionClientContact[] } }>(`/api/org/clients/${clientId}`)
    const contacts = res?.data?.contacts ?? []
    availableClientContacts.value = contacts.filter((c): c is MissionClientContact => Boolean(c.id))
  } catch {
    availableClientContacts.value = missionContacts.value
  }
}

const loadApplicationCatalog = async () => {
  try {
    const items = await listApplications()
    const byId = new Map<string, { id: string; libelle: string; active?: boolean }>()
    for (const a of items) {
      const appId = orgId(a)
      if (!appId) continue
      const active = a.active ?? a.Active ?? true
      if (active === false) continue
      byId.set(appId, { id: appId, libelle: orgLabel(a) || appId, active: true })
    }
    // Keep currently linked inactive apps visible (uncheckable to remove only).
    for (const linked of applications.value) {
      const appId = String(linked.applicationId ?? '')
      if (!appId) continue
      if (!byId.has(appId)) {
        byId.set(appId, {
          id: appId,
          libelle: linked.libelle || appId,
          active: linked.active !== false
        })
      }
    }
    applicationOptions.value = [...byId.values()].sort((a, b) =>
      a.libelle.localeCompare(b.libelle, 'fr')
    )
  } catch {
    applicationOptions.value = []
  }
}

const loadSitesCatalog = async () => {
  try {
    const items = await listSites()
    sites.value = items
      .map((s) => ({ id: orgId(s), label: orgLabel(s) || orgId(s) }))
      .filter((s) => s.id)
      .sort((a, b) => a.label.localeCompare(b.label, 'fr'))
  } catch {
    sites.value = []
  }
}

watch(
  () => mission.value?.clientId,
  () => {
    void loadClientContacts()
  },
  { immediate: true }
)

watch(
  collaborators,
  (items) => {
    selectedCollaboratorIds.value = items
      .map((c) => String(c.userId ?? ''))
      .filter(Boolean)
  },
  { immediate: true }
)

watch(
  applications,
  (items) => {
    if (appsDirty.value) return
    selectedApplicationIds.value = items
      .map((a) => String(a.applicationId ?? ''))
      .filter(Boolean)
  },
  { immediate: true }
)

const onApplicationsModelUpdate = (value: string[]) => {
  appsDirty.value = true
  selectedApplicationIds.value = value
}

onMounted(() => {
  if (canEditApplications.value) {
    void Promise.all([loadApplicationCatalog(), loadSitesCatalog()])
  }
})

const saveCollaborators = async () => {
  staffError.value = ''
  if (!selectedCollaboratorIds.value.length) {
    staffError.value = t('missions.collaborators_required')
    return
  }
  staffSaving.value = true
  try {
    await apiFetch(`/api/ssii/missions/${id.value}/collaborators`, {
      method: 'PUT',
      body: { collaboratorIds: selectedCollaboratorIds.value }
    })
    await refresh()
  } catch {
    staffError.value = t('missions.save_collaborators_error')
  } finally {
    staffSaving.value = false
  }
}

const saveApplications = async () => {
  appsError.value = ''
  appsSaving.value = true
  try {
    await apiFetch(`/api/ssii/missions/${id.value}/applications`, {
      method: 'PUT',
      body: { applicationIds: selectedApplicationIds.value }
    })
    appsDirty.value = false
    await refresh()
    await loadApplicationCatalog()
  } catch {
    appsError.value = t('missions.save_applications_error')
  } finally {
    appsSaving.value = false
  }
}

const onAppCreated = async (applicationId: string) => {
  await loadApplicationCatalog()
  appsDirty.value = true
  if (applicationId && !selectedApplicationIds.value.includes(applicationId)) {
    selectedApplicationIds.value = [...selectedApplicationIds.value, applicationId]
  }
}

const pageTitle = computed(() => {
  const title = mission.value?.title?.trim()
  if (title) return title
  if (mission.value?.clientName) {
    return `${t('fiche.mission_title')} — ${mission.value.clientName}`
  }
  return t('fiche.mission_title')
})

const rateUnitLabel = computed(() => {
  switch (normalizeRateUnit(mission.value?.rateUnit)) {
    case 'hourly':
      return t('missions.rate_unit_hourly')
    case 'tjm':
      return t('missions.rate_unit_tjm')
    default: {
      const _exhaustive: never = normalizeRateUnit(mission.value?.rateUnit)
      return _exhaustive
    }
  }
})

const rateLabel = computed(() => {
  const amount = formatMoney(Number(mission.value?.tjmAmount ?? 0), mission.value?.currency ?? 'EUR')
  switch (normalizeRateUnit(mission.value?.rateUnit)) {
    case 'hourly':
      return `${amount}/h`
    case 'tjm':
      return `${amount}/j`
    default: {
      const _exhaustive: never = normalizeRateUnit(mission.value?.rateUnit)
      return _exhaustive
    }
  }
})

const plannedWeekHoursLabel = computed(() => {
  const minutes = mission.value?.plannedWeekMinutes
  if (minutes == null || !Number.isFinite(Number(minutes)) || Number(minutes) <= 0) {
    return t('fiche.none')
  }
  return t('missions.planned_week_hours_value', { hours: minutesToHoursInput(Number(minutes)) })
})

const unwrapList = <T,>(payload: unknown): T[] => {
  if (Array.isArray(payload)) return payload as T[]
  if (payload && typeof payload === 'object' && Array.isArray((payload as { data?: unknown }).data)) {
    return (payload as { data: T[] }).data
  }
  return []
}

const loadBillingEvents = async () => {
  if (!id.value) return
  billingEventsPending.value = true
  try {
    const res = await apiFetch<unknown>(`/api/ssii/missions/${id.value}/billing-events`)
    billingEvents.value = unwrapList<MissionBillingEvent>(res).map((ev) => ({
      id: String(ev.id),
      missionId: ev.missionId,
      actorUserId: ev.actorUserId,
      eventType: String(ev.eventType ?? 'note'),
      message: ev.message ?? '',
      payload: ev.payload,
      createdAt: String(ev.createdAt ?? '')
    }))
  } catch {
    billingEvents.value = []
  } finally {
    billingEventsPending.value = false
  }
}

const billingEventTypeLabel = (eventType: string) => {
  switch (eventType) {
    case 'note':
      return t('missions.billing_event_note')
    case 'rate_updated':
      return t('missions.billing_event_rate_updated')
    case 'planned_hours_updated':
      return t('missions.billing_event_planned_hours')
    default:
      return eventType
  }
}

const billingEventSummary = (ev: MissionBillingEvent) => {
  const payload = ev.payload ?? {}
  if (ev.eventType === 'planned_hours_updated') {
    const after = (payload.after as { plannedWeekMinutes?: number | null } | undefined)?.plannedWeekMinutes
    if (after == null) return t('missions.billing_event_planned_cleared')
    return t('missions.billing_event_planned_set', { hours: minutesToHoursInput(after) })
  }
  if (ev.eventType === 'rate_updated') {
    return t('missions.billing_event_rate_summary')
  }
  return ''
}

const submitBillingNote = async () => {
  const message = billingNote.value.trim()
  if (!message) {
    billingNoteError.value = t('missions.billing_note_required')
    return
  }
  billingNoteSaving.value = true
  billingNoteError.value = ''
  try {
    await apiFetch(`/api/ssii/missions/${id.value}/billing-events`, {
      method: 'POST',
      body: { message }
    })
    billingNote.value = ''
    await loadBillingEvents()
  } catch (err) {
    billingNoteError.value = extractFetchError(err, t('missions.billing_note_error'))
  } finally {
    billingNoteSaving.value = false
  }
}

watch(id, () => {
  void loadBillingEvents()
}, { immediate: true })

const startBillingEdit = () => {
  billingForm.title = mission.value?.title ?? ''
  billingForm.rateUnit = normalizeRateUnit(mission.value?.rateUnit)
  billingForm.amountEuros = Number(mission.value?.tjmAmount ?? 0) / 100
  billingForm.plannedWeekHours = minutesToHoursInput(mission.value?.plannedWeekMinutes)
  billingForm.clientContactIds = [...(mission.value?.clientContactIds ?? [])]
  billingError.value = ''
  editingBilling.value = true
  void loadClientContacts()
}

const cancelBillingEdit = () => {
  editingBilling.value = false
  billingError.value = ''
}

const openContactModal = () => {
  contactForm.prenom = ''
  contactForm.nom = ''
  contactForm.role = ''
  contactForm.email = ''
  contactForm.telephone = ''
  contactError.value = ''
  contactModalOpen.value = true
}

const submitContact = async () => {
  const clientId = mission.value?.clientId
  if (!clientId) return
  creatingContact.value = true
  contactError.value = ''
  try {
    const existing = availableClientContacts.value.map((c) => ({
      id: c.id,
      nom: c.nom ?? '',
      prenom: c.prenom ?? '',
      email: c.email ?? '',
      role: c.role ?? '',
      telephone: c.telephone ?? ''
    }))
    const beforeIds = new Set(availableClientContacts.value.map((c) => c.id))
    const res = await replaceClientContacts(clientId, [
      ...existing,
      {
        prenom: contactForm.prenom.trim(),
        nom: contactForm.nom.trim(),
        role: contactForm.role.trim(),
        email: contactForm.email.trim(),
        telephone: contactForm.telephone.trim()
      }
    ])
    const updated = (res as { data?: { contacts?: MissionClientContact[] } })?.data?.contacts ?? []
    availableClientContacts.value = updated.filter((c): c is MissionClientContact => Boolean(c.id))
    const created = availableClientContacts.value.find((c) => !beforeIds.has(c.id))
    if (created?.id && !billingForm.clientContactIds.includes(created.id)) {
      billingForm.clientContactIds = [...billingForm.clientContactIds, created.id]
    }
    contactModalOpen.value = false
  } catch (err) {
    contactError.value = extractFetchError(err, t('clients.contact_save_error'))
  } finally {
    creatingContact.value = false
  }
}

const saveBilling = async () => {
  billingSaving.value = true
  billingError.value = ''
  try {
    await apiFetch(`/api/ssii/missions/${id.value}`, {
      method: 'PUT',
      body: {
        title: billingForm.title.trim(),
        rateUnit: billingForm.rateUnit,
        tjmAmount: Math.round(Number(billingForm.amountEuros) * 100),
        plannedWeekMinutes: hoursInputToMinutes(billingForm.plannedWeekHours),
        clientContactIds: billingForm.clientContactIds
      }
    })
    editingBilling.value = false
    await refresh()
    await loadBillingEvents()
  } catch (err) {
    billingError.value = extractFetchError(err, t('missions.save_billing_error'))
  } finally {
    billingSaving.value = false
  }
}

const periodLabel = computed(() => {
  if (!mission.value?.startDate) return '—'
  const start = formatDate(mission.value.startDate)
  if (!mission.value.endDate) return start
  return `${start} → ${formatDate(mission.value.endDate)}`
})

const staffColumns = computed(() => [
  { key: 'name', label: t('fiche.col_name') },
  { key: 'login', label: t('fiche.col_login') }
])

const staffRows = computed(() =>
  collaborators.value.map((c) => ({
    id: String(c.userId ?? ''),
    name: formatUserDisplayName(c.prenom, c.nom, c.login),
    login: c.login ?? '—'
  }))
)
</script>

<style scoped>
.muted { color: var(--kore-text-muted); }

.muted-small {
  color: var(--kore-text-muted);
  font-size: var(--kore-text-caption);
}

.fiche-grid {
  display: grid;
  gap: var(--kore-space-lg);
  margin-bottom: var(--kore-space-lg);
}

@media (min-width: 768px) {
  .fiche-grid {
    grid-template-columns: 2fr 1fr;
  }
}

.fiche-section-title {
  margin: 0 0 var(--kore-space-md);
  font-size: var(--kore-text-h3);
}

.fiche-dl {
  display: grid;
  gap: var(--kore-space-md);
  margin: 0;
}

.fiche-dl div {
  display: grid;
  gap: var(--kore-space-xs);
}

.fiche-dl dt {
  margin: 0;
  font-size: var(--kore-text-caption);
  color: var(--kore-text-muted);
  font-weight: 500;
}

.fiche-dl dd {
  margin: 0;
  color: var(--kore-text);
  font-weight: 500;
}

.fiche-tags {
  display: flex;
  flex-wrap: wrap;
  gap: var(--kore-space-xs);
}

.fiche-link {
  color: var(--kore-brand-blue);
  text-decoration: none;
  font-weight: 500;
}

.fiche-link:hover { text-decoration: underline; }

.fiche-strong { font-weight: 600; }

.fiche-table-wrap { overflow: hidden; }

.fiche-table-head {
  padding: var(--kore-space-lg) var(--kore-space-lg) 0;
}

.fiche-staff-edit {
  margin-top: var(--kore-space-lg);
  display: grid;
  gap: var(--kore-space-md);
}

.mission-billing-history {
  margin-bottom: var(--kore-space-lg);
}

.mission-billing-note {
  display: grid;
  gap: var(--kore-space-sm);
  margin-bottom: var(--kore-space-lg);
}

.mission-billing-timeline {
  list-style: none;
  margin: 0;
  padding: 0;
  display: grid;
  gap: var(--kore-space-md);
}

.mission-billing-timeline__item {
  padding-bottom: var(--kore-space-md);
  border-bottom: 1px solid var(--kore-border);
}

.mission-billing-timeline__item:last-child {
  border-bottom: none;
  padding-bottom: 0;
}

.mission-billing-timeline__meta {
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  gap: var(--kore-space-sm);
  margin-bottom: var(--kore-space-xs);
}

.mission-billing-timeline__message {
  margin: 0;
  font-size: var(--kore-text-small);
}

.fiche-apps-edit {
  margin-top: var(--kore-space-lg);
  display: grid;
  gap: var(--kore-space-md);
}

.mission-apps {
  padding: 0 var(--kore-space-lg) var(--kore-space-lg);
}

.mission-apps-empty {
  margin: 0;
  padding: 0 var(--kore-space-lg) var(--kore-space-lg);
}

.mission-contact-form__field {
  display: flex;
  flex-direction: column;
  gap: var(--kore-space-xs);
}

.mission-contact-form__field label {
  font-size: var(--kore-text-small);
  font-weight: 500;
  color: var(--kore-text-muted);
}

.mission-billing-form {
  display: grid;
  gap: var(--kore-space-md);
  max-width: var(--kore-form-max);
}

.mission-billing-form__field {
  display: flex;
  flex-direction: column;
  gap: var(--kore-space-xs);
}

.mission-billing-form__field label {
  font-size: var(--kore-text-small);
  font-weight: 500;
  color: var(--kore-text-muted);
}

.mission-billing-form__select {
  padding: 0.625rem 0.75rem;
  border: 1px solid var(--kore-border);
  border-radius: var(--kore-radius-md);
  background: var(--kore-bg);
  color: var(--kore-text);
  font-size: var(--kore-text-small);
}

.mission-billing-form__actions {
  display: flex;
  flex-wrap: wrap;
  gap: var(--kore-space-sm);
  justify-content: flex-end;
}

.mission-contacts {
  margin: 0;
  padding: 0;
  list-style: none;
  display: grid;
  gap: var(--kore-space-xs);
}

.mission-contact-form {
  display: grid;
  gap: var(--kore-space-md);
}

.mission-contact-form__title {
  margin: 0;
  font-size: var(--kore-text-h3);
}

.mission-contact-form__actions {
  display: flex;
  flex-wrap: wrap;
  gap: var(--kore-space-sm);
  justify-content: flex-end;
}

.flash--error {
  margin: 0;
  color: var(--kore-error);
  font-size: var(--kore-text-small);
}

@media (max-width: 768px) {
  .mission-edit-btn {
    width: 100%;
  }

  .mission-billing-form__actions,
  .mission-contact-form__actions {
    flex-direction: column-reverse;
  }

  .mission-billing-form__actions :deep(.app-button),
  .mission-contact-form__actions :deep(.app-button) {
    width: 100%;
  }
}
</style>
