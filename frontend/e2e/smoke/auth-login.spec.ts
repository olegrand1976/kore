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

  test('forgot password link opens reset page', async ({ page }) => {
    await ensureFrenchLocale(page)
    await page.goto('/login')
    await page.getByRole('link', { name: /mot de passe oublié|forgot password/i }).click()
    await expect(page).toHaveURL(/\/reset-password/)
    await expect(page.getByRole('heading', { name: /mot de passe oublié|forgot password/i })).toBeVisible()
    await page.getByLabel(/email/i).fill('unknown@example.com')
    await page.getByRole('button', { name: /envoyer le lien|send reset link/i }).click()
    await expect(page.getByText(/si cet email est connu|if this email is known/i)).toBeVisible({ timeout: 10_000 })
  })
})
