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

  test('creates a team, edits it, and attaches a collaborator to it', async ({ page }) => {
    const teamName = `E2E Équipe ${Date.now()}`
    const renamed = `${teamName} renommée`

    await page.goto('/admin/organisation')
    await page.getByRole('tab', { name: /structure/i }).click()

    // Crée une équipe sous la première application du seed.
    await page.getByRole('button', { name: /ajouter une équipe/i }).first().click()
    await page.locator('#org-node-libelle').fill(teamName)
    await page.getByRole('button', { name: /^créer$/i }).click()

    // L'équipe apparaît dans l'arbre avec son compteur de collaborateurs.
    await expect(page.getByText(teamName)).toBeVisible({ timeout: 20_000 })

    // Renomme l'équipe via le bouton Modifier du nœud feuille.
    const teamRow = page.locator('.org-tree__node--leaf', { hasText: teamName }).first()
    await teamRow.getByRole('button', { name: /modifier|edit/i }).click()
    await expect(page.getByRole('heading', { name: /modifier l'équipe|edit team/i })).toBeVisible()
    await page.locator('#org-rename-libelle').fill(renamed)
    await page.getByRole('button', { name: /^enregistrer$|^save$/i }).click()
    await expect(page.getByText(renamed)).toBeVisible({ timeout: 20_000 })

    // Rattache un collaborateur à cette équipe depuis l'écran utilisateurs.
    await page.goto('/admin/users')
    const collabRow = page.locator('tbody tr', { hasText: 'COL_collab' }).first()
    await expect(collabRow).toBeVisible({ timeout: 20_000 })
    await collabRow.getByRole('button', { name: /modifier|edit/i }).click()

    const profileCheckbox = page.locator('label.users-check', { hasText: "Chef d'équipe" }).locator('input[type="checkbox"]')
    await expect(profileCheckbox).toBeEnabled()
    await profileCheckbox.check()

    const equipeCheckbox = page.locator('label.users-check', { hasText: renamed }).locator('input[type="checkbox"]')
    await expect(equipeCheckbox).toBeVisible()
    await equipeCheckbox.check()
    await page.getByRole('button', { name: /^enregistrer$|^save$/i }).click()

    // La colonne Équipe de la liste reflète le rattachement.
    await expect(
      page.locator('tbody tr', { hasText: 'COL_collab' }).first().getByText(new RegExp(renamed))
    ).toBeVisible({ timeout: 20_000 })
    await expect(
      page.locator('tbody tr', { hasText: 'COL_collab' }).first().getByText(/chef d'équipe/i)
    ).toBeVisible({ timeout: 20_000 })
  })

  test('self-edit keeps Administrateur locked and allows other profiles', async ({ page }) => {
    await page.goto('/admin/users')
    const adminRow = page.locator('tbody tr', { hasText: 'ADM_admin' }).first()
    await expect(adminRow).toBeVisible({ timeout: 20_000 })
    await adminRow.getByRole('button', { name: /modifier|edit/i }).click()

    const adminProfile = page
      .locator('label.users-check', { hasText: 'Administrateur' })
      .locator('input[type="checkbox"]')
    const chefProfile = page
      .locator('label.users-check', { hasText: "Chef d'équipe" })
      .locator('input[type="checkbox"]')
    const activeToggle = page.locator('label.users-toggle input[type="checkbox"]')

    await expect(adminProfile).toBeChecked()
    await expect(adminProfile).toBeDisabled()
    await expect(chefProfile).toBeEnabled()
    await expect(activeToggle).toBeDisabled()

    await chefProfile.check()
    await page.getByRole('button', { name: /^enregistrer$|^save$/i }).click()

    await expect(
      page.locator('tbody tr', { hasText: 'ADM_admin' }).first().getByText(/chef d'équipe/i)
    ).toBeVisible({ timeout: 20_000 })
    await expect(
      page.locator('tbody tr', { hasText: 'ADM_admin' }).first().getByText(/administrateur/i)
    ).toBeVisible()
  })

  test('applications admin page can create and deactivate', async ({ page }) => {
    const appName = `E2E App ${Date.now()}`

    await page.goto('/admin/applications')
    await expect(page.getByRole('heading', { name: /applications/i })).toBeVisible({
      timeout: 20_000
    })

    await page.getByRole('button', { name: /nouvelle application|new application/i }).first().click()
    await expect(page.getByRole('heading', { name: /nouvelle application|new application/i })).toBeVisible()

    const shareCheckbox = page.locator('.apps-form__checklist input[type="checkbox"]').first()
    await expect(shareCheckbox).toBeVisible()
    await shareCheckbox.check()

    await page.locator('#app-libelle').fill(appName)
    await page.getByRole('button', { name: /^enregistrer$|^save$/i }).click()
    await expect(page.getByText(appName)).toBeVisible({ timeout: 20_000 })

    const row = page.locator('tbody tr', { hasText: appName }).first()
    page.once('dialog', (dialog) => dialog.accept())
    await row.getByRole('button', { name: /désactiver|deactivate/i }).click()
    await expect(row.getByText(/inactive/i)).toBeVisible({ timeout: 20_000 })
  })
})
