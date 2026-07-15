<template>
  <div>
    <AppPageHeader :title="$t('invoicing.title')" :subtitle="$t('invoicing.subtitle')">
      <template #actions>
        <AppButton v-if="guideRef?.dismissed" variant="ghost" size="sm" type="button" @click="guideRef?.showAgain()">
          {{ $t('guides.show') }}
        </AppButton>
      </template>
    </AppPageHeader>

    <AppSectionGuide ref="guideRef" guide-key="invoicing" />

    <AppCard padding="lg">
      <AppTable
        :columns="columns"
        :rows="rows"
        :loading="pending"
        :empty-title="$t('invoicing.empty')"
        :empty-description="$t('invoicing.empty_desc')"
      >
        <template #cell-status="{ row }">
          <AppBadge :variant="statusVariant(row.status)">{{ statusLabel(row.status) }}</AppBadge>
        </template>
        <template #cell-amount="{ row }">
          {{ formatAmount(row.totalAmount, row.currency) }}
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

type InvoiceRow = {
  id: string
  clientId: string
  status: string
  type: string
  currency: string
  totalAmount: number
  createdAt: string
}

const { data, pending } = await useFetch<{ data?: InvoiceRow[] } | InvoiceRow[]>('/api/invoices')

const invoices = computed(() => {
  const raw = data.value
  if (Array.isArray(raw)) return raw
  return raw?.data ?? []
})

const columns = computed(() => [
  { key: 'id', label: t('invoicing.col_id') },
  { key: 'status', label: t('invoicing.col_status') },
  { key: 'type', label: t('invoicing.col_type') },
  { key: 'amount', label: t('invoicing.col_amount') },
  { key: 'actions', label: t('invoicing.col_actions'), nowrap: true }
])

const rows = computed(() =>
  invoices.value.map((inv) => ({
    id: inv.id.slice(0, 8),
    rawId: inv.id,
    clientId: inv.clientId,
    status: inv.status,
    type: inv.type,
    currency: inv.currency ?? 'EUR',
    totalAmount: inv.totalAmount ?? 0,
    createdAt: inv.createdAt
  }))
)

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
