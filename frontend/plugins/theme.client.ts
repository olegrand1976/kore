import { applyTheme, readStoredTheme } from '~/composables/useTheme'

export default defineNuxtPlugin(() => {
  applyTheme(readStoredTheme() ?? 'dark')
})
