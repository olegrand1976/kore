import { expect, type Page } from '@playwright/test'

const baseURL = process.env.PLAYWRIGHT_BASE_URL || 'http://localhost:3001'

export const SEED_ADMIN = {
  login: 'ADM_admin',
  password: 'Admin123!'
} as const

/** Force FR so assertions stay deterministic regardless of browser language. */
export async function ensureFrenchLocale(page: Page) {
  await page.context().addCookies([
    { name: 'kore-locale', value: 'fr', url: baseURL }
  ])
}

/**
 * Authenticate via BFF login (httpOnly cookies on the shared browser context).
 * Relies on useRequestFetch in useAuth so SSR middleware sees the Cookie header.
 */
export async function loginAsAdmin(page: Page) {
  await ensureFrenchLocale(page)
  const res = await page.request.post('/api/auth/login', {
    data: {
      login: SEED_ADMIN.login,
      password: SEED_ADMIN.password
    }
  })
  expect(res.ok(), `login failed: ${res.status()} ${await res.text()}`).toBeTruthy()

  const cookies = await page.context().cookies()
  expect(
    cookies.some((c) => c.name === 'kore_access_token' && c.value.length > 0),
    'kore_access_token missing after login'
  ).toBeTruthy()

  const session = await page.request.get('/api/auth/session')
  expect(session.ok(), `session check failed: ${session.status()}`).toBeTruthy()

  await page.goto('/dashboard')
  await expect(page).toHaveURL(/\/dashboard(?:\/|$|\?)/, { timeout: 30_000 })
}

/** Full UI login — used by the auth smoke spec only. */
export async function loginAsAdminViaUI(page: Page) {
  await ensureFrenchLocale(page)
  await page.goto('/login')
  const login = page.getByLabel(/identifiant|identifier|^login$/i)
  const password = page.getByLabel(/mot de passe|password/i)
  await login.click()
  await login.fill('')
  await login.pressSequentially(SEED_ADMIN.login, { delay: 20 })
  await password.click()
  await password.fill('')
  await password.pressSequentially(SEED_ADMIN.password, { delay: 20 })
  await page.getByRole('button', { name: /se connecter|sign in|log in/i }).click()
  await expect(page).toHaveURL(/\/dashboard(?:\/|$|\?)/, { timeout: 30_000 })
}

/** Assert app shell is visible (desktop sidebar). */
export async function expectAppShell(page: Page) {
  await expect(page.locator('aside.sidebar')).toBeVisible({ timeout: 20_000 })
  await expect(page.locator('nav.sidebar__nav')).toBeVisible()
}

/** Assert list/detail content: table rows or a known empty-state title. */
export async function expectListOrEmpty(page: Page, emptyTitle: RegExp) {
  const table = page.locator('table tbody tr, .app-table tbody tr, [role="table"]')
  const empty = page.getByText(emptyTitle)
  await expect(table.or(empty).first()).toBeVisible({ timeout: 20_000 })
}
