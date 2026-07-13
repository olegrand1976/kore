<script setup lang="ts">
export type AiBadgeVariant = 'generated' | 'assistant'

defineProps<{ variant?: AiBadgeVariant }>()
</script>

<template>
  <span class="app-ai-badge" :class="[`app-ai-badge--${variant || 'generated'}`]">
    <AppIcon name="auto_awesome" class="app-ai-badge__icon" aria-hidden="true" />
    <span>{{ variant === 'assistant' ? $t('ai.badge.assistant') : $t('ai.badge.generated') }}</span>
  </span>
</template>

<style scoped>
.app-ai-badge {
  display: inline-flex;
  align-items: center;
  gap: 0.25rem;
  padding: 0.125rem 0.5rem;
  font-size: var(--kore-text-caption);
  font-weight: 600;
  border-radius: var(--kore-radius-full);
  background: rgba(99, 102, 241, 0.12);
  color: var(--kore-brand-blue, #6366f1);
}

.app-ai-badge--assistant {
  background: rgba(201, 162, 39, 0.15);
  color: var(--kore-brand-gold);
}

.app-ai-badge__icon {
  font-size: 0.875rem;
}
</style>
