<script setup lang="ts">
type ClientOption = { id: string; label: string }

type PrestationRow = {
  id: string
  userPrenom?: string
  userNom?: string
  userLogin?: string
  totalMinutes?: number
  status: string
}

type InvoicePreview = {
  timesheetId: string
  ok: boolean
  blockers?: string[]
  clientId?: string
  billableHours?: number
  unitPriceCents?: number
  currency?: string
  taxRate?: number
  description?: string
  missionLabel?: string
  userLabel?: string
}

type EditableDraft = {
  timesheetId: string
  userLabel: string
  ok: boolean
  editable: boolean
  hardBlocked: boolean
  blockers: string[]
  clientId: string
  billableHours: string
  unitPriceEur: string
  taxRate: string
  currency: string
  description: string
}

/** Blockers the user can clear by editing client / hours / unit price. */
const SOFT_BLOCKERS = new Set([
  'client_unresolved',
  'zero_unit_price',
  'no_billable_hours'
])

const isHardBlocked = (blockers: string[]) =>
  blockers.some((b) => !SOFT_BLOCKERS.has(b))

const isEditablePreview = (p: InvoicePreview) =>
  p.ok || !isHardBlocked(p.blockers ?? [])


const props = defineProps<{
  clients: ClientOption[]
  open: boolean
}>()

const emit = defineEmits<{
  'update:open': [open: boolean]
  created: [invoiceIds: string[]]
}>()

const { t } = useI18n()
const { apiFetch } = useApiFetch()
const { extractFetchError } = useApiError()
const { mapInvoiceDraftReason } = useCraError()

const step = ref<1 | 2 | 3>(1)
const month = ref(new Date().toISOString().slice(0, 7))
const loadingList = ref(false)
const loadingPreview = ref(false)
const creating = ref(false)
const error = ref('')
const prestations = ref<PrestationRow[]>([])
const selectedIds = ref<string[]>([])
const drafts = ref<EditableDraft[]>([])

const openModel = computed({
  get: () => props.open,
  set: (v: boolean) => emit('update:open', v)
})

const definitiveRows = computed(() =>
  prestations.value.filter((r) => r.status === 'Définitif')
)

const selectableDrafts = computed(() =>
  drafts.value.filter((d) => {
    if (d.hardBlocked) return false
    return !!d.clientId && Number(d.billableHours) > 0 && Number(d.unitPriceEur) > 0
  })
)

const resetWizard = () => {
  step.value = 1
  error.value = ''
  selectedIds.value = []
  drafts.value = []
  prestations.value = []
}

watch(
  () => props.open,
  async (open) => {
    if (!open) {
      resetWizard()
      return
    }
    await loadPrestations()
  }
)

const userLabel = (row: PrestationRow) => {
  const name = `${row.userPrenom ?? ''} ${row.userNom ?? ''}`.trim()
  return name || row.userLogin || row.id.slice(0, 8)
}

const loadPrestations = async () => {
  loadingList.value = true
  error.value = ''
  try {
    const res = await apiFetch<{ data?: PrestationRow[] }>(
      `/api/prestations?month=${encodeURIComponent(month.value)}`
    )
    prestations.value = res?.data ?? []
    selectedIds.value = []
  } catch (e) {
    error.value = extractFetchError(e, t('invoicing.wizard_load_error'))
    prestations.value = []
  } finally {
    loadingList.value = false
  }
}

const toggleSelect = (id: string) => {
  const idx = selectedIds.value.indexOf(id)
  if (idx >= 0) selectedIds.value.splice(idx, 1)
  else selectedIds.value.push(id)
}

const selectAllDefinitive = () => {
  selectedIds.value = definitiveRows.value.map((r) => r.id)
}

