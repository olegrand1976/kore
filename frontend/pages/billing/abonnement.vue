<template>
  <div>
    <AppPageHeader :title="$t('billing.title')">
      <template #actions>
        <AppButton variant="secondary" size="sm" :disabled="billingLoading" @click="onOpenPortal">
          {{ $t('billing.portal') }}
        </AppButton>
        <AppButton variant="ghost" size="sm" :disabled="billingLoading" @click="onCancel">
          {{ $t('billing.cancel') }}
        </AppButton>
      </template>
    </AppPageHeader>

    <AppCard padding="lg">
      <p v-if="pending" class="muted">{{ $t('billing.loading') }}</p>
      <dl v-else-if="subscription" class="meta">
        <div><dt>{{ $t('billing.status') }}</dt><dd><AppBadge>{{ statusLabel }}</AppBadge></dd></div>
        <div><dt>{{ $t('billing.seats') }}</dt><dd>{{ seats }}</dd></div>
        <div><dt>{{ $t('billing.modules') }}</dt><dd>{{ modulesLabel }}</dd></div>
      </dl>
      <AppEmptyState v-else icon="payments" :title="$t('billing.no_subscription')" />
    </AppCard>
  </div>
</template>

<script setup lang="ts">
definePageMeta({ layout: 'default', middleware: ['admin'] })

const { loading: billingLoading, openPortal, cancelSubscription } = useBilling()
const { data, pending, refresh } = await useFetch('/api/billing/subscription')

const subscription = computed(() => (data.value as any)?.data ?? null)
const statusLabel = computed(() => String(subscription.value?.status ?? subscription.value?.Status ?? '-'))
const seats = computed(() => subscription.value?.seats ?? subscription.value?.Seats ?? '-')
const modulesLabel = computed(() => {
  const mods = subscription.value?.modules ?? subscription.value?.Modules ?? []
  return mods.map((m: any) => m.moduleCode ?? m.ModuleCode).filter(Boolean).join(', ') || '-'
})

const onOpenPortal = async () => {
  await openPortal(window.location.href)
  await refresh()
}

const onCancel = async () => {
  await cancelSubscription()
  await refresh()
}
</script>

<style scoped>
.meta { display: grid; gap: var(--kore-space-md); margin: 0; }
.meta div { display: flex; justify-content: space-between; }
.meta dt { color: var(--kore-text-muted); }
.muted { color: var(--kore-text-muted); }
</style>
