<template>
  <div class="help-project">
    <AppPageHeader :title="$t('help.project.title')" :subtitle="$t('help.project.intro')">
      <template #actions>
        <AppButton variant="ghost" size="sm" @click="navigateTo('/aide')">{{ $t('help.back') }}</AppButton>
      </template>
    </AppPageHeader>

    <AppCard padding="lg">
      <ul class="help-project__list">
        <li>{{ $t('help.project.bullets.backlog') }}</li>
        <li>{{ $t('help.project.bullets.sprint') }}</li>
        <li>{{ $t('help.project.bullets.kanban') }}</li>
        <li>{{ $t('help.project.bullets.metrics') }}</li>
      </ul>
      <AppButton variant="primary" size="sm" @click="navigateTo('/projets')">{{ $t('nav.projets') }}</AppButton>
    </AppCard>
  </div>
</template>

<script setup lang="ts">
definePageMeta({ layout: 'default' })
</script>

<style scoped>
.help-project__list {
  margin: 0 0 var(--kore-space-lg);
  padding-left: var(--kore-space-md);
  line-height: 1.5;
}
</style>
