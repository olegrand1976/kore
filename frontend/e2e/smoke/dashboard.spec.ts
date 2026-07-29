import { test, expect } from '@playwright/test'
import { expectAppShell, loginAsAdmin } from '../fixtures/auth'

test.describe('dashboard', () => {
  test.beforeEach(async ({ page }) => {
    await loginAsAdmin(page)
  })

  test('loads welcome banner and app shell', async ({ page }) => {
    await page.goto('/dashboard')
    await expect(page).toHaveURL(/\/dashboard(?:\/|$|\?)/)
    await expectAppShell(page)
    await expect(page.getByText(/bon retour sur kore|welcome back to kore/i)).toBeVisible({ timeout: 20_000 })
    await expect(page.getByRole('heading', { name: /vue d'ensemble|overview of your activity/i })).toBeVisible()
  })
})
