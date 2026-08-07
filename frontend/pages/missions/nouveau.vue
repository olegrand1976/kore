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
          v-model="form.title"
          :label="$t('missions.field_title')"
          :placeholder="$t('missions.field_title_placeholder')"
        />
        <div class="mission-form__field">
          <label for="mission-rate-unit">{{ $t('missions.field_rate_unit') }}</label>
          <select id="mission-rate-unit" v-model="form.rateUnit" class="mission-form__select">
            <option value="tjm">{{ $t('missions.rate_unit_tjm') }}</option>
            <option value="hourly">{{ $t('missions.rate_unit_hourly') }}</option>
          </select>
        </div>
        <AppInput
          v-model.number="form.amountEuros"
          type="number"
          min="0"
          :step="form.rateUnit === 'hourly' ? '0.01' : '1'"
          :label="rateAmountLabel"
          required
        />
        <p class="mission-form__hint">{{ $t('missions.rate_billing_hint') }}</p>
        <div class="mission-form__field">
          <ClientContactMultiSelect
            id="mission-contacts"
            v-model="form.clientContactIds"
            :contacts="selectedClientContacts"
            :label="$t('missions.field_contact')"
          />
          <p class="mission-form__hint">{{ $t('missions.contact_select_hint') }}</p>
          <AppButton
            variant="ghost"
            size="sm"
            type="button"
            :disabled="!form.clientId"
            @click="openContactModal"
          >
            <AppIcon name="add" /> {{ $t('missions.contact_add_inline') }}
          </AppButton>
        </div>
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

    <AppModal v-model:open="contactModalOpen" width="sm" :aria-label="$t('missions.contact_add_inline')">
      <form class="mission-client-form" @submit.prevent="submitContact">
        <h2 class="mission-client-form__title">{{ $t('missions.contact_add_inline') }}</h2>
        <AppInput id="mission-contact-prenom" v-model="contactForm.prenom" :label="$t('clients.contact_prenom')" />
        <AppInput id="mission-contact-nom" v-model="contactForm.nom" :label="$t('clients.contact_nom')" />
        <AppInput id="mission-contact-role" v-model="contactForm.role" :label="$t('clients.contact_role')" />
        <AppInput id="mission-contact-email" v-model="contactForm.email" type="email" :label="$t('clients.contact_email')" />
        <AppInput id="mission-contact-phone" v-model="contactForm.telephone" :label="$t('clients.contact_phone')" />
        <p v-if="contactError" class="flash flash--error" role="alert">{{ contactError }}</p>
        <div class="mission-client-form__actions">
          <AppButton variant="ghost" type="button" @click="contactModalOpen = false">
            {{ $t('common.cancel') }}
          </AppButton>
          <AppButton variant="primary" type="submit" :loading="creatingContact">
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
const { listClients, createClient, replaceClientContacts, orgId } = useOrganisation()
const { extractFetchError } = useApiError()

type OrgClientContact = {
  id?: string
  nom?: string
  prenom?: string
  email?: string
  role?: string
  telephone?: string
}

type OrgClient = {
  id?: string
  ID?: string
  raisonSociale?: string
  RaisonSociale?: string
  contacts?: OrgClientContact[]
}

const form = reactive({
  clientId: '',
  startDate: '',
  endDate: '',
  title: '',
  rateUnit: 'tjm' as 'tjm' | 'hourly',
  amountEuros: 0,
  clientContactIds: [] as string[],
  collaboratorIds: [] as string[]
})

const rateAmountLabel = computed(() => {
  switch (form.rateUnit) {
    case 'hourly':
      return t('missions.field_hourly_amount')
    case 'tjm':
      return t('missions.field_tjm_amount')
    default: {
      const _exhaustive: never = form.rateUnit
      return _exhaustive
    }
  }
})

const clients = ref<{ id: string; label: string; contacts: OrgClientContact[] }[]>([])
const clientModalOpen = ref(false)
const creatingClient = ref(false)
const clientError = ref('')
const clientForm = reactive({ raisonSociale: '', tva: '' })
const contactModalOpen = ref(false)
const creatingContact = ref(false)
const contactError = ref('')
const contactForm = reactive({
  prenom: '',
  nom: '',
  role: '',
  email: '',
  telephone: ''
})

const selectedClientContacts = computed(() => {
  const client = clients.value.find((c) => c.id === form.clientId)
  return (client?.contacts ?? [])
    .filter((c): c is OrgClientContact & { id: string } => Boolean(c.id))
    .map((c) => ({
      id: c.id,
      nom: c.nom,
      prenom: c.prenom,
      email: c.email,
      role: c.role,
      telephone: c.telephone
    }))
})

watch(() => form.clientId, () => {
  form.clientContactIds = []
})

const loadClients = async () => {
  try {
    const items = await listClients() as OrgClient[]
    clients.value = items.map((c) => ({
      id: orgId(c),
      label: c.raisonSociale ?? c.RaisonSociale ?? orgId(c),
      contacts: c.contacts ?? []
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

const openContactModal = () => {
  contactForm.prenom = ''
  contactForm.nom = ''
  contactForm.role = ''
  contactForm.email = ''
  contactForm.telephone = ''
  contactError.value = ''
  contactModalOpen.value = true
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

const submitContact = async () => {
  if (!form.clientId) return
  creatingContact.value = true
  contactError.value = ''
  try {
    const client = clients.value.find((c) => c.id === form.clientId)
    const existing = (client?.contacts ?? []).map((c) => ({
      ...(c.id ? { id: c.id } : {}),
      nom: c.nom ?? '',
      prenom: c.prenom ?? '',
      email: c.email ?? '',
      role: c.role ?? '',
      telephone: c.telephone ?? ''
    }))
    const beforeIds = new Set(
      (client?.contacts ?? []).map((c) => c.id).filter((id): id is string => Boolean(id))
    )
    const res = await replaceClientContacts(form.clientId, [
      ...existing,
      {
        prenom: contactForm.prenom.trim(),
        nom: contactForm.nom.trim(),
        role: contactForm.role.trim(),
        email: contactForm.email.trim(),
        telephone: contactForm.telephone.trim()
      }
    ])
    await loadClients()
    const updated = (res as { data?: OrgClient })?.data
    const contacts = updated?.contacts ?? clients.value.find((c) => c.id === form.clientId)?.contacts ?? []
    const created = contacts.find((c) => c.id && !beforeIds.has(c.id))
    if (created?.id && !form.clientContactIds.includes(created.id)) {
      form.clientContactIds = [...form.clientContactIds, created.id]
    }
    contactModalOpen.value = false
  } catch (err) {
    contactError.value = extractFetchError(err, t('clients.contact_save_error'))
  } finally {
    creatingContact.value = false
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
      title: form.title.trim(),
      rateUnit: form.rateUnit,
      tjmAmount: Math.round(Number(form.amountEuros) * 100),
      currency: 'EUR',
      clientContactIds: form.clientContactIds,
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

.mission-form__client select,
.mission-form__select {
  flex: 1;
  padding: 0.625rem 0.75rem;
  border: 1px solid var(--kore-border);
  border-radius: var(--kore-radius-md);
  background: var(--kore-bg);
  color: var(--kore-text);
  font-size: var(--kore-text-small);
}

.mission-form__select {
  width: 100%;
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
