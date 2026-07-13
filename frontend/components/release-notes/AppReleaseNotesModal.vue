<script setup lang="ts">
export type ReleaseNotesCommit = {
  sha: string
  shortSha: string
  message: string
  authorName?: string
  date: string // ISO
  htmlUrl?: string
}

export type ReleaseNotesMonth = {
  key: string // YYYY-MM
  label: string
  items: ReleaseNotesCommit[]
}

const props = defineProps<{
  open: boolean
  months: ReleaseNotesMonth[]
  selectedMonthKey: string
  loading?: boolean
  currentVersion?: string
  lastSeenVersion?: string
  autoShowEnabled: boolean
}>()

const emit = defineEmits<{
  'update:open': [open: boolean]
  'update:selectedMonthKey': [key: string]
  'update:autoShowEnabled': [enabled: boolean]
  markSeen: []
  refresh: []
}>()

const { t } = useI18n()

const titleId = 'release-notes-title'

const modelOpen = computed({
  get: () => props.open,
  set: (v: boolean) => emit('update:open', v)
})

const selectedMonth = computed(() => props.months.find((m) => m.key === props.selectedMonthKey))
const hasContent = computed(() => (selectedMonth.value?.items?.length ?? 0) > 0)

const onToggleAutoShow = (e: Event) => {
  const enabled = (e.target as HTMLInputElement).checked
  emit('update:autoShowEnabled', enabled)
}

const onChangeMonth = (e: Event) => {
  emit('update:selectedMonthKey', (e.target as HTMLSelectElement).value)
}
</script>

<template>
  <AppModal v-model:open="modelOpen" width="md" :title-id="titleId" :aria-label="t('release_notes.title')">
    <header class="rn__header">
      <div>
        <h2 :id="titleId" class="rn__title">{{ t('release_notes.title') }}</h2>
        <p class="rn__subtitle">
          <span v-if="currentVersion">{{ t('release_notes.current_version', { version: currentVersion }) }}</span>
          <span v-if="lastSeenVersion" class="rn__sep">·</span>
          <span v-if="lastSeenVersion">{{ t('release_notes.last_seen_version', { version: lastSeenVersion }) }}</span>
        </p>
      </div>
      <button type="button" class="rn__close" :aria-label="t('common.close')" @click="$emit('update:open', false)">
        <AppIcon name="close" />
      </button>
    </header>

    <div class="rn__controls">
      <label class="rn__control">
        <span class="rn__label">{{ t('release_notes.month') }}</span>
        <select class="rn__select" :disabled="loading || months.length === 0" :value="selectedMonthKey" @change="onChangeMonth">
          <option v-for="m in months" :key="m.key" :value="m.key">{{ m.label }}</option>
        </select>
      </label>

      <label class="rn__check">
        <input type="checkbox" :checked="autoShowEnabled" @change="onToggleAutoShow" />
        {{ t('release_notes.auto_show') }}
      </label>
    </div>

    <div class="rn__body">
      <div v-if="loading" class="rn__loading">{{ t('common.loading') }}</div>
      <div v-else-if="months.length === 0" class="rn__empty">{{ t('release_notes.empty') }}</div>
      <div v-else-if="!hasContent" class="rn__empty">{{ t('release_notes.empty_month') }}</div>
      <ul v-else class="rn__list" :aria-label="t('release_notes.list_aria')">
        <li v-for="item in selectedMonth?.items" :key="item.sha" class="rn__item">
          <div class="rn__itemMain">
            <p class="rn__msg">{{ item.message }}</p>
            <p class="rn__meta">
              <span class="rn__sha">{{ item.shortSha }}</span>
              <span v-if="item.authorName" class="rn__sep">·</span>
              <span v-if="item.authorName">{{ item.authorName }}</span>
              <span class="rn__sep">·</span>
              <time :datetime="item.date">{{ new Date(item.date).toLocaleDateString() }}</time>
              <span v-if="item.htmlUrl" class="rn__sep">·</span>
              <a v-if="item.htmlUrl" class="rn__link" :href="item.htmlUrl" target="_blank" rel="noreferrer">
                {{ t('release_notes.open_commit') }}
              </a>
            </p>
          </div>
        </li>
      </ul>
    </div>

    <footer class="rn__footer">
      <AppButton variant="ghost" size="sm" :disabled="loading" @click="$emit('refresh')">
        <AppIcon name="refresh" />
        {{ t('release_notes.refresh') }}
      </AppButton>

      <div class="rn__footerRight">
        <AppButton variant="secondary" size="sm" :disabled="loading" @click="$emit('update:open', false)">
          {{ t('common.close') }}
        </AppButton>
        <AppButton variant="primary" size="sm" :disabled="loading" @click="$emit('markSeen')">
          {{ t('release_notes.mark_seen') }}
        </AppButton>
      </div>
    </footer>
  </AppModal>
