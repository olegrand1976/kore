<template>
  <div class="proforma-page">
    <PublicSection>
      <h1 class="proforma-page__title">{{ $t('proforma.title') }}</h1>
      <p class="proforma-page__lead">{{ $t('proforma.lead') }}</p>

      <p v-if="errorMsg" class="flash flash--error" role="alert">{{ errorMsg }}</p>
      <p v-else-if="validated" class="flash flash--ok" role="status">{{ $t('proforma.validated') }}</p>

      <PublicCard v-if="pending" padding="lg">
        <p class="muted">{{ $t('common.loading') }}</p>
      </PublicCard>

      <PublicCard v-else-if="preview" padding="lg" class="proforma-card">
        <dl class="meta">
          <div>
            <dt>{{ $t('proforma.client') }}</dt>
            <dd>{{ preview.clientName || '—' }}</dd>
          </div>
          <div>
            <dt>{{ $t('proforma.amount_ht') }}</dt>
            <dd>{{ formatAmount(preview.totalAmount, preview.currency) }}</dd>
          </div>
          <div>
            <dt>{{ $t('proforma.amount_tax') }}</dt>
            <dd>{{ formatAmount(preview.taxAmount, preview.currency) }}</dd>
          </div>
          <div>
            <dt>{{ $t('proforma.amount_ttc') }}</dt>
            <dd>{{ formatAmount(preview.totalAmount + preview.taxAmount, preview.currency) }}</dd>
          </div>
          <div v-if="preview.expiresAt">
            <dt>{{ $t('proforma.expires') }}</dt>
            <dd>{{ formatDate(preview.expiresAt) }}</dd>
          </div>
        </dl>

        <ul v-if="preview.lines?.length" class="lines">
          <li v-for="(line, idx) in preview.lines" :key="idx">
            <span>{{ line.description }}</span>
            <span class="lines__qty">{{ line.quantity }} × {{ formatAmount(line.unitPrice, preview.currency) }}</span>
          </li>
        </ul>

        <PublicButton
          v-if="!validated"
          variant="primary"
          class="proforma-card__cta"
          :disabled="validating"
          @click="onValidate"
        >
          {{ validating ? $t('proforma.validating') : $t('proforma.validate') }}
        </PublicButton>
      </PublicCard>
    </PublicSection>
  </div>
</template>

<script setup lang="ts">
definePageMeta({ layout: 'public' })

const route = useRoute()
const { t } = useI18n()
const token = computed(() => String(route.params.token || ''))

type ProformaLine = {
  description: string
  quantity: number
  unitPrice: number
  taxRate: number
}

type ProformaPreview = {
  invoiceId: string
  clientName?: string
  currency: string
  totalAmount: number
  taxAmount: number
  status: string
  expiresAt?: string
  lines?: ProformaLine[]
  validated?: boolean
}

const pending = ref(true)
const validating = ref(false)
const validated = ref(false)
const errorMsg = ref('')
const preview = ref<ProformaPreview | null>(null)

const unwrap = <T>(raw: { data?: T } | T | null | undefined): T | null => {
  if (!raw) return null
  if (typeof raw === 'object' && raw !== null && 'data' in raw && (raw as { data?: T }).data) {
    return (raw as { data: T }).data
  }
  return raw as T
}

const load = async () => {
  pending.value = true
  errorMsg.value = ''
  try {
    const raw = await $fetch<{ data?: ProformaPreview } | ProformaPreview>(
      `/api/public/proforma/${encodeURIComponent(token.value)}`
    )
    preview.value = unwrap(raw)
  } catch {
    preview.value = null
    errorMsg.value = t('proforma.load_error')
  } finally {
    pending.value = false
  }
}

await load()

const onValidate = async () => {
  validating.value = true
  errorMsg.value = ''
  try {
    const raw = await $fetch<{ data?: ProformaPreview } | ProformaPreview>(
      `/api/public/proforma/${encodeURIComponent(token.value)}/validate`,
      { method: 'POST' }
    )
    preview.value = unwrap(raw)
    validated.value = true
  } catch {
    errorMsg.value = t('proforma.validate_error')
  } finally {
    validating.value = false
  }
}

const formatAmount = (cents: number, currency: string) =>
  new Intl.NumberFormat(undefined, { style: 'currency', currency: currency || 'EUR' }).format(cents / 100)

const formatDate = (iso: string) =>
  new Intl.DateTimeFormat(undefined, { dateStyle: 'medium', timeStyle: 'short' }).format(new Date(iso))
</script>

<style scoped>
.proforma-page {
  padding: var(--kore-space-xl) 0;
}

.proforma-page__title {
  margin: 0 0 var(--kore-space-sm);
  font-size: var(--kore-text-h1);
  color: var(--kore-text);
}

.proforma-page__lead {
  margin: 0 0 var(--kore-space-lg);
  color: var(--kore-text-muted);
  max-width: var(--kore-prose-max);
}

.muted {
  margin: 0;
  color: var(--kore-text-muted);
}

.meta {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(9rem, 1fr));
  gap: var(--kore-space-md);
  margin: 0 0 var(--kore-space-lg);
}

.meta dt {
  font-size: var(--kore-text-small);
  color: var(--kore-text-muted);
}

.meta dd {
  margin: var(--kore-space-xs) 0 0;
  color: var(--kore-text);
}

.lines {
  list-style: none;
  margin: 0 0 var(--kore-space-lg);
  padding: 0;
  display: grid;
  gap: var(--kore-space-sm);
}

.lines li {
  display: flex;
  flex-wrap: wrap;
  justify-content: space-between;
  gap: var(--kore-space-sm);
  padding-bottom: var(--kore-space-sm);
  border-bottom: 1px solid var(--kore-border);
}

.lines__qty {
  color: var(--kore-text-muted);
  font-size: var(--kore-text-small);
}

.proforma-card__cta {
  width: 100%;
}

.flash {
  margin: 0 0 var(--kore-space-md);
  padding: var(--kore-space-sm) var(--kore-space-md);
  border-radius: var(--kore-radius-md);
}

.flash--error {
  background: color-mix(in srgb, var(--kore-error) 16%, transparent);
  color: var(--kore-text);
}

.flash--ok {
  background: color-mix(in srgb, var(--kore-success) 16%, transparent);
  color: var(--kore-text);
}

@media (min-width: 769px) {
  .proforma-card__cta {
    width: auto;
  }
}
</style>
