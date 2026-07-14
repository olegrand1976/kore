<template>
  <div>
    <AppPageHeader :title="pageTitle" :subtitle="$t('invoicing.detail_subtitle')">
      <template #actions>
        <AppButton variant="ghost" size="sm" @click="navigateTo('/facturation')">
          {{ $t('invoicing.back') }}
        </AppButton>
        <AppButton
          v-if="invoice?.status === 'preparee'"
          variant="primary"
          size="sm"
          :loading="transmitting"
          @click="onTransmit"
        >
          {{ $t('invoicing.transmit') }}
        </AppButton>
      </template>
    </AppPageHeader>

    <p v-if="errorMsg" class="flash flash--error" role="alert">{{ errorMsg }}</p>

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

const route = useRoute()
const { t } = useI18n()
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
  lines?: InvoiceLine[]
}

const { data, pending, refresh } = await useFetch<{ data?: InvoiceDetail } | InvoiceDetail>(
  () => `/api/invoices/${id.value}`
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
const errorMsg = ref('')

const onTransmit = async () => {
  errorMsg.value = ''
  transmitting.value = true
  try {
    await $fetch(`/api/invoices/${id.value}/transmit`, { method: 'POST' })
    await refresh()
  } catch {
    errorMsg.value = t('invoicing.transmit_error')
  } finally {
    transmitting.value = false
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
</style>
