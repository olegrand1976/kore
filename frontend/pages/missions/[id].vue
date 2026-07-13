<template>
  <div>
    <AppPageHeader :title="pageTitle" :subtitle="$t('fiche.mission_title')">
      <template #actions>
        <AppButton variant="ghost" size="sm" @click="navigateTo('/cra')">
          <AppIcon name="arrow_back" /> {{ $t('fiche.back_cra') }}
        </AppButton>
      </template>
    </AppPageHeader>

    <AppCard v-if="pending" padding="lg">
      <p class="muted">{{ $t('fiche.loading') }}</p>
    </AppCard>

    <AppCard v-else-if="error" padding="lg">
      <AppEmptyState icon="error" :title="$t('fiche.not_found')" />
    </AppCard>

    <template v-else-if="mission">
      <AppKpiGrid compact>
        <AppKpiCard
          icon="flag"
          tone="gold"
          :value="missionStatusLabel(mission.status)"
          :label="$t('fiche.col_status')"
        />
        <AppKpiCard
          icon="payments"
          tone="blue"
          :value="tjmLabel"
          :label="$t('fiche.col_tjm')"
        />
        <AppKpiCard
          icon="groups"
          tone="success"
          :value="collaborators.length"
          :label="$t('fiche.section_staffing')"
        />
        <AppKpiCard
          icon="event"
          tone="warn"
          :value="periodLabel"
          :label="$t('fiche.col_period')"
        />
      </AppKpiGrid>

      <div class="fiche-grid">
        <AppCard padding="lg">
          <h3 class="fiche-section-title">{{ $t('fiche.section_overview') }}</h3>
          <dl class="fiche-dl">
            <div>
              <dt>{{ $t('fiche.col_status') }}</dt>
              <dd>
                <AppBadge :variant="missionStatusVariant(mission.status)">
                  {{ missionStatusLabel(mission.status) }}
                </AppBadge>
              </dd>
            </div>
            <div>
              <dt>{{ $t('fiche.col_client') }}</dt>
              <dd>
                <NuxtLink
                  v-if="mission.clientId && mission.clientName"
                  :to="`/clients/${mission.clientId}`"
                  class="fiche-link"
                >
                  {{ mission.clientName }}
                </NuxtLink>
                <span v-else>{{ mission.clientName || $t('fiche.none') }}</span>
              </dd>
            </div>
            <div>
              <dt>{{ $t('fiche.col_start') }}</dt>
              <dd>{{ formatDate(mission.startDate) }}</dd>
            </div>
            <div>
              <dt>{{ $t('fiche.col_end') }}</dt>
              <dd>{{ mission.endDate ? formatDate(mission.endDate) : $t('fiche.none') }}</dd>
            </div>
            <div>
              <dt>{{ $t('fiche.col_tjm') }}</dt>
              <dd>{{ tjmLabel }}</dd>
            </div>
            <div v-if="mission.clientContact">
              <dt>{{ $t('fiche.col_client_contact') }}</dt>
              <dd>{{ mission.clientContact }}</dd>
            </div>
            <div>
              <dt>{{ $t('fiche.col_created') }}</dt>
              <dd>{{ formatDate(mission.createdAt) }}</dd>
            </div>
          </dl>
        </AppCard>

        <AppCard padding="lg">
          <h3 class="fiche-section-title">{{ $t('fiche.col_technologies') }}</h3>
          <div v-if="mission.technologies.length" class="fiche-tags">
            <AppBadge v-for="tech in mission.technologies" :key="tech" variant="default">
              {{ tech }}
            </AppBadge>
          </div>
          <p v-else class="muted">{{ $t('fiche.none') }}</p>
        </AppCard>
      </div>

      <AppCard padding="none" class="fiche-table-wrap">
        <div class="fiche-table-head">
          <h3 class="fiche-section-title">{{ $t('fiche.section_staffing') }}</h3>
        </div>
        <AppTable
          :columns="staffColumns"
          :rows="staffRows"
          row-key="id"
          :empty-title="$t('fiche.staffing_empty')"
        >
          <template #cell-name="{ row }">
            <NuxtLink :to="`/collaborateurs/${row.id}`" class="fiche-link fiche-strong">
              {{ row.name }}
            </NuxtLink>
          </template>
          <template #cell-login="{ value }">
            <span class="muted-small">{{ value }}</span>
          </template>
        </AppTable>
      </AppCard>
    </template>
  </div>
