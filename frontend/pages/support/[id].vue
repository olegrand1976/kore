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
          <div>
            <dt>{{ $t('support.col_state') }}</dt>
            <dd><AppBadge variant="neutral">{{ stateLabel }}</AppBadge></dd>
          </div>
          <div>
            <dt>{{ $t('requests.col_application') }}</dt>
            <dd>{{ applicationLabel }}</dd>
          </div>
          <div>
            <dt>{{ $t('requests.col_priority') }}</dt>
            <dd>{{ priorityLabel }}</dd>
          </div>
          <div>
            <dt>{{ $t('requests.col_due_at') }}</dt>
            <dd>{{ dueAtLabel }}</dd>
          </div>
          <div>
            <dt>{{ $t('support.col_description') }}</dt>
            <dd>{{ description }}</dd>
          </div>
        </dl>

        <div v-if="canAssign && state !== 'resolved'" class="assign-block">
          <label for="support-assignee" class="assign-block__label">{{ $t('requests.assign_to') }}</label>
          <div class="assign-block__row">
            <select id="support-assignee" v-model="assigneeId" class="assign-block__select">
              <option value="">{{ $t('requests.assign_to') }}</option>
              <option v-for="u in users" :key="pickUserId(u)" :value="pickUserId(u)">
                {{ pickUserLogin(u) }}
              </option>
            </select>
            <AppButton
              variant="primary"
              size="sm"
              :disabled="busy || !assigneeId"
              @click="onAssign"
            >
              {{ $t('support.assign_submit') }}
            </AppButton>
          </div>
        </div>

        <div class="actions">
          <AppButton v-if="state !== 'resolved'" variant="ghost" size="sm" :disabled="busy" @click="onTakeOver">
            {{ $t('support.take_over') }}
          </AppButton>
          <AppButton v-if="state !== 'resolved'" variant="ghost" size="sm" :disabled="busy" @click="onResolve">
            {{ $t('support.resolve') }}
          </AppButton>
        </div>
      </AppCard>

      <RequestAttachmentsPanel
        resource="support"
        :resource-id="id"
        :can-upload="can('support', 'E')"
      />

      <AppCard padding="lg">
        <h2 class="section-title">{{ $t('support.reply_title') }}</h2>
        <form class="reply-form" @submit.prevent="onReply">
          <div class="reply-form__field">
            <label for="support-reply" class="reply-form__label">{{ $t('support.reply_placeholder') }}</label>
            <textarea
              id="support-reply"
              v-model="replyContent"
              class="reply-form__textarea"
              rows="4"
              required
            />
          </div>
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
const { can } = usePermissions()
const { extractFetchError } = useApiError()
const {
  get,
  assign,
  takeOver,
  resolve,
  addReply,
  pickSubject,
  pickState,
  pickDescription,
  pickPriority,
  pickDueAt,
  pickApplicationId,
  pickAssigneeId
} = useSupport()
const { list: listUsers, pickUserId, pickUserLogin } = useUsers()
const { list: listApps, pickAppLabel, appById } = useApplications()

const id = computed(() => String(route.params.id))
const pending = ref(true)
const busy = ref(false)
const errorMsg = ref('')
const ticket = ref<Awaited<ReturnType<typeof get>> | null>(null)
const replyContent = ref('')
const users = ref<Awaited<ReturnType<typeof listUsers>>>([])
const apps = ref<Awaited<ReturnType<typeof listApps>>>([])
const assigneeId = ref('')

const canAssign = computed(() => can('support', 'E'))

const pageTitle = computed(() => pickSubject(ticket.value ?? {}) || t('support.title'))
const state = computed(() => pickState(ticket.value ?? {}))
const stateLabel = computed(() => t(`support.state_${state.value}` as const, state.value))
const description = computed(() => pickDescription(ticket.value ?? {}) || '—')
const priorityLabel = computed(() => {
  const p = pickPriority(ticket.value ?? {})
  return t(`requests.priority_${p}` as const, p)
})
const dueAtLabel = computed(() => {
  const raw = pickDueAt(ticket.value ?? {})
  if (!raw) return '—'
  const d = new Date(raw)
  return Number.isNaN(d.getTime()) ? '—' : d.toLocaleString()
})
const applicationLabel = computed(() => {
  const appId = pickApplicationId(ticket.value ?? {})
  if (!appId) return '—'
  return pickAppLabel(appById(apps.value).get(appId)) || appId
})

const load = async () => {
  pending.value = true
  errorMsg.value = ''
  try {
    const [loadedTicket, userList, appList] = await Promise.all([
      get(id.value),
      listUsers(),
      listApps()
    ])
    ticket.value = loadedTicket
    users.value = userList
    apps.value = appList
    assigneeId.value = pickAssigneeId(loadedTicket) || assigneeId.value
  } catch (e) {
    errorMsg.value = extractFetchError(e)
  } finally {
    pending.value = false
  }
}

const onAssign = async () => {
  if (!assigneeId.value) return
  busy.value = true
  try {
    ticket.value = await assign(id.value, assigneeId.value)
  } catch (e) {
    errorMsg.value = extractFetchError(e)
  } finally {
    busy.value = false
  }
}

const onTakeOver = async () => {
  busy.value = true
  try {
    ticket.value = await takeOver(id.value)
    assigneeId.value = pickAssigneeId(ticket.value ?? {})
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
.meta {
  display: grid;
  gap: var(--kore-space-md);
  margin: 0 0 var(--kore-space-lg);
}
.meta dt { font-size: var(--kore-text-small); color: var(--kore-text-muted); }
.meta dd { margin: 0.25rem 0 0; }
.assign-block {
  display: grid;
  gap: var(--kore-space-sm);
  margin-bottom: var(--kore-space-lg);
}
.assign-block__label {
  font-size: var(--kore-text-small);
  color: var(--kore-text-muted);
  font-weight: 500;
}
.assign-block__row {
  display: flex;
  flex-wrap: wrap;
  gap: var(--kore-space-sm);
  align-items: center;
}
.assign-block__select {
  min-width: 12rem;
  padding: 0.5rem 0.75rem;
  border: 1px solid var(--kore-border);
  border-radius: var(--kore-radius-md);
  background: var(--kore-bg-elevated);
  color: var(--kore-text);
}
.actions { display: flex; flex-wrap: wrap; gap: var(--kore-space-sm); }
.section-title { margin: 0 0 var(--kore-space-md); font-size: var(--kore-text-h3); }
.reply-form { display: flex; flex-direction: column; gap: var(--kore-space-md); max-width: var(--kore-form-wide-max); }
.reply-form__field { display: grid; gap: var(--kore-space-xs); }
.reply-form__label { font-size: var(--kore-text-small); color: var(--kore-text-muted); font-weight: 500; }
.reply-form__textarea {
  padding: 0.75rem 1rem;
  font-family: var(--kore-font);
  font-size: var(--kore-text-body);
  color: var(--kore-text);
  background: var(--kore-bg-elevated);
  border: 1px solid var(--kore-border);
  border-radius: var(--kore-radius-md);
  resize: vertical;
  min-height: 6rem;
}
.muted { color: var(--kore-text-muted); }
.flash { margin-bottom: var(--kore-space-md); font-size: var(--kore-text-small); }
.flash--error { color: var(--kore-status-danger); }
@media (max-width: 768px) {
  .actions :deep(.app-button),
  .assign-block__row :deep(.app-button),
  .reply-form :deep(.app-button) { width: 100%; }
  .assign-block__select { width: 100%; }
  .assign-block__row { flex-direction: column; align-items: stretch; }
}
</style>
