<template>
  <div>
    <AppPageHeader :title="pageTitle">
      <template #actions>
        <AppButton variant="ghost" size="sm" @click="navigateTo('/support')">
          {{ $t('support.back') }}
        </AppButton>
      </template>
    </AppPageHeader>

    <p v-if="errorMsg" class="flash flash--error" role="alert">{{ errorMsg }}</p>
    <AppCard v-if="pending" padding="lg"><p class="muted">{{ $t('support.loading') }}</p></AppCard>

    <template v-else-if="ticket">
      <AppCard padding="lg" class="mb">
        <dl class="meta">
          <div><dt>{{ $t('support.col_state') }}</dt><dd><AppBadge variant="neutral">{{ state }}</AppBadge></dd></div>
          <div><dt>{{ $t('support.col_description') }}</dt><dd>{{ description }}</dd></div>
        </dl>
        <div class="actions">
          <AppButton v-if="state !== 'resolved'" variant="primary" size="sm" :disabled="busy" @click="onTakeOver">
            {{ $t('support.take_over') }}
          </AppButton>
          <AppButton v-if="state !== 'resolved'" variant="ghost" size="sm" :disabled="busy" @click="onResolve">
            {{ $t('support.resolve') }}
          </AppButton>
        </div>
      </AppCard>

      <AppCard padding="lg">
        <h2 class="section-title">{{ $t('support.reply_title') }}</h2>
        <form class="reply-form" @submit.prevent="onReply">
          <AppInput v-model="replyContent" :label="$t('support.reply_placeholder')" multiline required />
          <AppButton variant="primary" size="sm" type="submit" :disabled="busy || !replyContent.trim()">
            {{ $t('support.reply_send') }}
          </AppButton>
        </form>
      </AppCard>
    </template>
  </div>
</template>

<script setup lang="ts">
definePageMeta({ layout: 'default' })

const route = useRoute()
const { t } = useI18n()
const { extractFetchError } = useApiError()
const { get, takeOver, resolve, addReply, pickSubject, pickState, pickDescription } = useSupport()

const id = computed(() => String(route.params.id))
const pending = ref(true)
const busy = ref(false)
const errorMsg = ref('')
const ticket = ref<Awaited<ReturnType<typeof get>> | null>(null)
const replyContent = ref('')

const pageTitle = computed(() => pickSubject(ticket.value ?? {}) || t('support.title'))
const state = computed(() => pickState(ticket.value ?? {}))
const description = computed(() => pickDescription(ticket.value ?? {}))

const load = async () => {
  pending.value = true
  errorMsg.value = ''
  try {
    ticket.value = await get(id.value)
  } catch (e) {
    errorMsg.value = extractFetchError(e)
  } finally {
    pending.value = false
  }
}

const onTakeOver = async () => {
  busy.value = true
  try {
    ticket.value = await takeOver(id.value)
  } catch (e) {
    errorMsg.value = extractFetchError(e)
  } finally {
    busy.value = false
  }
}

const onResolve = async () => {
  busy.value = true
  try {
    ticket.value = await resolve(id.value)
  } catch (e) {
    errorMsg.value = extractFetchError(e)
  } finally {
    busy.value = false
  }
}

const onReply = async () => {
  busy.value = true
  try {
    await addReply(id.value, replyContent.value.trim())
    replyContent.value = ''
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
.meta { display: grid; gap: var(--kore-space-md); margin: 0 0 var(--kore-space-lg); }
.meta dt { font-size: var(--kore-text-small); color: var(--kore-text-muted); }
.meta dd { margin: 0.25rem 0 0; }
.actions { display: flex; flex-wrap: wrap; gap: var(--kore-space-sm); }
.section-title { margin: 0 0 var(--kore-space-md); font-size: var(--kore-text-h3); }
.reply-form { display: flex; flex-direction: column; gap: var(--kore-space-md); max-width: var(--kore-form-wide-max); }
.muted { color: var(--kore-text-muted); }
.flash { margin-bottom: var(--kore-space-md); font-size: var(--kore-text-small); }
.flash--error { color: var(--kore-status-danger); }
@media (max-width: 768px) {
  .actions :deep(.app-button) { flex: 1 1 100%; }
  .reply-form :deep(.app-button) { width: 100%; }
}
</style>
