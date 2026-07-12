<template>
  <div class="page-shell">
    <PublicPageHero
      eyebrow="Kore SaaS"
      :title="$t('billing.checkout_success_title')"
      :subtitle="$t('billing.checkout_success_sub')"
      centered
    />
    <PublicButton variant="primary" to="/dashboard">{{ $t('nav.dashboard') }}</PublicButton>
  </div>
</template>

<script setup lang="ts">
definePageMeta({ layout: 'public' })
</script>
