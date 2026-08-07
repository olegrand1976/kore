<template>
  <div>
    <AppPageHeader :title="$t('invoicing.title')" :subtitle="$t('invoicing.subtitle')">
      <template #actions>
        <AppButton v-if="guideRef?.dismissed" variant="ghost" size="sm" type="button" @click="guideRef?.showAgain()">
          {{ $t('guides.show') }}
        </AppButton>
        <AppButton
          v-if="canFromCra && !showCreate && !wizardOpen"
          variant="primary"
          size="sm"
          type="button"
          @click="openCreate('cra')"
        >
          {{ $t('invoicing.create') }}
        </AppButton>
        <AppButton
          v-else-if="canWrite && !canFromCra && !showCreate"
          variant="primary"
          size="sm"
          type="button"
          @click="openCreate('manual')"
        >
          {{ $t('invoicing.create') }}
        </AppButton>
      </template>
    </AppPageHeader>

    <AppSectionGuide ref="guideRef" guide-key="invoicing" />

    <p v-if="flashMsg" class="flash" :class="{ 'flash--error': flashError }" role="status">{{ flashMsg }}</p>

    <AppCard v-if="showCreate" padding="lg" class="create-card">
      <div class="create-modes" role="tablist">
        <button
          v-if="canFromCra"
          type="button"
          role="tab"
          class="create-modes__tab"
          :class="{ 'create-modes__tab--active': createMode === 'cra' }"
          :aria-selected="createMode === 'cra'"
          @click="switchToCraMode"
        >
          {{ $t('invoicing.mode_from_cra') }}
        </button>
        <button
          type="button"
          role="tab"
          class="create-modes__tab"
          :class="{ 'create-modes__tab--active': createMode === 'manual' }"
          :aria-selected="createMode === 'manual'"
          @click="createMode = 'manual'"
        >
          {{ $t('invoicing.mode_manual') }}
        </button>
      </div>

      <template v-if="createMode === 'manual'">
        <h2 class="create-card__title">{{ $t('invoicing.create_title') }}</h2>
        <form class="create-form" @submit.prevent="onCreate">
          <label class="create-form__field">
            <span>{{ $t('invoicing.client') }}</span>
            <select v-model="createForm.clientId" required>
              <option disabled value="">{{ $t('invoicing.client_placeholder') }}</option>
              <option v-for="c in clients" :key="c.id" :value="c.id">{{ c.label }}</option>
            </select>
          </label>

          <label class="create-form__field">
            <span>{{ $t('invoicing.currency') }}</span>
            <select v-model="createForm.currency">
              <option value="EUR">EUR</option>
              <option value="USD">USD</option>
              <option value="GBP">GBP</option>
            </select>
          </label>

          <div class="create-form__lines">
            <div class="create-form__lines-head">
              <h3>{{ $t('invoicing.lines_title') }}</h3>
              <AppButton variant="ghost" size="sm" type="button" @click="addLine">
                {{ $t('invoicing.add_line') }}
              </AppButton>
            </div>
            <div v-for="(line, idx) in createForm.lines" :key="idx" class="create-form__line">
              <AppInput
                :id="`line-desc-${idx}`"
                v-model="line.description"
                :label="$t('invoicing.line_description')"
                required
              />
              <AppInput
                :id="`line-qty-${idx}`"
                v-model="line.quantity"
                type="number"
                :label="$t('invoicing.line_quantity')"
                required
              />
              <AppInput
                :id="`line-price-${idx}`"
                v-model="line.unitPriceEur"
                type="number"
                :label="$t('invoicing.line_unit_price_eur')"
                :placeholder="$t('invoicing.line_unit_price_placeholder')"
                required
              />
              <AppInput
                :id="`line-tax-${idx}`"
                v-model="line.taxRate"
                type="number"
                :label="$t('invoicing.line_tax')"
                required
              />
              <AppButton
                v-if="createForm.lines.length > 1"
                variant="ghost"
                size="sm"
                type="button"
                @click="removeLine(idx)"
              >
                {{ $t('common.delete') }}
              </AppButton>
            </div>
          </div>

          <p class="create-form__preview">
            {{ $t('invoicing.manual_preview_net', { amount: formatMoney(manualNet) }) }}
            ·
            {{ $t('invoicing.manual_preview_tax', { amount: formatMoney(manualTax) }) }}
            ·
            {{ $t('invoicing.manual_preview_ttc', { amount: formatMoney(manualNet + manualTax) }) }}
          </p>

          <div class="create-form__actions">
            <AppButton variant="ghost" type="button" :disabled="creating" @click="showCreate = false">
              {{ $t('common.cancel') }}
            </AppButton>
            <AppButton variant="primary" type="submit" :disabled="creating">
              {{ $t('invoicing.create_submit') }}
            </AppButton>
          </div>
          <p v-if="createError" class="flash flash--error" role="alert">{{ createError }}</p>
        </form>
      </template>
    </AppCard>

    <InvoiceFromCraWizard
      v-model:open="wizardOpen"
      :clients="clients"
      @created="onWizardCreated"
    />

    <AppCard padding="lg">
      <AppTable
        :columns="columns"
        :rows="rows"
        :loading="pending"
        :empty-title="$t('invoicing.empty')"
        :empty-description="$t('invoicing.empty_desc')"
      >
        <template v-if="canWrite" #empty>
          <div class="empty-actions">
            <AppButton v-if="canFromCra" variant="primary" type="button" @click="openCreate('cra')">
              {{ $t('invoicing.empty_cta_cra') }}
            </AppButton>
            <AppButton variant="secondary" type="button" @click="openCreate('manual')">
              {{ $t('invoicing.empty_cta_manual') }}
            </AppButton>
          </div>
        </template>
        <template #cell-client="{ row }">
          {{ clientLabel(String(row.clientId ?? '')) }}
        </template>
        <template #cell-status="{ row }">
          <AppBadge :variant="statusVariant(String(row.status))">{{ statusLabel(String(row.status)) }}</AppBadge>
        </template>
        <template #cell-amount="{ row }">
          {{ formatAmount(Number(row.totalAmount ?? 0), String(row.currency ?? 'EUR')) }}
        </template>
        <template #cell-actions="{ row }">
          <AppButton variant="ghost" size="sm" @click="navigateTo(`/facturation/${row.rawId}`)">
            {{ $t('invoicing.open') }}
          </AppButton>
        </template>
      </AppTable>
    </AppCard>
  </div>