</template>

<style scoped>
.rn__header {
  display: flex;
  justify-content: space-between;
  align-items: flex-start;
  gap: var(--kore-space-md);
  margin-bottom: var(--kore-space-lg);
}

.rn__title {
  margin: 0;
  font-size: var(--kore-text-h3);
}

.rn__subtitle {
  margin: var(--kore-space-xs) 0 0;
  font-size: var(--kore-text-caption);
  color: var(--kore-text-muted);
}

.rn__sep { margin: 0 var(--kore-space-xs); }

.rn__close {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  width: 2.25rem;
  height: 2.25rem;
  border-radius: var(--kore-radius-md);
  border: 1px solid var(--kore-border);
  background: var(--kore-bg);
  color: var(--kore-text-muted);
  cursor: pointer;
}

.rn__controls {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: var(--kore-space-md);
  margin-bottom: var(--kore-space-lg);
  flex-wrap: wrap;
}

.rn__control {
  display: grid;
  gap: var(--kore-space-xs);
}

.rn__label {
  font-size: var(--kore-text-caption);
  color: var(--kore-text-muted);
  font-weight: 600;
}

.rn__select {
  min-width: 14rem;
  padding: 0.625rem 0.875rem;
  border: 1px solid var(--kore-border);
  border-radius: var(--kore-radius-md);
  background: var(--kore-bg);
  color: var(--kore-text);
  font-family: var(--kore-font);
  font-size: var(--kore-text-small);
}

.rn__check {
  display: inline-flex;
  align-items: center;
  gap: var(--kore-space-sm);
  font-size: var(--kore-text-small);
  color: var(--kore-text);
}

.rn__body {
  border-top: 1px solid var(--kore-border);
  border-bottom: 1px solid var(--kore-border);
  padding: var(--kore-space-lg) 0;
  margin: var(--kore-space-lg) 0;
}

.rn__loading,
.rn__empty {
  color: var(--kore-text-muted);
  font-size: var(--kore-text-small);
}

.rn__list {
  list-style: none;
  padding: 0;
  margin: 0;
  display: grid;
  gap: var(--kore-space-md);
}

.rn__item {
  padding: var(--kore-space-md);
  border: 1px solid var(--kore-border);
  border-radius: var(--kore-radius-lg);
  background: color-mix(in srgb, var(--kore-bg) 65%, transparent);
}

.rn__msg {
  margin: 0;
  font-size: var(--kore-text-small);
  font-weight: 600;
}

.rn__meta {
  margin: var(--kore-space-xs) 0 0;
  font-size: var(--kore-text-caption);
  color: var(--kore-text-muted);
}

.rn__sha {
  font-family: ui-monospace, SFMono-Regular, Menlo, Monaco, Consolas, "Liberation Mono", "Courier New", monospace;
}

.rn__link {
  color: var(--kore-brand-gold);
  text-decoration: none;
}
.rn__link:hover { text-decoration: underline; }

.rn__footer {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: var(--kore-space-md);
  flex-wrap: wrap;
}

.rn__footerRight {
  display: inline-flex;
  align-items: center;
  gap: var(--kore-space-sm);
}

@media (max-width: 768px) {
  .rn__select { min-width: 100%; }
  .rn__footerRight { width: 100%; }
  .rn__footerRight :deep(.app-btn) { flex: 1; }
}
</style>
