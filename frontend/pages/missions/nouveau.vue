<template>
  <div>
    <AppPageHeader :title="$t('missions.new_title')" :subtitle="$t('missions.new_subtitle')">
      <template #actions>
        <AppButton variant="ghost" size="sm" to="/missions">
          <AppIcon name="arrow_back" /> {{ $t('missions.back_list') }}
        </AppButton>
      </template>
    </AppPageHeader>

    <AppCard padding="lg">
      <form class="mission-form" @submit.prevent="submit">
        <div class="mission-form__field">
          <label for="mission-client">{{ $t('missions.field_client') }}</label>
          <div class="mission-form__client">
            <select id="mission-client" v-model="form.clientId" required>
              <option value="">{{ $t('missions.client_placeholder') }}</option>
              <option v-for="c in clients" :key="c.id" :value="c.id">{{ c.label }}</option>
            </select>
            <AppButton variant="ghost" size="sm" type="button" @click="openClientModal">
              <AppIcon name="add" /> {{ $t('missions.new_client') }}
            </AppButton>
          </div>
          <p v-if="!clients.length" class="mission-form__hint">
            {{ $t('missions.no_client_hint') }}
          </p>
        </div>
        <AppInput
          v-model="form.startDate"
          type="date"
          :label="$t('fiche.col_start')"
          required
        />
        <AppInput
          v-model="form.endDate"
          type="date"
          :label="$t('fiche.col_end')"
        />
        <AppInput
          v-model.number="form.tjmAmount"
          type="number"
          min="0"
          step="100"
          :label="$t('fiche.col_tjm')"
          required
        />
        <AppInput
          v-model="form.clientContact"
          :label="$t('missions.field_contact')"
        />
        <AppInput
          v-model="form.countryCode"
          :label="$t('missions.field_country')"
          maxlength="2"
        />
        <AppUserMultiSelect
          id="mission-collaborators"
          v-model="form.collaboratorIds"
          :label="$t('missions.field_collaborators')"
          required
        />
        <p v-if="errorMsg" class="flash flash--error" role="alert">{{ errorMsg }}</p>
        <div class="mission-form__actions">
          <AppButton variant="primary" type="submit" :loading="submitting">
            {{ $t('missions.create') }}
          </AppButton>
        </div>
      </form>
    </AppCard>

    <AppModal v-model:open="clientModalOpen" width="sm" :aria-label="$t('missions.new_client')">
      <form class="mission-client-form" @submit.prevent="submitClient">
        <h2 class="mission-client-form__title">{{ $t('missions.new_client') }}</h2>
        <AppInput
          id="new-client-name"
          v-model="clientForm.raisonSociale"
          :label="$t('missions.client_name')"
          required
        />
        <AppInput id="new-client-tva" v-model="clientForm.tva" :label="$t('missions.client_tva')" />
        <p v-if="clientError" class="flash flash--error" role="alert">{{ clientError }}</p>
        <div class="mission-client-form__actions">
          <AppButton variant="ghost" type="button" @click="clientModalOpen = false">
            {{ $t('common.cancel') }}
          </AppButton>
          <AppButton variant="primary" type="submit" :disabled="creatingClient">
            {{ $t('common.save') }}
          </AppButton>
        </div>
      </form>
    </AppModal>
  </div>
</template>

<script setup lang="ts">
definePageMeta({ layout: 'default' })

const { apiFetch } = useApiFetch()
const { t } = useI18n()
const { user } = useAuth()
const { listClients, createClient, orgId } = useOrganisation()
const { extractFetchError } = useApiError()

const form = reactive({
  clientId: '',
  startDate: '',
  endDate: '',
  tjmAmount: 0,
  clientContact: '',
  countryCode: 'FR',
  collaboratorIds: [] as string[]
})

const clients = ref<{ id: string; label: string }[]>([])
const clientModalOpen = ref(false)
const creatingClient = ref(false)
const clientError = ref('')
const clientForm = reactive({ raisonSociale: '', tva: '' })