</template>

<script setup lang="ts">
definePageMeta({ layout: 'default' })

const guideRef = ref<{ showAgain: () => void; dismissed: boolean } | null>(null)

const { t } = useI18n()
const { apiFetch } = useApiFetch()
const { extractFetchError } = useApiError()
const { can } = usePermissions()
const { listClients, orgId } = useOrganisation()
const { isInvoicingEnabled, fetchSettings } = useRequestSettings()

await fetchSettings()
const invoicingAllowed = isInvoicingEnabled.value && can('invoicing', 'L')
if (!invoicingAllowed) {
  await navigateTo('/dashboard')
}

const canWrite = computed(() => can('invoicing', 'E'))
const canFromCra = computed(() => canWrite.value && can('cra', 'V'))

type InvoiceRow = {
  id: string
  clientId: string
  status: string
  type: string
  currency: string
  totalAmount: number
  createdAt: string
}

const { data, pending, refresh } = await useFetch<{ data?: InvoiceRow[] } | InvoiceRow[]>('/api/invoices', {
  immediate: invoicingAllowed
})

const invoices = computed(() => {
  const raw = data.value
  if (Array.isArray(raw)) return raw
  return raw?.data ?? []
})

const columns = computed(() => [
  { key: 'id', label: t('invoicing.col_id') },
  { key: 'client', label: t('invoicing.col_client') },
  { key: 'status', label: t('invoicing.col_status') },
  { key: 'type', label: t('invoicing.col_type') },
  { key: 'amount', label: t('invoicing.col_amount') },
  { key: 'actions', label: t('invoicing.col_actions'), nowrap: true }
])

