<script setup lang="ts">
import type { ImplementedRbacProfile } from '~/utils/rbac'

const props = defineProps<{
  profile: ImplementedRbacProfile
  highlight?: boolean
  showSeed?: boolean
}>()

const { t, te, tm } = useI18n()
const { isAdmin } = useAuth()

const profileKeyMap: Record<ImplementedRbacProfile, string> = {
  Administrateur: 'admin',
  Collaborateur: 'collaborateur',
  "Chef d'équipe": 'chef_equipe',
  'Responsable de service': 'responsable'
}

const i18nKey = computed(() => profileKeyMap[props.profile])

const summary = computed(() => {
  const key = `help.access.profiles.${i18nKey.value}.summary`
  return te(key) ? t(key) : ''
})

const bullets = computed(() => {
  const raw = tm(`help.access.profiles.${i18nKey.value}.bullets`)
  return Array.isArray(raw) ? (raw as string[]) : []
})

const seed = computed(() => {
  if (!props.showSeed || !isAdmin.value) return ''
  const key = `help.access.profiles.${i18nKey.value}.seed`
  const value = te(key) ? t(key) : ''
  return value.trim()
})

const slug = computed(() => i18nKey.value)
</script>

<template>
  <div :id="slug" class="help-profile">
    <div class="help-profile__head">
      <h3>{{ profile }}</h3>
      <AppBadge v-if="highlight" variant="success">{{ $t('help.access.badge_yours') }}</AppBadge>
    </div>
    <p v-if="summary" class="help-profile__summary">{{ summary }}</p>
    <ul v-if="bullets.length" class="help-profile__bullets">
      <li v-for="(item, idx) in bullets" :key="idx">{{ item }}</li>
    </ul>
    <p v-if="seed" class="help-profile__seed">
      {{ $t('help.access.seed_label') }} : <code>{{ seed }}</code>
    </p>
  </div>
</template>

<style scoped>
.help-profile {
  padding: var(--kore-space-md) 0 0;
}

.help-profile + .help-profile {
  margin-top: var(--kore-space-md);
  border-top: 1px solid var(--kore-border);
  padding-top: var(--kore-space-md);
}

.help-profile__head {
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  gap: var(--kore-space-sm);
  margin-bottom: var(--kore-space-sm);
}

.help-profile__head h3 {
  margin: 0;
  font-size: var(--kore-text-body);
  font-weight: 600;
  color: var(--kore-text);
}

.help-profile__summary {
  margin: 0 0 var(--kore-space-sm);
  color: var(--kore-text-muted);
  font-size: var(--kore-text-small);
  line-height: 1.45;
}

.help-profile__bullets {
  margin: 0;
  padding-left: 1.25rem;
  color: var(--kore-text);
  font-size: var(--kore-text-small);
  line-height: 1.55;
}

.help-profile__seed {
  margin: var(--kore-space-sm) 0 0;
  font-size: var(--kore-text-small);
  color: var(--kore-text-muted);
}

.help-profile__seed code {
  font-family: ui-monospace, SFMono-Regular, Menlo, Monaco, Consolas, monospace;
  font-size: 0.9em;
}
</style>
