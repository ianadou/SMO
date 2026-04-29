import { test, expect } from '@playwright/test'

test.describe('Login page', () => {
  test('root path redirects to /login', async ({ page }) => {
    await page.goto('/')
    await expect(page).toHaveURL(/\/login$/)
  })

  test('renders wordmark, title and form', async ({ page }) => {
    await page.goto('/login')
    await expect(page.getByLabel('SMO').first()).toBeVisible()
    await expect(page.getByRole('heading', { name: 'Connexion organisateur' })).toBeVisible()
    await expect(page.getByRole('textbox', { name: 'Email' })).toBeVisible()
    await expect(page.getByRole('textbox', { name: 'Mot de passe' })).toBeVisible()
    await expect(page.getByRole('button', { name: 'Se connecter' })).toBeVisible()
  })

  test('password field starts as type="password" and toggles to "text"', async ({ page }) => {
    await page.goto('/login')
    const password = page.getByRole('textbox', { name: 'Mot de passe' })
    await expect(password).toHaveAttribute('type', 'password')

    await page.getByRole('button', { name: 'Afficher le mot de passe' }).click()
    await expect(password).toHaveAttribute('type', 'text')

    await page.getByRole('button', { name: 'Masquer le mot de passe' }).click()
    await expect(password).toHaveAttribute('type', 'password')
  })

  test('submitting valid input shows the loading state then the inline error', async ({ page }) => {
    await page.goto('/login')
    await page.getByRole('textbox', { name: 'Email' }).fill('alex@example.fr')
    await page.getByRole('textbox', { name: 'Mot de passe' }).fill('hunter2hunter2')

    await page.getByRole('button', { name: 'Se connecter' }).click()
    await expect(page.getByRole('button', { name: /Connexion…/ })).toBeVisible()
    await expect(page.getByRole('alert')).toContainText('Identifiants incorrects.')
  })

  test('submitting empty input does NOT trigger the loading state', async ({ page }) => {
    await page.goto('/login')
    await page.getByRole('button', { name: 'Se connecter' }).click()
    await expect(page.getByRole('button', { name: 'Se connecter' })).toBeVisible()
    await expect(page.getByRole('alert')).toHaveCount(0)
  })

  test('signup link points to /register', async ({ page }) => {
    await page.goto('/login')
    const link = page.getByRole('link', { name: "S'inscrire" })
    await expect(link).toHaveAttribute('href', '/register')
  })
})
