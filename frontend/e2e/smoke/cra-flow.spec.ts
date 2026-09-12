import { test, expect } from '@playwright/test'
import { expectListOrEmpty, loginAsAdmin } from '../fixtures/auth'

test.describe('CRA flow', () => {
  test.beforeEach(async ({ page }) => {
    await loginAsAdmin(page)
  })

  test('CRA list loads and opens current month', async ({ page }) => {
    await page.goto('/cra')
    await expect(
      page.getByRole('heading', { name: /comptes-rendus d'activité|activity reports|timesheets/i })
    ).toBeVisible({ timeout: 20_000 })

    // Prefer the page-header CTA: empty state can render a second identical label.
    const openMonth = page
      .locator('.app-page-header')
      .getByRole('button', { name: /cra du mois en cours|current month timesheet/i })
    await expect(openMonth).toBeVisible({ timeout: 15_000 })
    await openMonth.click()
    await expect(page).toHaveURL(/\/cra\/[0-9a-f-]{8,}/i, { timeout: 20_000 })
    await expect(page.locator('.app-page-header__title')).toBeVisible({ timeout: 20_000 })
    await expect(
      page.getByRole('heading', { name: /informations de prestation|service delivery information/i })
    ).toBeVisible({ timeout: 15_000 })
  })

  test('CRA list can open another past period via modal', async ({ page }) => {
    await page.goto('/cra')
    await expect(
      page.getByRole('heading', { name: /comptes-rendus d'activité|activity reports|timesheets/i })
    ).toBeVisible({ timeout: 20_000 })

    const otherPeriod = page
      .locator('.app-page-header')
      .getByRole('button', { name: /autre période|other period/i })
    await expect(otherPeriod).toBeVisible({ timeout: 15_000 })
    await otherPeriod.click()

    const dialog = page.getByRole('dialog')
    await expect(dialog).toBeVisible()
    await expect(dialog.getByLabel(/mois|month/i)).toBeVisible()
    await expect(dialog.getByLabel(/année|year/i)).toBeVisible()
    await dialog.getByRole('button', { name: /ouvrir|open/i }).click()
    await expect(page).toHaveURL(/\/cra\/[0-9a-f-]{8,}/i, { timeout: 20_000 })
    await expect(
      page.getByRole('status').filter({
        hasText: /cra créé|timesheet created|existait déjà|already existed/i
      })
    ).toBeVisible({ timeout: 15_000 })
  })

  test('CRA list shows table or empty state', async ({ page }) => {
    await page.goto('/cra')
    await expect(
      page.getByRole('heading', { name: /comptes-rendus d'activité|activity reports|timesheets/i })
    ).toBeVisible({ timeout: 20_000 })
    await expectListOrEmpty(page, /aucun cra|no timesheet|no activity report/i)
  })
})
