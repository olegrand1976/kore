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

  test('sites admin page can create and rename', async ({ page }) => {
    const siteName = `E2E Site ${Date.now()}`
    const renamed = `${siteName} renommé`

    await page.goto('/admin/sites')
    await expect(page.getByRole('heading', { name: /^sites$/i })).toBeVisible({ timeout: 20_000 })

    await page.getByRole('button', { name: /nouveau site|new site/i }).first().click()
    await expect(page.getByRole('heading', { name: /nouveau site|new site/i })).toBeVisible()
    await page.locator('#site-libelle').fill(siteName)
    await page.getByRole('button', { name: /^enregistrer$|^save$/i }).click()
    await expect(page.getByText(siteName)).toBeVisible({ timeout: 20_000 })

    const row = page.locator('tbody tr', { hasText: siteName }).first()
    await row.getByRole('button', { name: /modifier|edit/i }).click()
    await expect(page.getByRole('heading', { name: /modifier le site|edit site/i })).toBeVisible()
    await page.locator('#site-libelle').fill(renamed)
    await page.getByRole('button', { name: /^enregistrer$|^save$/i }).click()
    await expect(page.getByText(renamed)).toBeVisible({ timeout: 20_000 })
  })

  test('services and equipes admin pages load', async ({ page }) => {
    await page.goto('/admin/services')
    await expect(page.getByRole('heading', { name: /^services$/i })).toBeVisible({ timeout: 20_000 })
    await expect(page.getByRole('button', { name: /nouveau service|new service/i }).first()).toBeVisible()

    await page.goto('/admin/equipes')
    await expect(page.getByRole('heading', { name: /équipes|teams/i })).toBeVisible({ timeout: 20_000 })
    await expect(page.getByRole('button', { name: /nouvelle équipe|new team/i }).first()).toBeVisible()
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

    // Renomme l'équipe et rattache COL_collab directement depuis le modal Modifier.
    const teamRow = page.locator('.org-tree__node--leaf', { hasText: teamName }).first()
    await teamRow.getByRole('button', { name: /modifier|edit/i }).click()
    await expect(page.getByRole('heading', { name: /modifier l'équipe|edit team/i })).toBeVisible()
    await page.locator('#org-rename-libelle').fill(renamed)
    const memberCheckbox = page
      .locator('.org-tree__check', { hasText: 'COL_collab' })
      .locator('input[type="checkbox"]')
    await expect(memberCheckbox).toBeVisible()
    await memberCheckbox.check()
    await page.getByRole('button', { name: /^enregistrer$|^save$/i }).click()
    await expect(page.getByText(renamed)).toBeVisible({ timeout: 20_000 })

    // Le compteur de l'arbre reflète le rattachement.
    const renamedRow = page.locator('.org-tree__node--leaf', { hasText: renamed }).first()
    await expect(renamedRow.getByText(/1 collaborateur|1 member/i)).toBeVisible({ timeout: 20_000 })

    // La colonne Équipe de la liste utilisateurs reflète aussi le rattachement.
    await page.goto('/admin/users')
    const collabRow = page.locator('tbody tr', { hasText: 'COL_collab' }).first()
    await expect(collabRow).toBeVisible({ timeout: 20_000 })
    await expect(collabRow.getByText(new RegExp(renamed))).toBeVisible({ timeout: 20_000 })
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
    const row = page.locator('tbody tr', { hasText: appName }).first()
    await expect(row).toBeVisible({ timeout: 20_000 })
    page.once('dialog', (dialog) => dialog.accept())
    await row.getByRole('button', { name: /désactiver|deactivate/i }).click()
    await expect(row.getByText(/inactive/i)).toBeVisible({ timeout: 20_000 })
  })

  test('applications admin can merge two applications', async ({ page }) => {
    const stamp = Date.now()
    const appAbsorbed = `E2E Merge Absorbed ${stamp}`
    const appReference = `E2E Merge Ref ${stamp}`

    const createApp = async (libelle: string) => {
      await page.getByRole('button', { name: /nouvelle application|new application/i }).first().click()
      await expect(
        page.getByRole('heading', { name: /nouvelle application|new application/i })
      ).toBeVisible()
      const shareCheckbox = page.locator('.apps-form__checklist input[type="checkbox"]').first()
      await expect(shareCheckbox).toBeVisible()
      await shareCheckbox.check()
      await page.locator('#app-libelle').fill(libelle)
      await page.getByRole('button', { name: /^enregistrer$|^save$/i }).click()
      const row = page.locator('tbody tr', { hasText: libelle }).first()
      await expect(row).toBeVisible({ timeout: 20_000 })
    }

    await page.goto('/admin/applications')
    await expect(page.getByRole('heading', { name: /applications/i })).toBeVisible({
      timeout: 20_000
    })

    await createApp(appAbsorbed)
    await createApp(appReference)

    const absorbedRow = page.locator('tbody tr', { hasText: appAbsorbed }).first()
    const referenceRow = page.locator('tbody tr', { hasText: appReference }).first()
    await absorbedRow.locator('input[type="checkbox"]').check()
    await referenceRow.locator('input[type="checkbox"]').check()

    await page.getByRole('button', { name: /^fusionner$|^merge$/i }).first().click()
    await expect(page.getByRole('heading', { name: /fusionner deux applications|merge two applications/i })).toBeVisible({
      timeout: 20_000
    })
    const mergeModal = page.getByRole('dialog', {
      name: /fusionner deux applications|merge two applications/i
    })
    await mergeModal.getByRole('radio', { name: new RegExp(appReference) }).check()
    await mergeModal.getByRole('button', { name: /^fusionner$|^merge$/i }).click()

    await expect(absorbedRow.getByText(/^inactive$/i)).toBeVisible({ timeout: 20_000 })
    await expect(referenceRow.getByText(/^active$/i)).toBeVisible({ timeout: 20_000 })
  })

  test('applications admin exposes Taiga import and edit modal', async ({ page }) => {
    await page.goto('/admin/applications')
    await expect(page.getByRole('button', { name: /importer depuis taiga|import from taiga/i })).toBeVisible({
      timeout: 20_000
    })

    const editBtn = page.locator('tbody tr').first().getByRole('button', { name: /modifier|edit/i })
    await editBtn.click()
    await expect(
      page.getByRole('heading', { name: /modifier l'application|edit application/i })
    ).toBeVisible({ timeout: 20_000 })
  })
})
