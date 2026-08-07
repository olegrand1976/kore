<template>
  <div>
    <AppPageHeader :title="pageTitle" :subtitle="$t('invoicing.detail_subtitle')">
      <template #actions>
        <AppButton variant="ghost" size="sm" @click="navigateTo('/facturation')">
          {{ $t('invoicing.back') }}
        </AppButton>
        <AppButton
          v-if="invoice?.status === 'preparee' || invoice?.status === 'proforma'"
          variant="secondary"
          size="sm"
          :disabled="emittingProforma"
          @click="onEmitProforma"
        >
          {{ invoice?.status === 'proforma' ? $t('invoicing.proforma_resend') : $t('invoicing.proforma_emit') }}
        </AppButton>
        <AppButton
          v-if="invoice?.status === 'preparee'"
          variant="primary"
          size="sm"
          :disabled="transmitting"
          @click="onTransmit"
        >
          {{ $t('invoicing.transmit') }}
        </AppButton>
      </template>
    </AppPageHeader>

    <p v-if="errorMsg" class="flash flash--error" role="alert">{{ errorMsg }}</p>
    <p v-if="successMsg" class="flash flash--ok" role="status">{{ successMsg }}</p>

    <AppCard v-if="pending" padding="lg"><p class="muted">{{ $t('common.loading') }}</p></AppCard>

    <AppCard v-else-if="!invoice" padding="lg">
      <AppEmptyState icon="error" :title="$t('invoicing.not_found')" />
    </AppCard>

    <template v-else>
      <AppCard padding="lg" class="mb">
        <dl class="meta-grid">
          <div>
            <dt>{{ $t('invoicing.col_status') }}</dt>
            <dd><AppBadge :variant="statusVariant(invoice.status)">{{ statusLabel(invoice.status) }}</AppBadge></dd>
          </div>
          <div>
            <dt>{{ $t('invoicing.col_type') }}</dt>
            <dd>{{ invoice.type }}</dd>
          </div>
          <div>
            <dt>{{ $t('invoicing.col_amount') }}</dt>
            <dd>{{ formatAmount(invoice.totalAmount, invoice.currency) }}</dd>
          </div>
          <div v-if="invoice.pdpReceiptId">
            <dt>{{ $t('invoicing.pdp_receipt') }}</dt>
            <dd class="mono">{{ invoice.pdpReceiptId }}</dd>
          </div>
          <div v-if="invoice.proformaRecipientEmail">
            <dt>{{ $t('invoicing.proforma_recipient') }}</dt>
            <dd>{{ invoice.proformaRecipientEmail }}</dd>
          </div>
          <div v-if="invoice.proformaSentAt">
            <dt>{{ $t('invoicing.proforma_sent_at') }}</dt>
            <dd>{{ formatDate(invoice.proformaSentAt) }}</dd>
          </div>
          <div v-if="invoice.proformaValidatedAt">
            <dt>{{ $t('invoicing.proforma_validated_at') }}</dt>
            <dd>{{ formatDate(invoice.proformaValidatedAt) }}</dd>
          </div>
          <div v-if="invoice.invoiceSentAt">
            <dt>{{ $t('invoicing.invoice_sent_at') }}</dt>
            <dd>{{ formatDate(invoice.invoiceSentAt) }}</dd>
          </div>
        </dl>
      </AppCard>

      <AppCard padding="lg">
        <h2 class="section-title">{{ $t('invoicing.lines_title') }}</h2>
        <AppTable
          :columns="lineColumns"
          :rows="lineRows"
          :empty-title="$t('invoicing.no_lines')"
        />
      </AppCard>
    </template>
  </div>
</template>

<script setup lang="ts">
definePageMeta({ layout: 'default' })

const { apiFetch } = useApiFetch()
const route = useRoute()
const { t } = useI18n()
const { can } = usePermissions()
const { isInvoicingEnabled, fetchSettings } = useRequestSettings()

await fetchSettings()
const invoicingAllowed = isInvoicingEnabled.value && can('invoicing', 'L')
if (!invoicingAllowed) {
  await navigateTo('/dashboard')
}

const id = computed(() => String(route.params.id))

type InvoiceLine = {
  description: string
  quantity: number
  unitPrice: number
  taxRate: number
}

