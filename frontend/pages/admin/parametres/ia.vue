<template>
  <div>
    <AppCard padding="lg" class="ia-status">
      <div class="ia-status__row">
        <AppBadge :variant="enabled ? 'success' : 'default'">
          {{ enabled ? $t('settings.ia.status_enabled') : $t('settings.ia.status_disabled') }}
        </AppBadge>
        <span v-if="provider" class="ia-status__provider">
          {{ $t('settings.ia.provider') }} : <strong>{{ provider }}</strong>
        </span>
      </div>
      <p v-if="noticeAcceptedAt" class="ia-status__meta">
        {{ $t('settings.ia.notice_accepted', { date: formatDate(noticeAcceptedAt) }) }}
      </p>
    </AppCard>

    <AppCard padding="lg" class="ia-deployer">
      <h3 class="ia-deployer__title">{{ $t('settings.ia.deployer_title') }}</h3>
      <p class="ia-deployer__text">{{ $t('settings.ia.deployer_intro') }}</p>
      <ul class="ia-deployer__list">
        <li>{{ $t('settings.ia.deployer_supervision') }}</li>
        <li>{{ $t('settings.ia.deployer_transparency') }}</li>
        <li>{{ $t('settings.ia.deployer_logging') }}</li>
        <li>{{ $t('settings.ia.notice_duties') }}</li>
      </ul>
    </AppCard>

    <AppCard v-if="!enabled" padding="lg" class="ia-notice">
      <h3 class="ia-notice__title">{{ $t('settings.ia.notice_title') }}</h3>
      <p class="ia-notice__text">{{ $t('settings.ia.notice_intro') }}</p>
      <ul class="ia-notice__list">
        <li>{{ $t('settings.ia.notice_capabilities') }}</li>
        <li>{{ $t('settings.ia.notice_limits') }}</li>
        <li>{{ $t('settings.ia.notice_duties') }}</li>
      </ul>
      <form class="ia-notice__form" @submit.prevent="enable">
        <label class="ia-toggle">
          <input v-model="noticeAccepted" type="checkbox" required />
          {{ $t('settings.ia.accept_notice') }}
        </label>
        <label class="ia-toggle">
          <input v-model="workersInformed" type="checkbox" required />
          {{ $t('settings.ia.workers_informed') }}
        </label>
        <div class="ia-notice__actions">
          <AppButton variant="primary" size="sm" type="submit" :disabled="enabling || !noticeAccepted || !workersInformed">
            {{ $t('settings.ia.enable') }}
          </AppButton>
        </div>
      </form>
      <p v-if="formError" class="ia-flash ia-flash--error" role="alert">{{ formError }}</p>
    </AppCard>

    <AppCard v-else padding="lg">
      <p class="ia-enabled-text">{{ $t('settings.ia.enabled_info') }}</p>
    </AppCard>

    <p v-if="flash" class="ia-flash" :class="{ 'ia-flash--error': flashError }" role="status">{{ flash }}</p>
  </div>
</template>

<script setup lang="ts">
type TenantAISettings = {
  enabled?: boolean
  Enabled?: boolean
  noticeAcceptedAt?: string
  NoticeAcceptedAt?: string
  workersInformedAt?: string
  WorkersInformedAt?: string
  llmProvider?: string
  LLMProvider?: string
}

definePageMeta({ layout: 'default', middleware: 'admin' })

const { t, locale } = useI18n()
const { extractFetchError } = useApiError()

const { data, refresh, pending } = await useFetch('/api/ai/settings')

const noticeAccepted = ref(false)
const workersInformed = ref(false)
const enabling = ref(false)
const formError = ref('')
const flash = ref('')
const flashError = ref(false)

const settings = computed(() => (data.value as { data?: TenantAISettings })?.data ?? {})

const enabled = computed(() => settings.value.enabled ?? settings.value.Enabled ?? false)
const provider = computed(() => settings.value.llmProvider ?? settings.value.LLMProvider ?? '')
const noticeAcceptedAt = computed(() => settings.value.noticeAcceptedAt ?? settings.value.NoticeAcceptedAt ?? '')

const formatDate = (iso: string) => {
  if (!iso) return '—'
  return new Intl.DateTimeFormat(locale.value, { dateStyle: 'medium', timeStyle: 'short' }).format(new Date(iso))
}

const enable = async () => {
  enabling.value = true
  formError.value = ''
  try {
    await $fetch('/api/ai/settings/enable', {
      method: 'POST',
      body: {
        noticeAccepted: noticeAccepted.value,
        workersInformed: workersInformed.value
      }
    })
    flash.value = t('settings.ia.enabled_success')
    flashError.value = false
    await refresh()
  } catch (e) {
    formError.value = extractFetchError(e)
  } finally {
    enabling.value = false
  }
}

watch(pending, (isPending) => {
  if (!isPending && enabled.value) {
    noticeAccepted.value = true
    workersInformed.value = true
  }
})
</script>

<style scoped>
.ia-status { margin-bottom: var(--kore-space-lg); }
.ia-deployer { margin-bottom: var(--kore-space-lg); }
.ia-deployer__title {
  margin: 0 0 var(--kore-space-sm);
  font-size: var(--kore-text-h3);
}
.ia-deployer__text,
.ia-deployer__list {
  margin: 0 0 var(--kore-space-md);
  color: var(--kore-text-muted);
  font-size: var(--kore-text-small);
  line-height: 1.5;
}
.ia-deployer__list {
  padding-left: 1.25rem;
  margin-bottom: 0;
}
.ia-status__row {
  display: flex;
  flex-wrap: wrap;
  gap: var(--kore-space-md);
  align-items: center;
}
.ia-status__provider,
.ia-status__meta {
  font-size: var(--kore-text-small);
  color: var(--kore-text-muted);
}
.ia-notice__title {
  margin: 0 0 var(--kore-space-sm);
  font-size: var(--kore-text-h3);
}
.ia-notice__text,
.ia-enabled-text {
  margin: 0 0 var(--kore-space-md);
  color: var(--kore-text-muted);
  font-size: var(--kore-text-small);
  line-height: 1.5;
}
.ia-notice__list {
  margin: 0 0 var(--kore-space-lg);
  padding-left: 1.25rem;
  color: var(--kore-text-muted);
  font-size: var(--kore-text-small);
}
.ia-notice__form {
  display: flex;
  flex-direction: column;
  gap: var(--kore-space-md);
  max-width: var(--kore-form-max);
}
.ia-toggle {
  display: flex;
  align-items: flex-start;
  gap: 0.5rem;
  font-size: var(--kore-text-small);
}
.ia-notice__actions {
  margin-top: var(--kore-space-sm);
}
.ia-flash {
  margin-top: var(--kore-space-md);
  padding: 0.75rem 1rem;
  border-radius: var(--kore-radius-sm);
  background: var(--kore-bg-elevated);
  font-size: var(--kore-text-small);
}
.ia-flash--error {
  color: var(--kore-status-danger);
  border: 1px solid var(--kore-status-danger);
}
@media (max-width: 768px) {
  .ia-notice__actions :deep(.app-button) {
    width: 100%;
  }
}
</style>
