import { test, expect } from '@playwright/test'
import { loginAsAdmin } from '../fixtures/auth'

test.describe('Agile projects', () => {
  test.beforeEach(async ({ page }) => {
    await loginAsAdmin(page)
  })

  test('projects hub lists seed agile application', async ({ page }) => {
    await page.goto('/projets')
    await expect(
      page.getByRole('heading', { name: /projets agile|agile projects/i })
    ).toBeVisible({ timeout: 20_000 })
    await expect(page.getByText(/portail client acme|scrum|kanban/i).first()).toBeVisible()
  })
})
