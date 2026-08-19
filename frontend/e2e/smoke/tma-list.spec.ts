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

  test('create form surfaces API errors instead of a silent failure', async ({ page }) => {
    await page.route('**/api/tma/demands', async (route) => {
      if (route.request().method() === 'POST') {
        await route.fulfill({
          status: 500,
          contentType: 'application/json',
          body: JSON.stringify({
            error: { code: 'INTERNAL', message: 'demand create failed' }
          })
        })
        return
      }
      await route.continue()
    })

    await page.goto('/tma')
    await page.getByRole('button', { name: /nouvelle demande|new request/i }).click()
    await page.getByLabel(/sujet|subject/i).fill("Suivi d'équipe")
    const submit = page.getByRole('button', { name: /créer la demande|create request/i })
    await expect(submit).toBeEnabled({ timeout: 15_000 })
    await submit.click()
    await expect(page.getByRole('alert')).toContainText(/impossible de créer|demand create failed|unable to create/i)
  })
})