type InvoiceDetail = {
  id: string
  status: string
  type: string
  currency: string
  totalAmount: number
  taxAmount: number
  pdpReceiptId?: string
  proformaRecipientEmail?: string
  proformaSentAt?: string
  proformaValidatedAt?: string
  invoiceSentAt?: string
  lines?: InvoiceLine[]
}

const { data, pending, refresh } = await useFetch<{ data?: InvoiceDetail } | InvoiceDetail>(
  () => `/api/invoices/${id.value}`,
  { immediate: invoicingAllowed }
)

const invoice = computed(() => {
  const raw = data.value
  if (!raw) return null
  if ('data' in raw && raw.data) return raw.data
  return raw as InvoiceDetail
})

const pageTitle = computed(() => t('invoicing.detail_title', { id: id.value.slice(0, 8) }))

const lineColumns = computed(() => [
  { key: 'description', label: t('invoicing.line_description') },
  { key: 'quantity', label: t('invoicing.line_quantity') },
  { key: 'unitPrice', label: t('invoicing.line_unit_price') },
  { key: 'taxRate', label: t('invoicing.line_tax') }
])

const lineRows = computed(() =>
  (invoice.value?.lines ?? []).map((line) => ({
    description: line.description,
    quantity: line.quantity,
    unitPrice: formatAmount(line.unitPrice, invoice.value?.currency ?? 'EUR'),
    taxRate: `${line.taxRate} %`
  }))
)

const transmitting = ref(false)
const emittingProforma = ref(false)
const errorMsg = ref('')
const successMsg = ref('')

const onTransmit = async () => {
  errorMsg.value = ''
  successMsg.value = ''
  transmitting.value = true
  try {
    await apiFetch(`/api/invoices/${id.value}/transmit`, { method: 'POST' })
    await refresh()
  } catch {
    errorMsg.value = t('invoicing.transmit_error')
  } finally {
    transmitting.value = false
  }
}

const onEmitProforma = async () => {
  errorMsg.value = ''
  successMsg.value = ''
  emittingProforma.value = true
  try {
    await apiFetch(`/api/invoices/${id.value}/emit-proforma`, { method: 'POST', body: {} })
    successMsg.value = t('invoicing.proforma_sent')
    await refresh()
  } catch {
    errorMsg.value = t('invoicing.proforma_error')
  } finally {
    emittingProforma.value = false
  }
}

const statusVariant = (status: string) => {
  switch (status) {
    case 'acceptee':
    case 'encaissee':
      return 'success'
    case 'refusee':
    case 'annulee':
      return 'error'
    case 'transmise':
    case 'proforma':
      return 'blue'
    case 'preparee':
      return 'gold'
    case 'virtuelle':
      return 'neutral'
    default:
      return 'neutral'
  }
}

const statusLabel = (status: string) => t(`invoicing.status.${status}`, status)

const formatAmount = (cents: number, currency: string) =>
  new Intl.NumberFormat(undefined, { style: 'currency', currency }).format(cents / 100)

const formatDate = (iso: string) =>
  new Intl.DateTimeFormat(undefined, { dateStyle: 'medium', timeStyle: 'short' }).format(new Date(iso))
</script>

<style scoped>
.mb {
  margin-bottom: var(--kore-space-lg);
}

.muted {
  margin: 0;
  color: var(--kore-text-muted);
}

.meta-grid {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(10rem, 1fr));
  gap: var(--kore-space-md);
  margin: 0;
}

.meta-grid dt {
  font-size: var(--kore-text-small);
  color: var(--kore-text-muted);
}

.meta-grid dd {
  margin: var(--kore-space-xs) 0 0;
}

.section-title {
  margin: 0 0 var(--kore-space-md);
  font-size: var(--kore-text-h3);
}

.mono {
  font-family: var(--kore-font-mono, monospace);
  font-size: var(--kore-text-small);
  word-break: break-all;
}

.flash {
  margin: 0 0 var(--kore-space-md);
  padding: var(--kore-space-sm) var(--kore-space-md);
  border-radius: var(--kore-radius-md);
}

.flash--error {
  background: color-mix(in srgb, var(--kore-error) 12%, transparent);
  color: var(--kore-error);
}

.flash--ok {
  background: color-mix(in srgb, var(--kore-success) 12%, transparent);
  color: var(--kore-success);
}
</style>
