<script setup lang="ts">
export type LogoVariant = 'full' | 'horizontal' | 'emblem' | 'wordmark' | 'hero'
export type LogoSize = 'sm' | 'md' | 'lg' | 'xl'
export type LogoTone = 'auto' | 'light' | 'dark' | 'color'

const props = withDefaults(defineProps<{
  variant?: LogoVariant
  size?: LogoSize
  showTagline?: boolean
  tone?: LogoTone
  alt?: string
}>(), {
  variant: 'horizontal',
  size: 'md',
  showTagline: false,
  tone: 'auto',
  alt: 'Kore'
})

const { theme } = useTheme()

const sizeMap: Record<LogoSize, string> = {
  sm: '120px',
  md: '160px',
  lg: '220px',
  xl: '320px'
}

const emblemSizeMap: Record<LogoSize, string> = {
  sm: '28px',
  md: '36px',
  lg: '48px',
  xl: '64px'
}

const heroSizeMap: Record<LogoSize, string> = {
  sm: '140px',
  md: '220px',
  lg: '300px',
  xl: 'min(420px, 88vw)'
}

const isDark = computed(() => theme.value === 'dark')

const src = computed(() => {
  if (props.variant === 'hero') {
    return '/brand/kore-logo-hero.png'
  }

  if (props.variant === 'emblem') {
    if (props.tone === 'light') return '/brand/kore-emblem-mono-light.svg'
    if (props.tone === 'dark') return '/brand/kore-emblem-mono-dark.svg'
    if (props.tone === 'color') return '/brand/kore-emblem.svg'
    return isDark.value
      ? '/brand/kore-emblem-mono-light.svg'
      : '/brand/kore-emblem-mono-dark.svg'
  }

  if (props.tone === 'auto') {
    const themed: Partial<Record<Exclude<LogoVariant, 'hero' | 'emblem'>, string>> = {
      full: isDark.value ? '/brand/kore-logo-full-dark.svg' : '/brand/kore-logo-full.svg',
      horizontal: isDark.value ? '/brand/kore-logo-horizontal-dark.svg' : '/brand/kore-logo-horizontal.svg',
      wordmark: isDark.value ? '/brand/kore-wordmark-dark.svg' : '/brand/kore-wordmark.svg'
    }
    const themedSrc = themed[props.variant as Exclude<LogoVariant, 'hero' | 'emblem'>]
    if (themedSrc) return themedSrc
  }

  const files: Record<Exclude<LogoVariant, 'hero'>, string> = {
    full: '/brand/kore-logo-full.svg',
    horizontal: '/brand/kore-logo-horizontal.svg',
    emblem: '/brand/kore-emblem.svg',
    wordmark: '/brand/kore-wordmark.svg'
  }
  return files[props.variant as Exclude<LogoVariant, 'hero'>]
})

const width = computed(() => {
  if (props.variant === 'hero') {
    return heroSizeMap[props.size]
  }
  return props.variant === 'emblem' || props.variant === 'wordmark'
    ? emblemSizeMap[props.size]
    : sizeMap[props.size]
})
</script>

<template>
  <div class="kore-logo" :class="[`kore-logo--${variant}`, `kore-logo--${size}`]">
    <img
      :src="src"
      :alt="alt"
      :style="{ width }"
      class="kore-logo__img"
      :class="{ 'kore-logo__img--hero': variant === 'hero' }"
      :fetchpriority="variant === 'hero' ? 'high' : undefined"
      :loading="variant === 'hero' ? 'eager' : 'lazy'"
    />
    <p v-if="showTagline && variant !== 'full'" class="kore-logo__tagline">
      TMA • TIMESHEET • BUDGETING
    </p>
  </div>
</template>

<style scoped>
.kore-logo {
  display: inline-flex;
  flex-direction: column;
  align-items: flex-start;
  gap: var(--kore-space-xs);
}

.kore-logo__img {
  height: auto;
  display: block;
}

.kore-logo__img--hero {
  filter: drop-shadow(0 8px 32px rgba(201, 162, 39, 0.22));
}

.kore-logo__tagline {
  margin: 0;
  font-size: var(--kore-text-caption);
  font-weight: 500;
  letter-spacing: 0.12em;
  text-transform: uppercase;
  color: var(--kore-text-muted);
}
</style>
