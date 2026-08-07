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

  test('empty / create opens Depuis CRA wizard with preview step', async ({ page }) => {
    await page.goto('/facturation')
    await expect(page.getByRole('heading', { name: /facturation|invoicing/i })).toBeVisible({
      timeout: 20_000
    })

    const emptyCta = page.getByRole('button', { name: /depuis un cra|from a timesheet/i })
    if (await emptyCta.isVisible().catch(() => false)) {
      await emptyCta.click()
    } else {
      await page.getByRole('button', { name: /nouvelle facture|new invoice/i }).click()
    }

    await expect(page.getByRole('dialog')).toBeVisible({ timeout: 15_000 })
    await expect(
      page.getByRole('heading', { name: /facture depuis cra|invoice from timesheet/i })
    ).toBeVisible()
    await expect(page.getByText(/étape 1|step 1/i)).toBeVisible()

    // Month picker is present; selecting a CRA (if any) leads to preview fields.
    await expect(page.locator('input[type="month"]')).toBeVisible()
    const firstCheck = page.locator('.wizard__check input[type="checkbox"]').first()
    if (await firstCheck.isVisible().catch(() => false)) {
      await firstCheck.check()
      await page.getByRole('button', { name: /continuer|continue/i }).click()
      await expect(page.getByText(/étape 2|step 2/i)).toBeVisible({ timeout: 15_000 })
      // Preview shows either editable fields or blockers — both prove dry-run ran.
      await expect(
        page.locator('.wizard__draft').or(page.getByText(/bloqué|blocked|client|heures|hours/i))
      ).toBeVisible({ timeout: 10_000 })
    }
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
