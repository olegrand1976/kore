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

  test('structure tab shows the org tree down to teams', async ({ page }) => {
    await page.goto('/admin/organisation')
    await page.getByRole('tab', { name: /structure/i }).click()

    // La société seed et au moins un niveau Application doivent être visibles.
    await expect(page.getByText(/société → site → service/i)).toBeVisible({ timeout: 20_000 })
    await expect(page.getByRole('button', { name: /ajouter un site/i }).first()).toBeVisible()
    await expect(
      page.getByRole('button', { name: /ajouter une équipe/i }).first()
    ).toBeVisible({ timeout: 20_000 })
  })

  test('creates a team and attaches a collaborator to it', async ({ page }) => {
    const teamName = `E2E Équipe ${Date.now()}`

    await page.goto('/admin/organisation')
    await page.getByRole('tab', { name: /structure/i }).click()

    // Crée une équipe sous la première application du seed.
    await page.getByRole('button', { name: /ajouter une équipe/i }).first().click()
    await page.locator('#org-node-libelle').fill(teamName)
    await page.getByRole('button', { name: /^créer$/i }).click()

    // L'équipe apparaît dans l'arbre avec son compteur de collaborateurs.
    await expect(page.getByText(teamName)).toBeVisible({ timeout: 20_000 })

    // Rattache un collaborateur à cette équipe depuis l'écran utilisateurs.
    await page.goto('/admin/users')
    const collabRow = page.locator('tbody tr', { hasText: 'COL_collab' }).first()
    await expect(collabRow).toBeVisible({ timeout: 20_000 })
    await collabRow.getByRole('button', { name: /modifier|edit/i }).click()

    const equipeCheckbox = page.locator('label.users-check', { hasText: teamName }).locator('input[type="checkbox"]')
    await expect(equipeCheckbox).toBeVisible()
    await equipeCheckbox.check()
    await page.getByRole('button', { name: /^enregistrer$|^save$/i }).click()

    // La colonne Équipe de la liste reflète le rattachement.
    await expect(
      page.locator('tbody tr', { hasText: 'COL_collab' }).first().getByText(new RegExp(teamName))
    ).toBeVisible({ timeout: 20_000 })
  })
})
