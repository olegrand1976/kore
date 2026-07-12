<template>
  <div class="page-shell">
    <PublicPageHero
      eyebrow="Kore SaaS"
      :title="$t('billing.checkout_cancel_title')"
      :subtitle="$t('billing.checkout_cancel_sub')"
      centered
    />
    <PublicButton variant="primary" to="/tarifs">{{ $t('brand.cta_pricing') }}</PublicButton>
  </div>
</template>

<script setup lang="ts">
definePageMeta({ layout: 'public' })
</script>
