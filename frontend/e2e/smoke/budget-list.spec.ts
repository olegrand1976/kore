import { test, expect } from '@playwright/test'
import { expectListOrEmpty, loginAsAdmin } from '../fixtures/auth'

test.describe('budget list', () => {
  test.beforeEach(async ({ page }) => {
    await loginAsAdmin(page)
  })

  test('budget page loads with list or empty state', async ({ page }) => {
    await page.goto('/budget')
    await expect(page.getByRole('heading', { name: /^budgets?$/i })).toBeVisible({ timeout: 20_000 })
    await expectListOrEmpty(page, /aucun budget|no budget/i)
  })
})
