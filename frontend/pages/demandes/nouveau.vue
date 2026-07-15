<script setup lang="ts">
import type { RequestChannel } from '~/composables/useRequestSettings'

definePageMeta({ layout: 'default' })

const { t } = useI18n()
const route = useRoute()
const { fetchSettings, isChannelEnabled, activeChannelCount, settings, loaded } = useRequestSettings()
const { hasModule } = useEntitlements()
const { can } = usePermissions()

type TriageCard = {
  channel: RequestChannel
  to: string
  icon: string
  titleKey: string
  descKey: string
}

const cards = computed<TriageCard[]>(() => {
  const out: TriageCard[] = []
  if (isChannelEnabled('tma') && hasModule('tma') && can('tma', 'E')) {
    out.push({
      channel: 'tma',
      to: '/tma?create=1',
      icon: 'support_agent',
      titleKey: 'demandes.triage_tma_title',
      descKey: 'demandes.triage_tma_desc'
    })
  }
  if (isChannelEnabled('support') && can('support', 'E')) {
    out.push({
      channel: 'support',
      to: '/support?create=1',
      icon: 'confirmation_number',
      titleKey: 'demandes.triage_support_title',
      descKey: 'demandes.triage_support_desc'
    })
  }
  if (isChannelEnabled('maintenance') && can('maintenance', 'E')) {
    out.push({
      channel: 'maintenance',
      to: '/maintenance?create=1',
      icon: 'build',
      titleKey: 'demandes.triage_maintenance_title',
      descKey: 'demandes.triage_maintenance_desc'
    })
  }
  return out
})

onMounted(async () => {
  await fetchSettings()
  if (activeChannelCount.value < 2) {
    const ch = settings.value?.channelsEnabled
    if (ch?.tma && hasModule('tma') && can('tma', 'E')) {
      await navigateTo('/tma?create=1', { replace: true })
      return
    }
    if (ch?.support && can('support', 'E')) {
      await navigateTo('/support?create=1', { replace: true })
      return
    }
    if (ch?.maintenance && can('maintenance', 'E')) {
      await navigateTo('/maintenance?create=1', { replace: true })
      return
    }
  }
})

watch(cards, (list) => {
  if (list.length === 1 && route.path === '/demandes/nouveau') {
    navigateTo(list[0].to, { replace: true })
  }
})
</script>

<template>
  <div>
    <AppPageHeader :title="t('demandes.triage_title')" :subtitle="t('demandes.triage_subtitle')" />

    <AppCard v-if="loaded && cards.length === 0" padding="lg">
      <p class="triage-empty">{{ $t('demandes.triage_empty') }}</p>
    </AppCard>

    <div v-else class="triage-grid">
      <AppCard
        v-for="card in cards"
        :key="card.channel"
        padding="lg"
        class="triage-card"
        role="button"
        tabindex="0"
        @click="navigateTo(card.to)"
        @keydown.enter="navigateTo(card.to)"
      >
        <AppIcon :name="card.icon" class="triage-card__icon" />
        <h2 class="triage-card__title">{{ t(card.titleKey) }}</h2>
        <p class="triage-card__desc">{{ t(card.descKey) }}</p>
      </AppCard>
    </div>
  </div>
</template>

<style scoped>
.triage-empty {
  margin: 0;
  color: var(--kore-text-muted);
  font-size: var(--kore-text-small);
}

.triage-grid {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(16rem, 1fr));
  gap: var(--kore-space-md);
}

.triage-card {
  cursor: pointer;
  transition: border-color 0.15s, box-shadow 0.15s;
}

.triage-card:hover,
.triage-card:focus-visible {
  border-color: var(--kore-gold);
  box-shadow: var(--kore-shadow-sm);
  outline: none;
}

.triage-card__icon {
  font-size: 2rem;
  color: var(--kore-gold);
  margin-bottom: var(--kore-space-sm);
}

.triage-card__title {
  margin: 0 0 0.5rem;
  font-size: var(--kore-text-body);
  font-weight: 600;
}

.triage-card__desc {
  margin: 0;
  font-size: var(--kore-text-small);
  color: var(--kore-text-muted);
}
</style>
