<template>
  <div class="help-access">
    <AppPageHeader :title="$t('help.access.title')" :subtitle="$t('help.access.subtitle')">
      <template #actions>
        <AppButton to="/aide" variant="ghost" size="sm">{{ $t('help.back') }}</AppButton>
      </template>
    </AppPageHeader>

    <AppCard padding="lg">
      <h2 class="help-access__section-title">{{ $t('help.access.legend_title') }}</h2>
      <ul class="help-access__legend">
        <li><strong>L</strong> — {{ $t('help.access.legend_l') }}</li>
        <li><strong>E</strong> — {{ $t('help.access.legend_e') }}</li>
        <li><strong>V</strong> — {{ $t('help.access.legend_v') }}</li>
        <li><strong>—</strong> — {{ $t('help.access.legend_none') }}</li>
      </ul>
      <p class="help-access__note">{{ $t('help.access.union_note') }}</p>
    </AppCard>

    <AppCard v-if="myUnknown.length" padding="lg">
      <p class="help-access__note help-access__note--flush">
        {{ $t('help.access.unknown_profiles', { profiles: myUnknown.join(', ') }) }}
      </p>
    </AppCard>

    <AppCard padding="lg">
      <h2 class="help-access__section-title">{{ $t('help.access.matrix_title') }}</h2>
      <p class="help-access__note">{{ $t('help.access.matrix_hint') }}</p>
      <div class="help-access__table-wrap">
        <table class="help-access__table">
          <thead>
            <tr>
              <th scope="col">{{ $t('help.access.col_profile') }}</th>
              <th v-for="mod in matrixModules" :key="mod" scope="col">
                {{ $t(`help.access.modules.${mod}`) }}
              </th>
            </tr>
          </thead>
          <tbody>
            <tr
              v-for="profile in implementedProfiles"
              :key="profile"
              :class="{ 'help-access__table-row--mine': myProfiles.includes(profile) }"
            >
              <th scope="row">{{ profile }}</th>
              <td v-for="mod in matrixModules" :key="mod">
                {{ formatRbacCell(profile, mod) }}
              </td>
            </tr>
          </tbody>
        </table>
      </div>
    </AppCard>

    <AppCard padding="lg">
      <h2 class="help-access__section-title">{{ $t('help.access.profiles_title') }}</h2>
      <HelpProfileCard
        v-for="profile in implementedProfiles"
        :key="profile"
        :profile="profile"
        :highlight="myProfiles.includes(profile)"
        show-seed
      />
    </AppCard>

    <AppCard padding="lg">
      <h2 class="help-access__section-title">{{ $t('help.access.planned_title') }}</h2>
      <p class="help-access__note">{{ $t('help.access.planned_hint') }}</p>
      <ul class="help-access__planned">
        <li v-for="profile in plannedProfiles" :key="profile">
          <strong>{{ profile }}</strong>
          <span>{{ plannedDescription(profile) }}</span>
        </li>
      </ul>
    </AppCard>
  </div>
</template>

<script setup lang="ts">
import {
  formatRbacCell,
  HELP_MATRIX_MODULES,
  IMPLEMENTED_RBAC_PROFILES,
  isImplementedRbacProfile,
  PLANNED_RBAC_PROFILES,
  type RbacModule
} from '~/utils/rbac'

definePageMeta({ layout: 'default' })

const { t, te } = useI18n()
const { effectiveProfiles } = useAuth()

const matrixModules: RbacModule[] = HELP_MATRIX_MODULES
const implementedProfiles = [...IMPLEMENTED_RBAC_PROFILES]
const plannedProfiles = [...PLANNED_RBAC_PROFILES]

const myProfiles = computed(() => effectiveProfiles.value)

const myUnknown = computed(() => myProfiles.value.filter((p) => !isImplementedRbacProfile(p)))

const plannedKeyMap: Record<string, string> = {
  Utilisateur: 'utilisateur',
  Commercial: 'commercial',
  Support: 'support',
  'Chef utilisateur': 'chef_utilisateur',
  'Client externe': 'client_externe',
  'Sous-traitant': 'sous_traitant'
}

function plannedDescription(profile: string): string {
  const key = plannedKeyMap[profile]
  if (!key) return t('help.access.planned_fallback')
  const i18nKey = `help.access.planned.${key}`
  return te(i18nKey) ? t(i18nKey) : t('help.access.planned_fallback')
}
</script>

<style scoped>
.help-access {
  display: flex;
  flex-direction: column;
  gap: var(--kore-space-lg);
}

.help-access__section-title {
  margin: 0 0 var(--kore-space-md);
  font-size: var(--kore-text-h3);
  font-weight: 600;
  color: var(--kore-text);
}

.help-access__legend {
  margin: 0;
  padding-left: 1.25rem;
  color: var(--kore-text);
  line-height: 1.55;
}

.help-access__note {
  margin: var(--kore-space-md) 0 0;
  font-size: var(--kore-text-small);
  color: var(--kore-text-muted);
  line-height: 1.45;
}

.help-access__note--flush {
  margin: 0;
}

.help-access__table-wrap {
  overflow-x: auto;
  -webkit-overflow-scrolling: touch;
  margin-top: var(--kore-space-md);
}

.help-access__table {
  width: 100%;
  border-collapse: collapse;
  font-size: var(--kore-text-small);
  min-width: 720px;
}

.help-access__table th,
.help-access__table td {
  padding: var(--kore-space-sm) var(--kore-space-md);
  border: 1px solid var(--kore-border);
  text-align: center;
  white-space: nowrap;
}

.help-access__table thead th {
  background: var(--kore-bg-subtle);
  font-weight: 600;
  color: var(--kore-text);
}

.help-access__table tbody th {
  text-align: left;
  font-weight: 600;
  color: var(--kore-text);
  background: var(--kore-bg-subtle);
}

.help-access__table-row--mine td,
.help-access__table-row--mine th {
  background: color-mix(in srgb, var(--kore-brand-gold) 14%, transparent);
}

.help-access__planned {
  margin: var(--kore-space-md) 0 0;
  padding: 0;
  list-style: none;
  display: flex;
  flex-direction: column;
  gap: var(--kore-space-md);
}

.help-access__planned li {
  display: flex;
  flex-direction: column;
  gap: 0.25rem;
  font-size: var(--kore-text-small);
  color: var(--kore-text-muted);
  line-height: 1.45;
}

.help-access__planned strong {
  color: var(--kore-text);
}
</style>
