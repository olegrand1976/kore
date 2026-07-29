import { test, expect } from '@playwright/test'
import { ensureFrenchLocale, loginAsAdminViaUI, SEED_ADMIN } from '../fixtures/auth'

test.describe('auth login', () => {
  test('admin can sign in and reach dashboard', async ({ page }) => {
    await ensureFrenchLocale(page)
    await loginAsAdminViaUI(page)
    await expect(page.getByText(/bon retour sur kore|welcome back to kore/i)).toBeVisible({ timeout: 20_000 })
    // sanity: credentials used by the helper stay documented for reviewers
    expect(SEED_ADMIN.login).toBe('ADM_admin')
  })
})
