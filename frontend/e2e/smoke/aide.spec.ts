import { test, expect } from '@playwright/test'
import { expectAppShell, loginAsAdmin } from '../fixtures/auth'

test.describe('aide', () => {
  test.beforeEach(async ({ page }) => {
    await loginAsAdmin(page)
  })

  test('hub and access guide are reachable from nav', async ({ page }) => {
    await page.goto('/aide')
    await expect(page).toHaveURL(/\/aide(?:\/|$|\?)/)
    await expectAppShell(page)
    await expect(page.getByRole('heading', { name: /^aide$|^help$/i })).toBeVisible({
      timeout: 20_000
    })
    await expect(page.getByText(/vos profils|your profiles/i)).toBeVisible()
    await expect(page.getByRole('link', { name: /accès par profil|access by profile/i })).toBeVisible()

    await page.getByRole('link', { name: /accès par profil|access by profile/i }).click()
    await expect(page).toHaveURL(/\/aide\/acces/)
    await expect(page.getByRole('heading', { name: /accès par profil|access by profile/i })).toBeVisible({
      timeout: 20_000
    })
    await expect(page.getByText(/matrice profil|profile × module|profile x module/i)).toBeVisible()
    await expect(page.getByRole('cell', { name: 'Administrateur' }).or(page.getByRole('rowheader', { name: 'Administrateur' }))).toBeVisible()
    await expect(page.getByText(/ADM_admin/)).toBeVisible()
  })
})