const rows = computed(() =>
  invoices.value.map((inv) => {
    const rawId = inv.id ?? ''
    return {
      id: rawId.slice(0, 8) || '—',
      rawId,
      clientId: inv.clientId,
      status: inv.status,
      type: inv.type,
      currency: inv.currency ?? 'EUR',
      totalAmount: inv.totalAmount ?? 0,
      createdAt: inv.createdAt
    }
  })
)

type ClientOption = { id: string; label: string }
type LineForm = {
  description: string
  quantity: string
  unitPriceEur: string
  taxRate: string
}

const showCreate = ref(false)
const createMode = ref<'cra' | 'manual'>('cra')
const wizardOpen = ref(false)
const creating = ref(false)
const createError = ref('')
const flashMsg = ref('')
const flashError = ref(false)
const clients = ref<ClientOption[]>([])
const createForm = reactive({
  clientId: '',
  currency: 'EUR',
  lines: [
    { description: '', quantity: '1', unitPriceEur: '', taxRate: '20' }
  ] as LineForm[]
})

const openCreate = (mode: 'cra' | 'manual') => {
  createMode.value = mode
  if (mode === 'cra') {
    if (!canFromCra.value) return
    wizardOpen.value = true
    showCreate.value = false
    return
  }
  showCreate.value = true
  wizardOpen.value = false
}

const switchToCraMode = () => {
  if (!canFromCra.value) return
  showCreate.value = false
  wizardOpen.value = true
  createMode.value = 'cra'
}

onMounted(async () => {
  try {
    const list = await listClients()
    clients.value = list.map((c) => ({
      id: orgId(c),
      label: c.raisonSociale ?? c.RaisonSociale ?? orgId(c)
    })).filter((c) => c.id)
  } catch {
    clients.value = []
  }
})

const addLine = () => {
  createForm.lines.push({ description: '', quantity: '1', unitPriceEur: '', taxRate: '20' })
}

const removeLine = (idx: number) => {
  if (createForm.lines.length <= 1) return
  createForm.lines.splice(idx, 1)
}

const manualNet = computed(() =>
  createForm.lines.reduce((sum, line) => {
    const qty = Number(line.quantity)
    const price = Number(line.unitPriceEur)
    if (!(qty > 0) || !(price > 0)) return sum
    return sum + qty * price
  }, 0)
)

const manualTax = computed(() =>
  createForm.lines.reduce((sum, line) => {
    const qty = Number(line.quantity)
    const price = Number(line.unitPriceEur)
    const tax = Number(line.taxRate)
    if (!(qty > 0) || !(price > 0) || !(tax >= 0) || Number.isNaN(tax)) return sum
    return sum + (qty * price * tax) / 100
  }, 0)
)

const formatMoney = (amount: number) =>
  new Intl.NumberFormat(undefined, { style: 'currency', currency: createForm.currency || 'EUR' }).format(amount)

const clientLabel = (id: string) => {
  if (!id) return '—'
  return clients.value.find((c) => c.id === id)?.label || id.slice(0, 8)
}

const onCreate = async () => {
  creating.value = true
  createError.value = ''
  try {
    const lines = createForm.lines.map((line) => ({
      description: line.description.trim(),
      quantity: Number(line.quantity),
      unitPrice: Math.round(Number(line.unitPriceEur) * 100),
      taxRate: Number(line.taxRate)
    }))
    if (lines.some((l) => !l.description || !(l.quantity > 0) || !(l.unitPrice > 0) || Number.isNaN(l.taxRate) || l.taxRate < 0)) {
      createError.value = t('invoicing.create_validation')
      return
    }
    const res = await apiFetch<{ data?: { id?: string } }>('/api/invoices', {
      method: 'POST',
      body: {
        clientId: createForm.clientId,
        type: 'standard',
        currency: createForm.currency,
        lines
      }
    })
    const id = res.data?.id
    showCreate.value = false
    await refresh()
    if (id) await navigateTo(`/facturation/${id}`)
  } catch (e) {
    createError.value = extractFetchError(e, t('invoicing.create_error'))
  } finally {
    creating.value = false
  }
}

