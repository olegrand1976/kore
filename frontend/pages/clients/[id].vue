<template>
  <div>
    <AppPageHeader :title="title" :subtitle="$t('fiche.client_title')">
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

    <template v-else-if="client">
      <AppKpiGrid compact>
        <AppKpiCard
          icon="corporate_fare"
          tone="gold"
          :value="title"
          :label="$t('fiche.col_company')"
        />
        <AppKpiCard
          icon="receipt_long"
          tone="blue"
          :value="client.tva || $t('fiche.none')"
          :label="$t('fiche.col_vat')"
        />
        <AppKpiCard
          icon="groups"
          tone="success"
          :value="contacts.length"
          :label="$t('fiche.section_contacts')"
        />
        <AppKpiCard
          icon="work"
          tone="warn"
          :value="missions.length"
          :label="$t('fiche.section_missions')"
        />
      </AppKpiGrid>

      <AppCard padding="lg" class="fiche-block">
        <h3 class="fiche-section-title">{{ $t('fiche.section_company') }}</h3>
        <dl class="fiche-dl fiche-dl--grid">
          <div>
            <dt>{{ $t('fiche.col_company') }}</dt>
            <dd>{{ title }}</dd>
          </div>
          <div>
            <dt>{{ $t('fiche.col_vat') }}</dt>
            <dd>{{ client.tva || $t('fiche.none') }}</dd>
          </div>
          <div>
            <dt>{{ $t('fiche.col_created') }}</dt>
            <dd>{{ formatDate(client.createdAt) }}</dd>
          </div>
        </dl>
      </AppCard>

      <AppCard padding="none" class="fiche-table-wrap fiche-block">
        <div class="fiche-table-head">
          <h3 class="fiche-section-title">{{ $t('fiche.section_contacts') }}</h3>
        </div>
        <AppTable
          :columns="contactColumns"
          :rows="contactRows"
          row-key="email"
          :empty-title="$t('fiche.contacts_empty')"
        >
          <template #cell-name="{ row }">
            <span class="fiche-strong">{{ row.name }}</span>
          </template>
          <template #cell-email="{ value }">
            <a v-if="value" :href="`mailto:${value}`" class="fiche-link">{{ value }}</a>
            <span v-else>{{ $t('fiche.none') }}</span>
          </template>
        </AppTable>
      </AppCard>

      <AppCard padding="none" class="fiche-table-wrap">
        <div class="fiche-table-head">
          <h3 class="fiche-section-title">{{ $t('fiche.section_missions') }}</h3>
        </div>
        <AppTable
          :columns="missionColumns"
          :rows="missionRows"
          row-key="id"
          :loading="missionsPending"
          :empty-title="$t('fiche.missions_empty')"
        >
          <template #cell-status="{ value }">
            <AppBadge :variant="missionStatusVariant(String(value))">
              {{ missionStatusLabel(String(value)) }}
            </AppBadge>
          </template>
          <template #cell-period="{ row }">
            <span class="fiche-nowrap">{{ row.period }}</span>
          </template>
          <template #cell-tjm="{ value }">
            <span class="fiche-nowrap">{{ value }}</span>
          </template>
          <template #cell-actions="{ row }">
            <AppButton variant="ghost" size="sm" @click="navigateTo(`/missions/${row.id}`)">
              {{ $t('fiche.open_mission') }}
            </AppButton>
          </template>
        </AppTable>
      </AppCard>
    </template>
  </div>
</template>

<script setup lang="ts">
definePageMeta({ layout: 'default' })

type ClientContact = {
  nom?: string
  prenom?: string
  email?: string
  role?: string
  telephone?: string
}

type ClientDetail = {
  id?: string
  raisonSociale?: string
  tva?: string
  contacts?: ClientContact[]
  createdAt?: string
}

type MissionSummary = {
  id: string
  clientId?: string
  status?: string
  startDate?: string
  endDate?: string
  tjmAmount?: number
  currency?: string
}

const route = useRoute()
const { t, locale } = useI18n()
const { formatDate, formatMoney, missionStatusLabel, missionStatusVariant } = useFicheFormat()

const id = computed(() => String(route.params.id ?? ''))

const { data, pending, error } = await useFetch<ClientDetail>(() => `/api/org/clients/${id.value}`, {
  watch: [id]
})

const { data: missionsData, pending: missionsPending } = await useFetch('/api/ssii/missions', {
  lazy: true,
  server: false
})

const client = computed(() => {
  const payload = (data.value as { data?: ClientDetail })?.data ?? data.value
  return payload && typeof payload === 'object' ? payload : null
})

const title = computed(() => client.value?.raisonSociale ?? '—')

const contacts = computed(() => client.value?.contacts ?? [])

const contactColumns = computed(() => [
  { key: 'name', label: t('fiche.col_contact_name') },
  { key: 'role', label: t('fiche.col_contact_role') },
  { key: 'email', label: t('fiche.col_email') },
  { key: 'phone', label: t('fiche.col_contact_phone') }
])

const contactRows = computed(() =>
  contacts.value.map((c) => ({
    email: c.email ?? `${c.prenom}-${c.nom}`,
    name: [c.prenom, c.nom].filter(Boolean).join(' ') || '—',
    role: c.role || '—',
    phone: c.telephone || '—'
  }))
)

const missions = computed((): MissionSummary[] => {
  const payload = (missionsData.value as { data?: MissionSummary[] })?.data ?? missionsData.value
  if (!Array.isArray(payload)) return []
  return payload.filter((m) => String(m.clientId ?? '') === id.value)
})

const missionColumns = computed(() => [
  { key: 'period', label: t('fiche.col_period'), nowrap: true },
  { key: 'status', label: t('fiche.col_status') },
  { key: 'tjm', label: t('fiche.col_tjm'), nowrap: true },
  { key: 'actions', label: '' }
])

const formatPeriod = (start?: string, end?: string) => {
  if (!start) return '—'
  const startLabel = formatDate(start)
  if (!end) return startLabel
  return `${startLabel} → ${formatDate(end)}`
}

const missionRows = computed(() =>
  missions.value.map((m) => ({
    id: m.id,
    period: formatPeriod(m.startDate, m.endDate),
    status: m.status ?? '',
    tjm: formatMoney(Number(m.tjmAmount ?? 0), m.currency ?? 'EUR'),
    actions: ''
  }))
)
</script>

<style scoped>
.muted { color: var(--kore-text-muted); }

.fiche-block { margin-bottom: var(--kore-space-lg); }

.fiche-section-title {
  margin: 0 0 var(--kore-space-md);
  font-size: var(--kore-text-h3);
}

.fiche-dl {
  display: grid;
  gap: var(--kore-space-md);
  margin: 0;
}

.fiche-dl--grid {
  grid-template-columns: 1fr;
}

@media (min-width: 768px) {
  .fiche-dl--grid {
    grid-template-columns: repeat(3, minmax(0, 1fr));
  }
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

.fiche-table-wrap { overflow: hidden; }

.fiche-table-head {
  padding: var(--kore-space-lg) var(--kore-space-lg) 0;
}

.fiche-strong { font-weight: 600; }

.fiche-link {
  color: var(--kore-brand-blue);
  text-decoration: none;
}

.fiche-link:hover { text-decoration: underline; }

.fiche-nowrap { white-space: nowrap; }
</style>
