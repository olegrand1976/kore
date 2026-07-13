<template>
  <div class="cra-skeleton" aria-hidden="true">
    <div class="cra-skeleton__block cra-skeleton__block--title" />
    <div class="cra-skeleton__row">
      <div class="cra-skeleton__block cra-skeleton__block--kpi" />
      <div class="cra-skeleton__block cra-skeleton__block--kpi" />
      <div class="cra-skeleton__block cra-skeleton__block--kpi" />
    </div>
    <div class="cra-skeleton__tabs">
      <div v-for="n in 4" :key="n" class="cra-skeleton__block cra-skeleton__block--tab" />
    </div>
    <div v-for="n in 3" :key="`day-${n}`" class="cra-skeleton__block cra-skeleton__block--day" />
  </div>
</template>

<style scoped>
.cra-skeleton {
  display: grid;
  gap: var(--kore-space-md);
}

.cra-skeleton__block {
  border-radius: var(--kore-radius-md);
  background: linear-gradient(
    90deg,
    var(--kore-border) 25%,
    var(--kore-bg-elevated) 50%,
    var(--kore-border) 75%
  );
  background-size: 200% 100%;
  animation: cra-skeleton-pulse 1.2s ease-in-out infinite;
}

.cra-skeleton__block--title {
  height: 2.5rem;
}

.cra-skeleton__row {
  display: grid;
  grid-template-columns: repeat(3, 1fr);
  gap: var(--kore-space-sm);
}

.cra-skeleton__block--kpi {
  height: 4rem;
}

.cra-skeleton__tabs {
  display: flex;
  gap: var(--kore-space-sm);
  overflow: hidden;
}

.cra-skeleton__block--tab {
  flex: 0 0 5rem;
  height: 2.25rem;
}

.cra-skeleton__block--day {
  height: 5rem;
}

@keyframes cra-skeleton-pulse {
  0% { background-position: 100% 0; }
  100% { background-position: -100% 0; }
}

@media (max-width: 768px) {
  .cra-skeleton__row {
    grid-template-columns: 1fr;
  }
}
</style>
