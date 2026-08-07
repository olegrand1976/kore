import { test, expect } from '@playwright/test'
import { expectListOrEmpty, loginAsAdmin } from '../fixtures/auth'

test.describe('missions list', () => {
  test.beforeEach(async ({ page }) => {
    await loginAsAdmin(page)
  })

  test('missions page loads with list or empty state', async ({ page }) => {
    await page.goto('/missions')
    await expect(page.getByRole('heading', { name: /missions/i })).toBeVisible({ timeout: 20_000 })
    await expectListOrEmpty(page, /aucune mission|no mission/i)
  })

  test('new mission form exposes optional applications field', async ({ page }) => {
    await page.goto('/missions/nouveau')
    await expect(page.getByRole('heading', { name: /créer une mission|create a mission/i })).toBeVisible({
      timeout: 20_000
    })
    await expect(page.getByText(/applications/i).first()).toBeVisible()
  })
})
