<script setup lang="ts">
export type LogoSize = 'sm' | 'md' | 'lg'

const props = withDefaults(defineProps<{
  logoUrl?: string | null
  alt?: string
  size?: LogoSize
  /** Soft light plate — integrates PNG logos with opaque white backgrounds. */
  framed?: boolean
  fallback?: 'kore-emblem' | 'kore-horizontal'
}>(), {
  logoUrl: null,
  alt: 'Logo société',
  size: 'sm',
  framed: true,
  fallback: 'kore-emblem'
})

const emit = defineEmits<{
  error: []
}>()

const maxHeightMap: Record<LogoSize, string> = {
  sm: '32px',
  md: '40px',
  lg: '52px'
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
  emit('error')
}
</script>

<template>
  <div
    class="tenant-logo"
    :class="{ 'tenant-logo--framed': framed && showTenant }"
  >
    <img
      v-if="showTenant"
      :src="logoUrl ?? undefined"
      :alt="alt"
      class="tenant-logo__img"
      :style="{ maxHeight: maxHeightMap[size] }"
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
  justify-content: center;
  max-width: 100%;
}

.tenant-logo--framed {
  width: 100%;
  padding: var(--kore-space-sm) var(--kore-space-md);
  border-radius: var(--kore-radius-md);
  border: 1px solid var(--kore-border);
  background: var(--kore-bg-elevated);
  box-shadow: var(--kore-shadow-sm);
}

:global([data-theme='dark']) .tenant-logo--framed {
  /* Light canvas for tenant PNGs that assume a white background */
  background: color-mix(in srgb, var(--kore-text) 94%, var(--kore-bg-elevated));
  border-color: color-mix(in srgb, var(--kore-border) 55%, transparent);
  box-shadow: none;
}

.tenant-logo__img {
  display: block;
  width: auto;
  max-width: 100%;
  height: auto;
  object-fit: contain;
}

.tenant-logo--framed .tenant-logo__img {
  width: 100%;
}
</style>