const goPreview = async () => {
  if (!selectedIds.value.length) {
    error.value = t('invoicing.wizard_pick_required')
    return
  }
  loadingPreview.value = true
  error.value = ''
  try {
    const res = await apiFetch<{ data?: { previews?: InvoicePreview[] } }>(
      '/api/prestations/preview-invoices',
      {
        method: 'POST',
        body: { timesheetIds: selectedIds.value }
      }
    )
    const previews = res?.data?.previews ?? []
    drafts.value = previews.map((p) => {
      const blockers = p.blockers ?? []
      const hardBlocked = !p.ok && isHardBlocked(blockers)
      return {
        timesheetId: p.timesheetId,
        userLabel: p.userLabel || p.timesheetId.slice(0, 8),
        ok: p.ok,
        editable: isEditablePreview(p),
        hardBlocked,
        blockers,
        clientId: p.clientId ?? '',
        billableHours: p.billableHours != null ? String(p.billableHours) : '',
        unitPriceEur: p.unitPriceCents != null && p.unitPriceCents > 0 ? String(p.unitPriceCents / 100) : '',
        taxRate: p.taxRate != null ? String(p.taxRate) : '20',
        currency: p.currency || 'EUR',
        description: p.description || ''
      }
    })
    step.value = 2
  } catch (e) {
    error.value = extractFetchError(e, t('invoicing.wizard_preview_error'))
  } finally {
    loadingPreview.value = false
  }
}

const lineNet = (d: EditableDraft) => {
  const qty = Number(d.billableHours)
  const price = Number(d.unitPriceEur)
  if (!(qty > 0) || !(price > 0)) return 0
  return qty * price
}

const lineTax = (d: EditableDraft) => {
  const net = lineNet(d)
  const tax = Number(d.taxRate)
  if (!(tax >= 0) || Number.isNaN(tax)) return 0
  return (net * tax) / 100
}

const goConfirm = () => {
  error.value = ''
  const okDrafts = selectableDrafts.value
  if (!okDrafts.length) {
    error.value = t('invoicing.wizard_no_ok')
    return
  }
  for (const d of okDrafts) {
    if (!d.clientId || !(Number(d.billableHours) > 0) || !(Number(d.unitPriceEur) > 0)) {
      error.value = t('invoicing.wizard_edit_validation')
      return
    }
  }
  step.value = 3
}

const onCreate = async () => {
  creating.value = true
  error.value = ''
  try {
    const items = selectableDrafts.value.map((d) => ({
      timesheetId: d.timesheetId,
      clientId: d.clientId,
      billableHours: Number(d.billableHours),
      unitPriceCents: Math.round(Number(d.unitPriceEur) * 100),
      taxRate: Number(d.taxRate),
      currency: d.currency,
      description: d.description.trim()
    }))
    const res = await apiFetch<{
      data?: {
        created?: number
        outcomes?: Array<{ status?: string; invoiceId?: string; timesheetId?: string; reason?: string }>
      }
    }>('/api/prestations/create-invoices', {
      method: 'POST',
      body: { items }
    })
    const outcomes = res?.data?.outcomes ?? []
    const createdIds = outcomes
      .filter((o) => o.status === 'created' && o.invoiceId)
      .map((o) => o.invoiceId as string)
    if (!createdIds.length) {
      const skipped = outcomes
        .filter((o) => o.status !== 'created')
        .map((o) => mapInvoiceDraftReason(o.reason))
        .join(', ')
      error.value = skipped
        ? t('invoicing.wizard_create_none', { details: skipped })
        : t('invoicing.wizard_create_error')
      return
    }
    openModel.value = false
    emit('created', createdIds)
  } catch (e) {
    error.value = extractFetchError(e, t('invoicing.wizard_create_error'))
  } finally {
    creating.value = false
  }
}

const clientLabel = (id: string) => props.clients.find((c) => c.id === id)?.label || id.slice(0, 8)

const formatMoney = (amount: number, currency: string) =>
  new Intl.NumberFormat(undefined, { style: 'currency', currency: currency || 'EUR' }).format(amount)
</script>

