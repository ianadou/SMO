import { test, expect } from '@playwright/test'

test.describe('Register page', () => {
  test('renders wordmark, title, all four fields and a disabled CTA', async ({ page }) => {
    await page.goto('/register')
    await expect(page.getByLabel('SMO').first()).toBeVisible()
    await expect(page.getByRole('heading', { name: 'Créer un compte' })).toBeVisible()
    await expect(page.getByRole('textbox', { name: /Comment te présenter/ })).toBeVisible()
    await expect(page.getByRole('textbox', { name: 'Email' })).toBeVisible()
    await expect(page.getByRole('textbox', { name: 'Mot de passe' })).toBeVisible()
    await expect(page.getByRole('button', { name: 'Compléter le formulaire' })).toBeDisabled()
  })

  test('the CTA enables only when every field is valid AND the CGU box is checked', async ({ page }) => {
    await page.goto('/register')
    await page.getByRole('textbox', { name: /Comment te présenter/ }).fill('Alex L.')
    await page.getByRole('textbox', { name: 'Email' }).fill('alex@example.fr')
    await page.getByRole('textbox', { name: 'Mot de passe' }).fill('hunter2hunter2')
    await expect(page.getByRole('button', { name: 'Compléter le formulaire' })).toBeDisabled()
    await page.getByRole('checkbox').check({ force: true })
    await expect(page.getByRole('button', { name: 'Créer mon compte' })).toBeEnabled()
  })

  test('strength meter reacts as the password gets longer', async ({ page }) => {
    await page.goto('/register')
    const password = page.getByRole('textbox', { name: 'Mot de passe' })

    await password.fill('a')
    await expect(page.getByText('Faible')).toBeVisible()

    await password.fill('aaaaaaaa')
    await expect(page.getByText('Moyen')).toBeVisible()

    await password.fill('Aaaaaaaa1')
    await expect(page.getByText('Fort')).toBeVisible()

    await password.fill('Aaaaaaaa1234')
    await expect(page.getByText('Très fort')).toBeVisible()
  })

  test('submitting a complete form shows loading then the inline error stub', async ({ page }) => {
    await page.goto('/register')
    await page.getByRole('textbox', { name: /Comment te présenter/ }).fill('Alex L.')
    await page.getByRole('textbox', { name: 'Email' }).fill('alex@example.fr')
    await page.getByRole('textbox', { name: 'Mot de passe' }).fill('hunter2hunter2')
    await page.getByRole('checkbox').check({ force: true })

    await page.getByRole('button', { name: 'Créer mon compte' }).click()
    await expect(page.getByRole('button', { name: /Création…/ })).toBeVisible()
    await expect(page.getByRole('alert')).toContainText('Un compte existe déjà')
  })

  test('clicking "Se connecter" goes to /login', async ({ page }) => {
    await page.goto('/register')
    await expect(page.getByRole('link', { name: 'Se connecter' })).toHaveAttribute('href', '/login')
  })
})
