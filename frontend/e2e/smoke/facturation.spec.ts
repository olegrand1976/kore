import { test, expect } from '@playwright/test'
import { expectAppShell, expectListOrEmpty, loginAsAdmin } from '../fixtures/auth'

test.describe('facturation', () => {
  test.beforeEach(async ({ page }) => {
    await loginAsAdmin(page)
  })

  test('nav Facturation visible and list loads', async ({ page }) => {
    await expectAppShell(page)
    const navLink = page.locator('nav.sidebar__nav').getByRole('link', { name: /facturation|invoicing/i })
    await expect(navLink).toBeVisible({ timeout: 20_000 })

    await page.goto('/facturation')
    await expect(page.getByRole('heading', { name: /facturation|invoicing/i })).toBeVisible({
      timeout: 20_000
    })
    await expect(
      page.getByRole('button', { name: /nouvelle facture|new invoice|create invoice/i })
    ).toBeVisible({ timeout: 15_000 })
    await expectListOrEmpty(page, /aucune facture|no invoice/i)
  })

  test('prestations shows Créer factures when invoicing is enabled', async ({ page }) => {
    await page.goto('/prestations')
    await expect(
      page.getByRole('heading', { name: /suivi des prestations|deliverables|services/i })
    ).toBeVisible({ timeout: 20_000 })
    await expect(
      page.getByRole('button', { name: /créer factures|create invoices/i })
    ).toBeVisible({ timeout: 20_000 })
  })

  test('disabling org invoicing hides menu and blocks /facturation', async ({ page }) => {
    await page.goto('/admin/organisation')
    await page.getByRole('tab', { name: /modules/i }).click()

    const checkbox = page
      .locator('label.org-modules__check', { hasText: /facturation client|client invoicing/i })
      .locator('input[type="checkbox"]')
    await expect(checkbox).toBeVisible({ timeout: 20_000 })
    await expect(checkbox).toBeChecked()
    await checkbox.uncheck()
    await page.locator('form.org-modules__form').getByRole('button', { name: /^enregistrer$|^save$/i }).click()
    await expect(page.getByText(/paramètres modules enregistrés|modules (saved|settings)/i)).toBeVisible({
      timeout: 20_000
    })

    await page.goto('/dashboard')
    await expectAppShell(page)
    await expect(
      page.locator('nav.sidebar__nav').getByRole('link', { name: /facturation|invoicing/i })
    ).toHaveCount(0)

    await page.goto('/facturation')
    await expect(page).toHaveURL(/\/dashboard(?:\/|$|\?)/, { timeout: 20_000 })

    // Restore seed default so later smoke specs keep seeing Facturation.
    await page.goto('/admin/organisation')
    await page.getByRole('tab', { name: /modules/i }).click()
    await expect(checkbox).toBeVisible({ timeout: 20_000 })
    await checkbox.check()
    await page.locator('form.org-modules__form').getByRole('button', { name: /^enregistrer$|^save$/i }).click()
    await expect(page.getByText(/paramètres modules enregistrés|modules (saved|settings)/i)).toBeVisible({
      timeout: 20_000
    })
  })
})
