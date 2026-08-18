<script setup lang="ts">
import {
  timesheetAdminAction,
  timesheetAdminConfirmKey,
  type AdminTimesheetAction
} from '~/utils/craTimesheetAdmin'

const props = withDefaults(defineProps<{
  timesheetId: string
  status: string
  userLabel?: string
  month: string
  disabled?: boolean
}>(), {
  userLabel: '',
  disabled: false
})

const emit = defineEmits<{
  changed: [action: AdminTimesheetAction]
  error: [message: string]
}>()

const { t, locale } = useI18n()
const { isAdmin } = useAuth()
const { mapCraError } = useCraError()
const { apiFetch } = useApiFetch()

const open = ref(false)
const busy = ref(false)
const titleId = `cra-admin-${props.timesheetId}`

const action = computed(() => timesheetAdminAction(props.status))

const monthLabel = computed(() => {
  const [y, m] = props.month.split('-').map(Number)
  if (!y || !m) return props.month
  return new Date(y, m - 1, 1).toLocaleDateString(locale.value === 'en' ? 'en-US' : 'fr-FR', {
    month: 'long',
    year: 'numeric'
  })
})

const confirmLabel = computed(() => {
  const user = props.userLabel.trim()
  const key = timesheetAdminConfirmKey(action.value, Boolean(user))
  return t(key, { user, month: monthLabel.value })
})

const submitLabel = computed(() =>
  action.value === 'unvalidate' ? t('cra.unvalidate') : t('common.delete')
)

const confirm = async () => {
  busy.value = true
  const current = action.value
  try {
    switch (current) {
      case 'unvalidate':
        await apiFetch(`/api/cra/timesheets/${props.timesheetId}/unvalidate`, { method: 'POST' })
        break
      case 'delete':
        await apiFetch(`/api/cra/timesheets/${props.timesheetId}`, { method: 'DELETE' })
        break
      default: {
        const _exhaustive: never = current
        return _exhaustive
      }
    }
    open.value = false
    emit('changed', current)
  } catch (err) {
    const fallback = current === 'unvalidate' ? t('cra.unvalidate_error') : t('cra.delete_error')
    emit('error', mapCraError(err, fallback))
  } finally {
    busy.value = false
  }
}
</script>

<template>
  <template v-if="isAdmin">
    <AppButton
      :variant="action === 'delete' ? 'danger' : 'secondary'"
      size="sm"
      :disabled="disabled || busy"
      @click="open = true"
    >
      {{ submitLabel }}
    </AppButton>
    <AppModal v-model:open="open" width="md" :title-id="titleId" :aria-label="submitLabel">
      <form class="cra-delete" @submit.prevent="confirm">
        <h2 :id="titleId" class="cra-delete__title">{{ submitLabel }}</h2>
        <p>{{ confirmLabel }}</p>
        <p class="cra-delete__hint">{{ $t('cra.admin_invoice_warning') }}</p>
        <div class="cra-delete__actions">
          <AppButton variant="ghost" size="sm" type="button" @click="open = false">
            {{ $t('common.cancel') }}
          </AppButton>
          <AppButton
            :variant="action === 'delete' ? 'danger' : 'primary'"
            size="sm"
            type="submit"
            :disabled="busy"
          >
            {{ submitLabel }}
          </AppButton>
        </div>
      </form>
    </AppModal>
  </template>
</template>

<style scoped>
.cra-delete {
  display: grid;
  gap: var(--kore-space-md);
}

.cra-delete__title {
  margin: 0;
  font-size: var(--kore-text-h3);
  color: var(--kore-text);
}

.cra-delete p {
  margin: 0;
  color: var(--kore-text);
}

.cra-delete p.cra-delete__hint {
  color: var(--kore-text-muted);
  font-size: var(--kore-text-small);
}

.cra-delete__actions {
  display: flex;
  justify-content: flex-end;
  gap: var(--kore-space-sm);
}

@media (max-width: 768px) {
  .cra-delete__actions {
    flex-direction: column;
  }

  .cra-delete__actions :deep(.app-btn) {
    width: 100%;
  }
}
</style>
