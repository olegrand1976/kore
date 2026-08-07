<script setup lang="ts">
export type LogoSize = 'sm' | 'md' | 'lg'

const props = withDefaults(defineProps<{
  logoUrl?: string | null
  alt?: string
  size?: LogoSize
  fallback?: 'kore-emblem' | 'kore-horizontal'
}>(), {
  logoUrl: null,
  alt: 'Logo société',
  size: 'sm',
  fallback: 'kore-emblem'
})

const sizeMap: Record<LogoSize, string> = {
  sm: '28px',
  md: '36px',
  lg: '48px'
}

const loadFailed = ref(false)

watch(
  () => props.logoUrl,
  () => {
    loadFailed.value = false
  }
)

const showTenant = computed(() => !!props.logoUrl && !loadFailed.value)

const onLogoError = () => {
  loadFailed.value = true
}
</script>

<template>
  <div class="tenant-logo">
    <img
      v-if="showTenant"
      :src="logoUrl ?? undefined"
      :alt="alt"
      class="tenant-logo__img"
      :style="{ height: sizeMap[size] }"
      @error="onLogoError"
    />
    <KoreLogo
      v-else
      :variant="fallback === 'kore-horizontal' ? 'horizontal' : 'emblem'"
      :size="size"
      tone="auto"
      :alt="'Kore'"
    />
  </div>
</template>

<style scoped>
.tenant-logo {
  display: inline-flex;
  align-items: center;
}

.tenant-logo__img {
  width: auto;
  object-fit: contain;
}
</style>
