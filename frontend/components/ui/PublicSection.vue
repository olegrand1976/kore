<script setup lang="ts">
defineProps<{
  title?: string
  subtitle?: string
  eyebrow?: string
  centered?: boolean
  variant?: 'default' | 'elevated'
}>()
</script>

<template>
  <section
    class="pub-section"
    :class="{
      'pub-section--centered': centered,
      'pub-section--elevated': variant === 'elevated'
    }"
  >
    <header v-if="eyebrow || title || subtitle" class="pub-section__header">
      <span v-if="eyebrow" class="page-hero__eyebrow">{{ eyebrow }}</span>
      <h2 v-if="title" class="pub-section__title">{{ title }}</h2>
      <p v-if="subtitle" class="pub-section__subtitle">{{ subtitle }}</p>
    </header>
    <slot />
  </section>
</template>

<style scoped>
.pub-section {
  padding: var(--kore-space-2xl) 0;
}

.pub-section--elevated {
  margin: var(--kore-space-xl) 0;
  padding: var(--kore-space-2xl) var(--kore-space-xl);
  border-radius: var(--kore-radius-lg);
  background: var(--kore-bg-elevated);
  border: 1px solid var(--kore-border);
}

.pub-section--centered {
  text-align: center;
}

.pub-section--centered .pub-section__header {
  max-width: 640px;
  margin-left: auto;
  margin-right: auto;
}

.pub-section__header .page-hero__eyebrow {
  margin-bottom: var(--kore-space-md);
}

.pub-section__title {
  font-size: var(--kore-text-h2);
  color: var(--kore-text);
}

.pub-section__subtitle {
  font-size: var(--kore-text-body);
  color: var(--kore-text-muted);
  margin: 0;
}
</style>