const onWizardCreated = async (invoiceIds: string[]) => {
  showCreate.value = false
  flashError.value = false
  flashMsg.value = t('invoicing.wizard_created_flash', { n: invoiceIds.length })
  await refresh()
  if (invoiceIds.length === 1) {
    await navigateTo(`/facturation/${invoiceIds[0]}`)
  }
}

const statusVariant = (status: string) => {
  switch (status) {
    case 'acceptee':
    case 'encaissee':
      return 'success'
    case 'refusee':
    case 'proforma_refusee':
    case 'annulee':
      return 'error'
    case 'transmise':
    case 'proforma':
      return 'blue'
    case 'preparee':
      return 'gold'
    default:
      return 'neutral'
  }
}

const statusLabel = (status: string) => t(`invoicing.status.${status}`, status)

const formatAmount = (cents: number, currency: string) =>
  new Intl.NumberFormat(undefined, { style: 'currency', currency }).format(cents / 100)
</script>

<style scoped>
.create-card {
  margin-bottom: var(--kore-space-lg);
}

.create-card__title {
  margin: 0 0 var(--kore-space-md);
  font-size: var(--kore-text-h3);
}

.create-modes {
  display: flex;
  flex-wrap: wrap;
  gap: var(--kore-space-xs);
  margin-bottom: var(--kore-space-md);
}

.create-modes__tab {
  flex: 1;
  min-width: 8rem;
  padding: var(--kore-space-sm) var(--kore-space-md);
  border: 1px solid var(--kore-border);
  border-radius: var(--kore-radius-md);
  background: var(--kore-bg);
  color: var(--kore-text);
  font-family: var(--kore-font);
  cursor: pointer;
}

.create-modes__tab--active {
  border-color: var(--kore-accent);
  background: var(--kore-bg-subtle);
  font-weight: 600;
}

.empty-actions {
  display: flex;
  flex-wrap: wrap;
  gap: var(--kore-space-sm);
  margin-top: var(--kore-space-md);
}

.create-form {
  display: flex;
  flex-direction: column;
  gap: var(--kore-space-md);
  max-width: var(--kore-form-wide-max);
}

.create-form__field {
  display: flex;
  flex-direction: column;
  gap: var(--kore-space-xs);
  font-size: var(--kore-text-small);
  color: var(--kore-text);
}

.create-form__field select {
  padding: var(--kore-space-sm) var(--kore-space-md);
  border: 1px solid var(--kore-border);
  border-radius: var(--kore-radius-md);
  background: var(--kore-bg);
  color: var(--kore-text);
  font-family: var(--kore-font);
}

.create-form__lines {
  display: flex;
  flex-direction: column;
  gap: var(--kore-space-md);
}

.create-form__lines-head {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: var(--kore-space-sm);
}

.create-form__lines-head h3 {
  margin: 0;
  font-size: var(--kore-text-small);
}

.create-form__line {
  display: grid;
  gap: var(--kore-space-sm);
  padding: var(--kore-space-md);
  border: 1px solid var(--kore-border);
  border-radius: var(--kore-radius-md);
}

.create-form__preview {
  margin: 0;
  font-size: var(--kore-text-small);
  color: var(--kore-text-muted);
}

.create-form__actions {
  display: flex;
  flex-wrap: wrap;
  gap: var(--kore-space-sm);
  justify-content: flex-end;
}

@media (max-width: 768px) {
  .create-form__actions :deep(.app-button),
  .create-form__actions button,
  .empty-actions :deep(.app-button) {
    width: 100%;
  }
}

@media (min-width: 769px) {
  .create-form__line {
    grid-template-columns: 2fr 1fr 1fr 1fr auto;
    align-items: end;
  }
}

.flash {
  margin: 0 0 var(--kore-space-md);
  font-size: var(--kore-text-small);
}

.flash--error {
  color: var(--kore-error);
}
</style>
