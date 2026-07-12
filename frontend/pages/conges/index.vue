<template>
  <div>
    <AppPageHeader :title="$t('conges.title')">
      <template #actions>
        <AppButton variant="primary" size="sm" @click="showForm = !showForm">
          {{ $t('conges.new_request') }}
        </AppButton>
      </template>
    </AppPageHeader>

    <AppCard v-if="showForm" padding="lg" class="mb">
      <form class="form" @submit.prevent="submitRequest">
        <AppInput id="from" v-model="form.from" type="date" :label="$t('conges.from')" />
        <AppInput id="to" v-model="form.to" type="date" :label="$t('conges.to')" />
        <AppInput id="motif" v-model="form.motif" :label="$t('conges.motif')" />
        <AppButton variant="primary" size="sm" type="submit" :disabled="submitting">
          {{ $t('conges.submit') }}
        </AppButton>
      </form>
    </AppCard>

    <AppCard padding="lg">
      <AppTable
        :columns="columns"
        :rows="rows"
        :loading="pending"
        :empty-title="$t('conges.empty')"
      />
    </AppCard>
  </div>
</template>

<script setup lang="ts">
definePageMeta({ layout: 'default' })

const { t } = useI18n()
const { data, pending, refresh } = await useFetch('/api/conges/leave-requests')
const showForm = ref(false)
const submitting = ref(false)
const form = reactive({ from: '', to: '', motif: '' })

const items = computed(() => (data.value as any)?.data ?? [])

const columns = computed(() => [
  { key: 'type', label: t('conges.col_type') },
  { key: 'from', label: t('conges.from') },
  { key: 'to', label: t('conges.to') },
  { key: 'status', label: t('conges.col_status') }
])

const rows = computed(() =>
  items.value.map((item: any) => ({
    type: item.type ?? item.Type,
    from: String(item.from ?? item.From ?? '').slice(0, 10),
    to: String(item.to ?? item.To ?? '').slice(0, 10),
    status: item.status ?? item.Status
  }))
)

const submitRequest = async () => {
  submitting.value = true
  try {
    await $fetch('/api/conges/leave-requests', {
      method: 'POST',
      body: { type: 'CP', from: form.from, to: form.to, motif: form.motif }
    })
    showForm.value = false
    await refresh()
  } finally {
    submitting.value = false
  }
}
</script>

<style scoped>
.mb { margin-bottom: var(--kore-space-lg); }
.form { display: grid; gap: var(--kore-space-md); max-width: 420px; }
</style>
