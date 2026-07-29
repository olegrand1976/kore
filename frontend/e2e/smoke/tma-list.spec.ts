import { test, expect } from '@playwright/test'
import { expectListOrEmpty, loginAsAdmin } from '../fixtures/auth'

test.describe('TMA list', () => {
  test.beforeEach(async ({ page }) => {
    await loginAsAdmin(page)
  })

  test('TMA page loads with list or empty state', async ({ page }) => {
    await page.goto('/tma')
    await expect(
      page.getByRole('heading', { name: /évolutions & incidents|evolutions & incidents/i })
    ).toBeVisible({ timeout: 20_000 })
    await expectListOrEmpty(page, /aucune demande|no request/i)
  })
})