<template>
  <AppModal v-model:open="openModel" width="lg" :aria-label="$t('invoicing.wizard_title')">
    <div class="wizard">
      <header class="wizard__head">
        <h2 id="invoice-from-cra-title" class="wizard__title">{{ $t('invoicing.wizard_title') }}</h2>
        <p class="wizard__steps">
          {{ $t('invoicing.wizard_step', { n: step, total: 3 }) }}
        </p>
      </header>

      <!-- Step 1: pick -->
      <div v-if="step === 1" class="wizard__body">
        <label class="wizard__field">
          <span>{{ $t('invoicing.wizard_month') }}</span>
          <input v-model="month" type="month" @change="loadPrestations" >
        </label>

        <div class="wizard__toolbar">
          <p class="wizard__hint">{{ $t('invoicing.wizard_pick_hint') }}</p>
          <AppButton
            variant="ghost"
            size="sm"
            type="button"
            :disabled="!definitiveRows.length"
            @click="selectAllDefinitive"
          >
            {{ $t('invoicing.wizard_select_all') }}
          </AppButton>
        </div>

        <p v-if="loadingList" class="wizard__muted">{{ $t('common.loading') }}</p>
        <p v-else-if="!definitiveRows.length" class="wizard__muted">
          {{ $t('invoicing.wizard_no_definitive') }}
        </p>
        <ul v-else class="wizard__list">
          <li v-for="row in definitiveRows" :key="row.id">
            <label class="wizard__check">
              <input
                type="checkbox"
                :checked="selectedIds.includes(row.id)"
                @change="toggleSelect(row.id)"
              >
              <span>
                <strong>{{ userLabel(row) }}</strong>
                · {{ $t('cra.hours_value', { n: Math.round((row.totalMinutes ?? 0) / 60) }) }}
                · {{ row.id.slice(0, 8) }}
              </span>
            </label>
          </li>
        </ul>
      </div>

      <!-- Step 2: edit preview -->
      <div v-else-if="step === 2" class="wizard__body">
        <article
          v-for="d in drafts"
          :key="d.timesheetId"
          class="wizard__draft"
          :class="{ 'wizard__draft--blocked': d.hardBlocked }"
        >
          <header class="wizard__draft-head">
            <h3>{{ d.userLabel }} · {{ d.timesheetId.slice(0, 8) }}</h3>
            <AppBadge v-if="d.hardBlocked" variant="warning">{{ $t('invoicing.wizard_blocked') }}</AppBadge>
            <AppBadge v-else-if="!d.ok" variant="gold">{{ $t('invoicing.wizard_needs_edit') }}</AppBadge>
          </header>
          <ul v-if="d.blockers.length" class="wizard__blockers">
            <li v-for="b in d.blockers" :key="b">{{ mapInvoiceDraftReason(b) }}</li>
          </ul>
          <template v-if="d.editable">
            <label class="wizard__field">
              <span>{{ $t('invoicing.client') }}</span>
              <select v-model="d.clientId" required>
                <option disabled value="">{{ $t('invoicing.client_placeholder') }}</option>
                <option v-for="c in clients" :key="c.id" :value="c.id">{{ c.label }}</option>
              </select>
            </label>
            <div class="wizard__grid">
              <AppInput
                :id="`wh-${d.timesheetId}`"
                v-model="d.billableHours"
                type="number"
                :label="$t('invoicing.wizard_hours')"
              />
              <AppInput
                :id="`pu-${d.timesheetId}`"
                v-model="d.unitPriceEur"
                type="number"
                :label="$t('invoicing.line_unit_price_eur')"
              />
              <AppInput
                :id="`tax-${d.timesheetId}`"
                v-model="d.taxRate"
                type="number"
                :label="$t('invoicing.line_tax')"
              />
            </div>
            <AppInput
              :id="`desc-${d.timesheetId}`"
              v-model="d.description"
              :label="$t('invoicing.line_description')"
            />
            <p class="wizard__totals">
              {{ $t('invoicing.wizard_line_net', { amount: formatMoney(lineNet(d), d.currency) }) }}
              ·
              {{ $t('invoicing.wizard_line_tax', { amount: formatMoney(lineTax(d), d.currency) }) }}
            </p>
          </template>
        </article>
      </div>

      <!-- Step 3: confirm -->
      <div v-else class="wizard__body">
        <p class="wizard__hint">{{ $t('invoicing.wizard_confirm_hint', { n: selectableDrafts.length }) }}</p>
        <ul class="wizard__summary">
          <li v-for="d in selectableDrafts" :key="d.timesheetId">
            <strong>{{ d.userLabel }}</strong>
            — {{ clientLabel(d.clientId) }}
            · {{ d.billableHours }} h × {{ d.unitPriceEur }} €
            → {{ formatMoney(lineNet(d) + lineTax(d), d.currency) }}
          </li>
        </ul>
      </div>

      <p v-if="error" class="wizard__error" role="alert">{{ error }}</p>

      <footer class="wizard__actions">
        <AppButton variant="ghost" type="button" :disabled="creating" @click="openModel = false">
          {{ $t('common.cancel') }}
        </AppButton>
        <AppButton
          v-if="step > 1"
          variant="ghost"
          type="button"
          :disabled="creating || loadingPreview"
          @click="step = (step - 1) as 1 | 2"
        >
          {{ $t('invoicing.back') }}
        </AppButton>
        <AppButton
          v-if="step === 1"
          variant="primary"
          type="button"
          :disabled="loadingList || loadingPreview || !selectedIds.length"
          @click="goPreview"
        >
          {{ $t('invoicing.wizard_next') }}
        </AppButton>
        <AppButton
          v-else-if="step === 2"
          variant="primary"
          type="button"
          :disabled="!selectableDrafts.length"
          @click="goConfirm"
        >
          {{ $t('invoicing.wizard_next') }}
        </AppButton>
        <AppButton
          v-else
          variant="primary"
          type="button"
          :disabled="creating"
          @click="onCreate"
        >
          {{ $t('invoicing.wizard_create') }}
        </AppButton>
      </footer>
    </div>
  </AppModal>
