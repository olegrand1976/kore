import { apiAuthHeaders, apiBase } from '../../utils/auth'

type BackendPrefs = {
  lastSeenVersion: string | null
  autoShowEnabled: boolean
}

type GitHubCompareCommit = {
  sha: string
  html_url?: string
  commit?: {
    message?: string
    author?: { name?: string; date?: string }
  }
  author?: { login?: string }
}

function currentVersion(): string {
  return process.env.KORE_VERSION || process.env.NUXT_PUBLIC_KORE_VERSION || process.env.APP_VERSION || '0.0.0'
}

function repoFromEnv(): string {
  const repo = process.env.KORE_GITHUB_REPO || process.env.GITHUB_REPOSITORY
  if (!repo) {
    throw createError({ statusCode: 500, statusMessage: 'Missing KORE_GITHUB_REPO' })
  }
  return repo
}

function tokenFromEnv(): string | null {
  return process.env.KORE_GITHUB_TOKEN || process.env.GITHUB_TOKEN || null
}

function normalizeTag(v: string): string {
  const trimmed = v.trim()
  if (!trimmed) return trimmed
  return trimmed.startsWith('v') ? trimmed : `v${trimmed}`
}

function localeFromEvent(event: Parameters<typeof defineEventHandler>[0] extends never ? never : import('h3').H3Event): string {
  const cookie = getCookie(event, 'kore-locale')
  if (cookie === 'en') return 'en-US'
  if (cookie === 'fr') return 'fr-FR'
  const header = getRequestHeader(event, 'accept-language') || ''
  return header.toLowerCase().startsWith('fr') ? 'fr-FR' : 'en-US'
}

type MonthItem = {
  sha: string
  shortSha: string
  message: string
  authorName?: string
  date: string
  htmlUrl?: string
}

type MonthGroup = { key: string; label: string; items: MonthItem[] }

export default defineEventHandler(async (event) => {
  const headers = apiAuthHeaders(event)
  if (!headers.Authorization) {
    throw createError({ statusCode: 401, statusMessage: 'Unauthorized' })
  }

  const version = currentVersion()
  const repo = repoFromEnv()
  const token = tokenFromEnv()

  let prefs: BackendPrefs = { lastSeenVersion: null, autoShowEnabled: true }
  try {
    prefs = await $fetch<BackendPrefs>(`${apiBase()}/api/v1/users/me/release-notes`, { headers })
  } catch {
    // ok
  }

  const base = prefs.lastSeenVersion
  const head = version

  if (!base || base === head) {
    return { months: [], defaultMonthKey: '' }
  }

  const baseTag = normalizeTag(base)
  const headTag = normalizeTag(head)

  const ghHeaders: Record<string, string> = {
    Accept: 'application/vnd.github+json'
  }
  if (token) ghHeaders.Authorization = `Bearer ${token}`

  const compare = await $fetch<{ commits?: GitHubCompareCommit[] }>(
    `https://api.github.com/repos/${repo}/compare/${encodeURIComponent(baseTag)}...${encodeURIComponent(headTag)}`,
    { headers: ghHeaders }
  )

  const commits = Array.isArray(compare?.commits) ? compare.commits : []
  const loc = localeFromEvent(event)
  const fmtMonth = new Intl.DateTimeFormat(loc, { month: 'long', year: 'numeric' })

  const byMonth = new Map<string, MonthGroup>()
  for (const c of commits) {
    const date = c.commit?.author?.date || new Date().toISOString()
    const d = new Date(date)
    const key = `${d.getUTCFullYear()}-${String(d.getUTCMonth() + 1).padStart(2, '0')}`
    const label = fmtMonth.format(d)
    const existing = byMonth.get(key) || { key, label, items: [] }
    const msg = (c.commit?.message || '').split('\n')[0].trim()
    existing.items.push({
      sha: c.sha,
      shortSha: c.sha.slice(0, 7),
      message: msg || c.sha.slice(0, 7),
      authorName: c.commit?.author?.name || c.author?.login,
      date,
      htmlUrl: c.html_url
    })
    byMonth.set(key, existing)
  }

  const months = Array.from(byMonth.values()).sort((a, b) => (a.key < b.key ? 1 : -1))
  const defaultMonthKey = months[0]?.key || ''
  return { months, defaultMonthKey }
})