</template>

<script setup lang="ts">
import { formatUserDisplayName } from '~/composables/useUserDisplay'

definePageMeta({ layout: 'default' })

type MissionCollaborator = {
  userId?: string
  login?: string
  prenom?: string
  nom?: string
}

type MissionDetail = {
  id?: string
  clientId?: string
  clientName?: string
  status: string
  startDate?: string
  endDate?: string | null
  tjmAmount?: number
  currency?: string
  technologies: string[]
  clientContact?: string
  createdAt?: string
  collaborators?: MissionCollaborator[]
}

const route = useRoute()
const { t } = useI18n()
const { formatDate, formatMoney, missionStatusLabel, missionStatusVariant } = useFicheFormat()

const id = computed(() => String(route.params.id ?? ''))

const { data, pending, error } = await useFetch<MissionDetail>(() => `/api/ssii/missions/${id.value}`, {
  watch: [id]
})

const mission = computed(() => {
  const payload = (data.value as { data?: MissionDetail })?.data ?? data.value
  if (!payload || typeof payload !== 'object') return null
  return {
    ...payload,
    status: payload.status ?? 'active',
    technologies: payload.technologies ?? [],
    collaborators: payload.collaborators ?? []
  }
})

const collaborators = computed(() => mission.value?.collaborators ?? [])

const pageTitle = computed(() => {
  if (mission.value?.clientName) {
    return `${t('fiche.mission_title')} — ${mission.value.clientName}`
  }
  return t('fiche.mission_title')
})

const tjmLabel = computed(() =>
  formatMoney(Number(mission.value?.tjmAmount ?? 0), mission.value?.currency ?? 'EUR')
)

const periodLabel = computed(() => {
  if (!mission.value?.startDate) return '—'
  const start = formatDate(mission.value.startDate)
  if (!mission.value.endDate) return start
  return `${start} → ${formatDate(mission.value.endDate)}`
})

const staffColumns = computed(() => [
  { key: 'name', label: t('fiche.col_name') },
  { key: 'login', label: t('fiche.col_login') }
])

const staffRows = computed(() =>
  collaborators.value.map((c) => ({
    id: String(c.userId ?? ''),
    name: formatUserDisplayName(c.prenom, c.nom, c.login),
    login: c.login ?? '—'
  }))
)
</script>

<style scoped>
.muted { color: var(--kore-text-muted); }

.muted-small {
  color: var(--kore-text-muted);
  font-size: var(--kore-text-caption);
}

.fiche-grid {
  display: grid;
  gap: var(--kore-space-lg);
  margin-bottom: var(--kore-space-lg);
}

@media (min-width: 768px) {
  .fiche-grid {
    grid-template-columns: 2fr 1fr;
  }
}

.fiche-section-title {
  margin: 0 0 var(--kore-space-md);
  font-size: var(--kore-text-h3);
}

.fiche-dl {
  display: grid;
  gap: var(--kore-space-md);
  margin: 0;
}

.fiche-dl div {
  display: grid;
  gap: var(--kore-space-xs);
}

.fiche-dl dt {
  margin: 0;
  font-size: var(--kore-text-caption);
  color: var(--kore-text-muted);
  font-weight: 500;
}

.fiche-dl dd {
  margin: 0;
  color: var(--kore-text);
  font-weight: 500;
}

.fiche-tags {
  display: flex;
  flex-wrap: wrap;
  gap: var(--kore-space-xs);
}

.fiche-link {
  color: var(--kore-brand-blue);
  text-decoration: none;
  font-weight: 500;
}

.fiche-link:hover { text-decoration: underline; }

.fiche-strong { font-weight: 600; }

.fiche-table-wrap { overflow: hidden; }

.fiche-table-head {
  padding: var(--kore-space-lg) var(--kore-space-lg) 0;
}
</style>