</template>

<style scoped>
.wizard {
  display: flex;
  flex-direction: column;
  gap: var(--kore-space-md);
  max-height: min(85vh, 720px);
}

.wizard__head {
  display: flex;
  flex-wrap: wrap;
  align-items: baseline;
  justify-content: space-between;
  gap: var(--kore-space-sm);
}

.wizard__title {
  margin: 0;
  font-size: var(--kore-text-h3);
}

.wizard__steps {
  margin: 0;
  color: var(--kore-text-muted);
  font-size: var(--kore-text-small);
}

.wizard__body {
  overflow-y: auto;
  display: flex;
  flex-direction: column;
  gap: var(--kore-space-md);
  flex: 1;
  min-height: 0;
}

.wizard__field {
  display: flex;
  flex-direction: column;
  gap: var(--kore-space-xs);
  font-size: var(--kore-text-small);
}

.wizard__field input,
.wizard__field select {
  padding: var(--kore-space-sm) var(--kore-space-md);
  border: 1px solid var(--kore-border);
  border-radius: var(--kore-radius-md);
  background: var(--kore-bg);
  color: var(--kore-text);
  font-family: var(--kore-font);
}

.wizard__toolbar {
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  justify-content: space-between;
  gap: var(--kore-space-sm);
}

.wizard__hint,
.wizard__muted {
  margin: 0;
  color: var(--kore-text-muted);
  font-size: var(--kore-text-small);
}

.wizard__list,
.wizard__summary,
.wizard__blockers {
  margin: 0;
  padding: 0;
  list-style: none;
  display: flex;
  flex-direction: column;
  gap: var(--kore-space-sm);
}

.wizard__check {
  display: flex;
  align-items: flex-start;
  gap: var(--kore-space-sm);
  padding: var(--kore-space-sm);
  border: 1px solid var(--kore-border);
  border-radius: var(--kore-radius-md);
  cursor: pointer;
}

.wizard__draft {
  display: flex;
  flex-direction: column;
  gap: var(--kore-space-sm);
  padding: var(--kore-space-md);
  border: 1px solid var(--kore-border);
  border-radius: var(--kore-radius-md);
}

.wizard__draft--blocked {
  background: var(--kore-bg-subtle);
}

.wizard__draft-head {
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  justify-content: space-between;
  gap: var(--kore-space-sm);
}

.wizard__draft-head h3 {
  margin: 0;
  font-size: var(--kore-text-small);
}

.wizard__grid {
  display: grid;
  gap: var(--kore-space-sm);
}

@media (min-width: 769px) {
  .wizard__grid {
    grid-template-columns: repeat(3, 1fr);
  }
}

.wizard__totals {
  margin: 0;
  font-size: var(--kore-text-small);
  color: var(--kore-text-muted);
}

.wizard__error {
  margin: 0;
  color: var(--kore-error);
  font-size: var(--kore-text-small);
}

.wizard__actions {
  display: flex;
  flex-wrap: wrap;
  gap: var(--kore-space-sm);
  justify-content: flex-end;
}

@media (max-width: 768px) {
  .wizard__actions :deep(.app-button),
  .wizard__actions button {
    width: 100%;
  }
}
</style>
