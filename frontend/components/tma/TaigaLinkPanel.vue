<script setup lang="ts">
const props = defineProps<{
  demandId: string
  canPush?: boolean
}>()

const { apiFetch } = useApiFetch()
const { extractFetchError } = useApiError()
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
const pushing = ref(false)
const pushError = ref('')

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

const hasLink = computed(() => externalRef.value != null || externalUrl.value !== '')

let loadGeneration = 0

async function loadLink(demandId: string) {
  const generation = ++loadGeneration
  loaded.value = false
  link.value = null
  pushError.value = ''
  try {
    const res = await apiFetch<{ data?: TaigaLink }>(
      `/api/integrations/taiga/links/by-demand/${demandId}`
    )
    if (generation !== loadGeneration) return
    link.value = res?.data ?? null
  } catch {
    if (generation !== loadGeneration) return
    link.value = null
  } finally {
    if (generation === loadGeneration) {
      loaded.value = true
    }
  }
}

async function pushToTaiga() {
  if (!props.demandId || pushing.value) return
  pushing.value = true
  pushError.value = ''
  try {
    const res = await apiFetch<{ data?: TaigaLink }>(
      `/api/integrations/taiga/demands/${props.demandId}/push`,
      { method: 'POST' }
    )
    link.value = res?.data ?? res ?? null
  } catch (e) {
    pushError.value = extractFetchError(e) || t('tma.taiga_push_error')
  } finally {
    pushing.value = false
  }
}

watch(
  () => props.demandId,
  (demandId) => {
    if (demandId) {
      void loadLink(demandId)
    }
  },
  { immediate: true }
)
</script>

<template>
  <AppCard v-if="loaded" padding="lg" class="taiga-panel mb">
    <h2 class="taiga-panel__title">{{ $t('tma.taiga_title') }}</h2>
    <template v-if="hasLink">
      <p v-if="externalRef != null" class="taiga-panel__ref">
        {{ $t('tma.taiga_ref', { ref: externalRef }) }}
      </p>
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
    <template v-else>
      <p class="muted">{{ $t('tma.taiga_not_linked') }}</p>
      <AppButton
        v-if="canPush"
        variant="secondary"
        size="sm"
        type="button"
        class="taiga-panel__push"
        :disabled="pushing"
        @click="pushToTaiga"
      >
        {{ pushing ? $t('common.loading') : $t('tma.taiga_push') }}
      </AppButton>
      <p v-if="pushError" class="taiga-panel__error" role="alert">{{ pushError }}</p>
    </template>
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
.taiga-panel__push {
  margin-top: var(--kore-space-sm);
  width: 100%;
}
.taiga-panel__error {
  margin: var(--kore-space-sm) 0 0;
  color: var(--kore-error);
  font-size: var(--kore-text-small);
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
  .taiga-panel__push {
    width: auto;
  }
}
</style>