const loadClients = async () => {
  try {
    const items = await listClients()
    clients.value = items.map((c) => ({
      id: orgId(c),
      label: c.raisonSociale ?? c.RaisonSociale ?? orgId(c)
    }))
  } catch {
    clients.value = []
  }
}

const openClientModal = () => {
  clientForm.raisonSociale = ''
  clientForm.tva = ''
  clientError.value = ''
  clientModalOpen.value = true
}

const submitClient = async () => {
  creatingClient.value = true
  clientError.value = ''
  try {
    const res = await createClient({
      raisonSociale: clientForm.raisonSociale,
      tva: clientForm.tva || undefined
    })
    await loadClients()
    // Présélectionne le client fraîchement créé pour éviter un second aller-retour.
    const created = (res as { data?: { id?: string; ID?: string } })?.data
    const createdId = created?.id ?? created?.ID ?? ''
    if (createdId) form.clientId = createdId
    clientModalOpen.value = false
  } catch (err) {
    clientError.value = extractFetchError(err, t('missions.client_create_error'))
  } finally {
    creatingClient.value = false
  }
}

onMounted(async () => {
  const selfId = user.value?.userId ?? user.value?.id
  if (selfId && !form.collaboratorIds.includes(selfId)) {
    form.collaboratorIds = [selfId]
  }
  await loadClients()
})

const submitting = ref(false)
const errorMsg = ref('')

async function submit() {
  errorMsg.value = ''
  if (!form.collaboratorIds.length) {
    errorMsg.value = t('missions.collaborators_required')
    return
  }
  submitting.value = true
  try {
    const body: Record<string, unknown> = {
      clientId: form.clientId,
      startDate: new Date(form.startDate).toISOString(),
      tjmAmount: Math.round(form.tjmAmount),
      currency: 'EUR',
      clientContact: form.clientContact,
      countryCode: form.countryCode || 'FR',
      technologies: [],
      collaboratorIds: form.collaboratorIds
    }
    if (form.endDate) {
      body.endDate = new Date(form.endDate).toISOString()
    }
    const res = await apiFetch<{ data?: { id?: string } }>('/api/ssii/missions', {
      method: 'POST',
      body
    })
    const id = res?.data?.id
    if (id) {
      await navigateTo(`/missions/${id}`)
      return
    }
    await navigateTo('/missions')
  } catch {
    errorMsg.value = t('missions.create_error')
  } finally {
    submitting.value = false
  }
}
</script>

<style scoped>
.mission-form {
  display: grid;
  gap: var(--kore-space-md);
  max-width: var(--kore-form-max);
}

.mission-form__actions {
  display: flex;
  gap: var(--kore-space-sm);
}

.mission-form__field {
  display: flex;
  flex-direction: column;
  gap: var(--kore-space-xs);
}

.mission-form__field label {
  font-size: var(--kore-text-small);
  font-weight: 500;
}

.mission-form__client {
  display: flex;
  align-items: center;
  gap: var(--kore-space-sm);
}

.mission-form__client select {
  flex: 1;
  padding: 0.625rem 0.75rem;
  border: 1px solid var(--kore-border);
  border-radius: var(--kore-radius-md);
  background: var(--kore-bg);
  color: var(--kore-text);
  font-size: var(--kore-text-small);
}

.mission-form__hint {
  margin: 0;
  font-size: var(--kore-text-caption);
  color: var(--kore-text-muted);
}

.mission-client-form {
  display: flex;
  flex-direction: column;
  gap: var(--kore-space-lg);
}

.mission-client-form__title {
  margin: 0;
  font-size: var(--kore-text-h3);
}

.mission-client-form__actions {
  display: flex;
  justify-content: flex-end;
  gap: var(--kore-space-sm);
}

.flash--error {
  color: var(--kore-error);
}

@media (max-width: 768px) {
  .mission-form__actions :deep(.app-button) {
    width: 100%;
  }

  .mission-form__client {
    flex-direction: column;
    align-items: stretch;
  }

  .mission-client-form__actions {
    flex-direction: column-reverse;
  }

  .mission-client-form__actions :deep(.app-button) {
    width: 100%;
  }
}
</style>
