<template>
  <div class="proforma-page">
    <PublicSection>
      <h1 class="proforma-page__title">{{ $t('proforma.title') }}</h1>
      <p class="proforma-page__lead">{{ $t('proforma.lead') }}</p>

      <p v-if="errorMsg" class="flash flash--error" role="alert">{{ errorMsg }}</p>
      <p v-else-if="validated" class="flash flash--ok" role="status">
        {{ emailWarning ? $t('proforma.validated_no_email') : $t('proforma.validated') }}
      </p>
      <p v-else-if="rejected" class="flash flash--error" role="status">{{ $t('proforma.rejected') }}</p>

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

        <div v-if="!validated && !rejected" class="proforma-form">
          <label class="proforma-form__label" for="proforma-comment">{{ $t('proforma.comment_label') }}</label>
          <textarea
            id="proforma-comment"
            v-model="comment"
            class="proforma-form__textarea"
            rows="3"
            :placeholder="$t('proforma.comment_placeholder')"
          />
          <p class="proforma-form__hint">{{ $t('proforma.comment_hint') }}</p>
          <div class="proforma-form__actions">
            <PublicButton
              variant="primary"
              class="proforma-card__cta"
              :disabled="busy"
              @click="onValidate"
            >
              {{ busy === 'validate' ? $t('proforma.validating') : $t('proforma.validate') }}
            </PublicButton>
            <PublicButton
              variant="danger"
              class="proforma-card__cta"
              :disabled="busy"
              @click="onReject"
            >
              {{ busy === 'reject' ? $t('proforma.rejecting') : $t('proforma.reject') }}
            </PublicButton>
          </div>
        </div>
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
  rejected?: boolean
  comment?: string
  invoiceEmailSent?: boolean
}

const pending = ref(true)
const busy = ref<'validate' | 'reject' | ''>('')
const validated = ref(false)
const rejected = ref(false)
const errorMsg = ref('')
const comment = ref('')
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
  busy.value = 'validate'
  errorMsg.value = ''
  try {
    const raw = await $fetch<{ data?: ProformaPreview } | ProformaPreview>(
      `/api/public/proforma/${encodeURIComponent(token.value)}/validate`,
      { method: 'POST', body: { comment: comment.value } }
    )
    preview.value = unwrap(raw)
    validated.value = true
    rejected.value = false
  } catch {
    errorMsg.value = t('proforma.validate_error')
  } finally {
    busy.value = ''
  }
}

const onReject = async () => {
  if (!comment.value.trim()) {
    errorMsg.value = t('proforma.reject_comment_required')
    return
  }
  busy.value = 'reject'
  errorMsg.value = ''
  try {
    const raw = await $fetch<{ data?: ProformaPreview } | ProformaPreview>(
      `/api/public/proforma/${encodeURIComponent(token.value)}/reject`,
      { method: 'POST', body: { comment: comment.value } }
    )
    preview.value = unwrap(raw)
    rejected.value = true
    validated.value = false
  } catch {
    errorMsg.value = t('proforma.reject_error')
  } finally {
    busy.value = ''
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

.proforma-form {
  display: grid;
  gap: var(--kore-space-sm);
}

.proforma-form__label {
  font-weight: 600;
  color: var(--kore-text);
}

.proforma-form__textarea {
  width: 100%;
  max-width: var(--kore-form-wide-max);
  padding: var(--kore-space-sm) var(--kore-space-md);
  border: 1px solid var(--kore-border);
  border-radius: var(--kore-radius-md);
  background: var(--kore-bg);
  color: var(--kore-text);
  font-family: var(--kore-font);
  resize: vertical;
}

.proforma-form__hint {
  margin: 0;
  font-size: var(--kore-text-small);
  color: var(--kore-text-muted);
}

.proforma-form__actions {
  display: flex;
  flex-wrap: wrap;
  gap: var(--kore-space-sm);
  margin-top: var(--kore-space-sm);
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
