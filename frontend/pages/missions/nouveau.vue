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
        <AppInput
          v-model="form.clientId"
          :label="$t('missions.field_client_id')"
          required
        />
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
  </div>
</template>

<script setup lang="ts">
definePageMeta({ layout: 'default' })

const { t } = useI18n()
const { user } = useAuth()

const form = reactive({
  clientId: '',
  startDate: '',
  endDate: '',
  tjmAmount: 0,
  clientContact: '',
  countryCode: 'FR',
  collaboratorIds: [] as string[]
})

onMounted(() => {
  const selfId = user.value?.userId ?? user.value?.id
  if (selfId && !form.collaboratorIds.includes(selfId)) {
    form.collaboratorIds = [selfId]
  }
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
    const res = await $fetch<{ data?: { id?: string } }>('/api/ssii/missions', {
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

.flash--error {
  color: var(--kore-error);
}

@media (max-width: 768px) {
  .mission-form__actions :deep(.app-button) {
    width: 100%;
  }
}
</style>
