<script setup lang="ts">
defineProps<{
  eyebrow?: string
  title: string
  subtitle?: string
  centered?: boolean
}>()
</script>

<template>
  <section class="page-hero" :class="{ 'page-hero--centered': centered }">
    <div class="page-hero__inner">
      <span v-if="eyebrow" class="page-hero__eyebrow">{{ eyebrow }}</span>
      <h1 class="page-hero__title">{{ title }}</h1>
      <p v-if="subtitle" class="page-hero__subtitle">{{ subtitle }}</p>
      <div v-if="$slots.default" class="page-hero__actions">
        <slot />
      </div>
    </div>
  </section>
</template>

<style scoped>
.page-hero__actions {
  display: flex;
  flex-wrap: wrap;
  gap: var(--kore-space-md);
  margin-top: var(--kore-space-xl);
  justify-content: inherit;
}

.page-hero--centered .page-hero__actions {
  justify-content: center;
}
</style>
