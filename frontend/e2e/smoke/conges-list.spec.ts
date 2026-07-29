import { test, expect } from '@playwright/test'
import { expectListOrEmpty, loginAsAdmin } from '../fixtures/auth'

test.describe('congés list', () => {
  test.beforeEach(async ({ page }) => {
    await loginAsAdmin(page)
  })

  test('congés page loads KPIs and leave list or empty state', async ({ page }) => {
    await page.goto('/conges')
    await expect(page.getByText(/répartition par statut|breakdown by status/i)).toBeVisible({ timeout: 20_000 })
    await expect(page.getByText(/^demandes$|^requests$/i).first()).toBeVisible()
    await expectListOrEmpty(page, /aucune demande|no request|no leave|aucun résultat|no results/i)
  })
})
