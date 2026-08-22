<template>
  <div class="help-tma">
    <AppPageHeader :title="$t('help.tma.title')" :subtitle="$t('help.tma.intro')">
      <template #actions>
        <AppButton variant="ghost" size="sm" @click="navigateTo('/aide')">{{ $t('help.back') }}</AppButton>
      </template>
    </AppPageHeader>

    <AppCard padding="lg">
      <h2 class="help-tma__section-title">{{ $t('help.tma.panel_title') }}</h2>
      <p class="help-tma__text">{{ $t('help.tma.panel_body') }}</p>
      <ul class="help-tma__list">
        <li>{{ $t('help.tma.bullets.visible') }}</li>
        <li>{{ $t('help.tma.bullets.linked') }}</li>
        <li>{{ $t('help.tma.bullets.open') }}</li>
      </ul>
    </AppCard>

    <AppCard padding="lg">
      <h2 class="help-tma__section-title">{{ $t('help.tma.link_title') }}</h2>
      <p class="help-tma__text">{{ $t('help.tma.link_body') }}</p>
      <p class="help-tma__note muted">{{ $t('help.tma.link_note') }}</p>
    </AppCard>

    <AppCard padding="lg">
      <AppButton variant="primary" size="sm" @click="navigateTo('/tma')">{{ $t('nav.tma') }}</AppButton>
    </AppCard>
  </div>
</template>

<script setup lang="ts">
definePageMeta({ layout: 'default' })
</script>

<style scoped>
.help-tma__section-title {
  margin: 0 0 var(--kore-space-sm);
  font-size: var(--kore-text-body);
  font-weight: 600;
}
.help-tma__text {
  margin: 0 0 var(--kore-space-md);
  line-height: 1.5;
}
.help-tma__list {
  margin: 0;
  padding-left: var(--kore-space-md);
  line-height: 1.5;
}
.help-tma__note {
  margin: 0;
  font-size: var(--kore-text-small);
}
.muted {
  color: var(--kore-text-muted);
}
</style>
