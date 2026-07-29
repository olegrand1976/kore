import { test, expect } from '@playwright/test'
import { loginAsAdmin } from '../fixtures/auth'

test.describe('org admin', () => {
  test.beforeEach(async ({ page }) => {
    await loginAsAdmin(page)
  })

  test('organisation admin page loads', async ({ page }) => {
    await page.goto('/admin/organisation')
    await expect(page.getByRole('heading', { name: /organisation|organization/i })).toBeVisible({
      timeout: 20_000
    })
    await expect(
      page.getByText(/raison sociale|company name|identité visuelle|visual identity/i).first()
    ).toBeVisible({ timeout: 20_000 })
  })
})
