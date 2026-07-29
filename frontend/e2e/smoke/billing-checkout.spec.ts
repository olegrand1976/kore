import { test, expect } from '@playwright/test'
import { loginAsAdmin } from '../fixtures/auth'

test.describe('billing checkout', () => {
  test.beforeEach(async ({ page }) => {
    await loginAsAdmin(page)
  })

  test('billing subscription page is reachable', async ({ page }) => {
    await page.goto('/billing/abonnement')
    await expect(page.getByRole('heading', { name: /abonnement|subscription/i })).toBeVisible({
      timeout: 20_000
    })
    await expect(
      page.getByText(/statut|status|portail client|customer portal|aucun abonnement|no (active )?subscription/i).first()
    ).toBeVisible({ timeout: 20_000 })
  })
})
