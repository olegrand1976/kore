<script setup lang="ts">
const props = defineProps<{
  demandId: string
}>()

const { apiFetch } = useApiFetch()
const { t } = useI18n()

type TaigaLink = {
  externalRef?: number | null
  ExternalRef?: number | null
  externalUrl?: string
  ExternalURL?: string
  lastSyncAt?: string
  LastSyncAt?: string
}

const link = ref<TaigaLink | null>(null)
const loaded = ref(false)

const externalRef = computed(() => link.value?.externalRef ?? link.value?.ExternalRef ?? null)
const externalUrl = computed(() => {
  const raw = link.value?.externalUrl ?? link.value?.ExternalURL ?? ''
  return typeof raw === 'string' ? raw.trim() : ''
})
const lastSyncAt = computed(() => link.value?.lastSyncAt ?? link.value?.LastSyncAt ?? '')

const lastSyncLabel = computed(() => {
  if (!lastSyncAt.value) return ''
  const d = new Date(lastSyncAt.value)
  if (Number.isNaN(d.getTime())) return ''
  return t('tma.taiga_last_sync', { date: d.toLocaleString() })
})

onMounted(async () => {
  try {
    const res = await apiFetch<{ data?: TaigaLink }>(
      `/api/integrations/taiga/links/by-demand/${props.demandId}`
    )
    link.value = res?.data ?? null
  } catch {
    link.value = null
  } finally {
    loaded.value = true
  }
})
</script>

<template>
  <AppCard v-if="loaded" padding="lg" class="taiga-panel mb">
    <h2 class="taiga-panel__title">{{ $t('tma.taiga_title') }}</h2>
    <template v-if="externalRef != null">
      <p class="taiga-panel__ref">{{ $t('tma.taiga_ref', { ref: externalRef }) }}</p>
      <p v-if="lastSyncLabel" class="taiga-panel__sync muted">{{ lastSyncLabel }}</p>
      <a
        v-if="externalUrl"
        class="taiga-panel__link"
        :href="externalUrl"
        target="_blank"
        rel="noopener noreferrer"
      >
        {{ $t('tma.taiga_open') }}
      </a>
    </template>
    <p v-else class="muted">{{ $t('tma.taiga_not_linked') }}</p>
  </AppCard>
</template>

<style scoped>
.taiga-panel__title {
  margin: 0 0 var(--kore-space-md);
  font-size: var(--kore-text-body);
}
.taiga-panel__ref {
  margin: 0 0 var(--kore-space-sm);
}
.taiga-panel__sync {
  margin: 0 0 var(--kore-space-sm);
  font-size: var(--kore-text-small);
}
.taiga-panel__link {
  display: inline-flex;
  align-items: center;
  color: var(--kore-accent);
  font-size: var(--kore-text-small);
  text-decoration: none;
  width: 100%;
  justify-content: center;
  padding: var(--kore-space-sm) var(--kore-space-md);
  border: 1px solid var(--kore-border);
  border-radius: var(--kore-radius-md);
}
.taiga-panel__link:hover {
  background: color-mix(in srgb, var(--kore-accent) 8%, transparent);
}
.muted {
  color: var(--kore-text-muted);
}
.mb {
  margin-bottom: var(--kore-space-lg);
}
@media (min-width: 641px) {
  .taiga-panel__link {
    width: auto;
    justify-content: flex-start;
  }
}
</style>
