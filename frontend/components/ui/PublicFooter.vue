<script setup lang="ts">
const year = new Date().getFullYear()
</script>

<template>
  <footer class="pub-footer">
    <div class="pub-footer__inner">
      <div class="pub-footer__brand">
        <KoreLogo variant="horizontal" size="sm" tone="auto" :alt="$t('brand.name')" />
        <p class="pub-footer__tagline">{{ $t('brand.tagline') }}</p>
      </div>
      <nav class="pub-footer__nav" aria-label="Navigation">
        <NuxtLink to="/modules">{{ $t('nav.modules') }}</NuxtLink>
        <NuxtLink to="/tarifs">{{ $t('nav.pricing') }}</NuxtLink>
        <NuxtLink to="/reserver">{{ $t('nav.book') }}</NuxtLink>
        <NuxtLink to="/login">{{ $t('nav.login') }}</NuxtLink>
      </nav>
      <p class="pub-footer__copy">© {{ year }} Kore. {{ $t('footer.rights') }}</p>
    </div>
  </footer>
</template>

<style scoped>
.pub-footer {
  margin-top: auto;
  background: var(--kore-bg-subtle);
  border-top: 1px solid var(--kore-border);
}

.pub-footer__inner {
  max-width: var(--kore-public-container-max);
  margin: 0 auto;
  padding: var(--kore-space-2xl) var(--kore-space-xl);
  display: grid;
  gap: var(--kore-space-xl);
}

@media (min-width: 768px) {
  .pub-footer__inner {
    grid-template-columns: 1fr auto;
    align-items: start;
  }

  .pub-footer__copy {
    grid-column: 1 / -1;
  }
}

.pub-footer__tagline {
  margin: var(--kore-space-sm) 0 0;
  font-size: var(--kore-text-caption);
  letter-spacing: 0.12em;
  text-transform: uppercase;
  color: var(--kore-text-muted);
}

.pub-footer__nav {
  display: flex;
  flex-wrap: wrap;
  gap: var(--kore-space-md) var(--kore-space-xl);
}

.pub-footer__nav a {
  color: var(--kore-text-muted);
  text-decoration: none;
  font-size: var(--kore-text-small);
  font-weight: 500;
  transition: color 0.15s;
}

.pub-footer__nav a:hover { color: var(--kore-brand-gold); }

.pub-footer__copy {
  margin: 0;
  font-size: var(--kore-text-caption);
  color: var(--kore-text-muted);
}
</style>
