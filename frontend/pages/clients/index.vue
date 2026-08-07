<template>
  <div>
    <AppPageHeader :title="$t('clients.title')" :subtitle="$t('clients.subtitle')">
      <template #actions>
        <AppButton
          v-if="canCreate"
          variant="primary"
          size="sm"
          type="button"
          class="clients-new-btn"
          @click="openCreateModal"
        >
          <AppIcon name="add" /> {{ $t('clients.new') }}
        </AppButton>
      </template>
    </AppPageHeader>

    <AppCard v-if="pending" padding="lg">
      <p class="muted">{{ $t('fiche.loading') }}</p>
    </AppCard>

    <AppCard v-else-if="loadError" padding="lg">
      <p class="flash flash--error">{{ $t('clients.load_error') }}</p>
    </AppCard>

    <AppCard v-else padding="none">
      <AppTable
        :columns="columns"
        :rows="rows"
        :empty-title="$t('clients.empty')"
        :empty-description="canCreate ? $t('clients.empty_desc_create') : $t('clients.empty_desc')"
      >
        <template #cell-actions="{ row }">
          <AppButton variant="ghost" size="sm" :to="`/clients/${row.id}`">
            {{ $t('clients.open') }}
          </AppButton>
        </template>
      </AppTable>
    </AppCard>

    <AppModal v-model:open="createOpen" width="md" :aria-label="$t('clients.new')">
      <form class="clients-form" @submit.prevent="submitCreate">
        <h2 class="clients-form__title">{{ $t('clients.new') }}</h2>
        <p class="clients-form__hint">{{ $t('clients.billing_hint') }}</p>
        <ClientBillingForm v-model="createForm" id-prefix="client-create" />
        <p v-if="createError" class="flash flash--error" role="alert">{{ createError }}</p>
        <div class="clients-form__actions">
          <AppButton variant="ghost" type="button" @click="createOpen = false">
            {{ $t('common.cancel') }}
          </AppButton>
          <AppButton variant="primary" type="submit" :loading="creating">
            {{ $t('common.save') }}
          </AppButton>
        </div>
      </form>
    </AppModal>
  </div>
</template>

<script setup lang="ts">
import {
  clientBillingPayload,
  emptyClientBillingFields,
  type ClientBillingFields
} from '~/composables/useBillingCountry'

definePageMeta({ layout: 'default' })

const { t } = useI18n()
const { can } = usePermissions()
const { listClients, createClient, orgId } = useOrganisation()
const { extractFetchError } = useApiError()

const canCreate = computed(() => can('org', 'E'))

type ClientRow = {
  id: string
  name: string
  tva: string
}

const rows = ref<ClientRow[]>([])
const pending = ref(true)
const loadError = ref(false)

const columns = computed(() => [
  { key: 'name', label: t('clients.col_name') },
  { key: 'tva', label: t('clients.col_tva') },
  { key: 'actions', label: t('prestations.col_actions'), nowrap: true }
])

const load = async () => {
  pending.value = true
  loadError.value = false
  try {
    const items = await listClients()
    rows.value = items.map((c) => ({
      id: orgId(c),
      name: c.raisonSociale ?? c.RaisonSociale ?? '—',
      tva: c.tva || '—'
    }))
  } catch {
    loadError.value = true
    rows.value = []
  } finally {
    pending.value = false
  }
}

await load()

const createOpen = ref(false)
const creating = ref(false)
const createError = ref('')
const createForm = reactive<ClientBillingFields>(emptyClientBillingFields())

const openCreateModal = () => {
  Object.assign(createForm, emptyClientBillingFields())
  createError.value = ''
  createOpen.value = true
}

const submitCreate = async () => {
  creating.value = true
  createError.value = ''
  const payload = clientBillingPayload(createForm)
  if (!payload.raisonSociale) {
    createError.value = t('clients.create_error')
    creating.value = false
    return
  }
  try {
    const res = await createClient(payload)
    const created = (res as { data?: { id?: string; ID?: string } })?.data
    const createdId = created?.id ?? created?.ID ?? ''
    createOpen.value = false
    if (createdId) {
      await navigateTo(`/clients/${createdId}`)
      return
    }
    await load()
  } catch (err) {
    createError.value = extractFetchError(err, t('clients.create_error'))
  } finally {
    creating.value = false
  }
}
</script>

<style scoped>
.muted {
  margin: 0;
  color: var(--kore-text-muted);
}

.flash--error {
  color: var(--kore-error);
}

.clients-form {
  display: grid;
  gap: var(--kore-space-md);
}

.clients-form__title {
  margin: 0;
  font-size: var(--kore-text-h3);
  color: var(--kore-text);
}

.clients-form__hint {
  margin: 0;
  font-size: var(--kore-text-small);
  color: var(--kore-text-muted);
}

.clients-form__actions {
  display: flex;
  flex-wrap: wrap;
  gap: var(--kore-space-sm);
  justify-content: flex-end;
}

@media (max-width: 768px) {
  .clients-new-btn {
    width: 100%;
  }

  .clients-form__actions {
    flex-direction: column-reverse;
  }

  .clients-form__actions :deep(.app-button) {
    width: 100%;
  }
}
</style>
