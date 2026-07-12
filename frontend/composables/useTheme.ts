export type ThemeMode = 'dark' | 'light'

const STORAGE_KEY = 'kore-theme'

function resolveSystemTheme(): ThemeMode {
  if (import.meta.client && window.matchMedia('(prefers-color-scheme: light)').matches) {
    return 'light'
  }
  return 'dark'
}

function readStoredTheme(): ThemeMode | null {
  if (!import.meta.client) {
    return null
  }
  const stored = localStorage.getItem(STORAGE_KEY)
  if (stored === 'dark' || stored === 'light') {
    return stored
  }
  return null
}

function applyTheme(theme: ThemeMode) {
  if (!import.meta.client) {
    return
  }
  document.documentElement.setAttribute('data-theme', theme)
  document.documentElement.style.colorScheme = theme
}

export function useTheme() {
  const theme = useState<ThemeMode>('kore-theme', () => 'dark')

  const setTheme = (next: ThemeMode) => {
    theme.value = next
    if (import.meta.client) {
      localStorage.setItem(STORAGE_KEY, next)
      applyTheme(next)
    }
  }

  const toggleTheme = () => {
    setTheme(theme.value === 'dark' ? 'light' : 'dark')
  }

  const initTheme = () => {
    const resolved = readStoredTheme() ?? 'dark'
    theme.value = resolved
    applyTheme(resolved)
  }

  if (import.meta.client) {
    onMounted(initTheme)
  }

  return { theme, setTheme, toggleTheme, initTheme }
}

export { applyTheme, readStoredTheme, resolveSystemTheme, STORAGE_KEY }
