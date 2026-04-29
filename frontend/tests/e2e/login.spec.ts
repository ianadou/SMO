import { test, expect, type Page } from '@playwright/test'

async function mockLogin(page: Page, status: number, body: unknown) {
  await page.route('**/api/v1/auth/login', async (route) => {
    await route.fulfill({
      status,
      contentType: 'application/json',
      body: JSON.stringify(body),
    })
  })
}

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

  test('submitting valid input with wrong credentials shows the API error', async ({ page }) => {
    await mockLogin(page, 401, { error: 'invalid credentials' })
    await page.goto('/login')
    await page.getByRole('textbox', { name: 'Email' }).fill('alex@example.fr')
    await page.getByRole('textbox', { name: 'Mot de passe' }).fill('wrongpassword')

    await page.getByRole('button', { name: 'Se connecter' }).click()
    await expect(page.getByRole('alert')).toContainText('invalid credentials')
  })

  test('successful login redirects to /groups', async ({ page }) => {
    await mockLogin(page, 200, {
      token: 'fake-jwt-token',
      organizer: { id: 'org-1', email: 'alex@example.fr', display_name: 'Alex', created_at: '2026-01-01T00:00:00Z' },
    })
    await page.goto('/login')
    await page.getByRole('textbox', { name: 'Email' }).fill('alex@example.fr')
    await page.getByRole('textbox', { name: 'Mot de passe' }).fill('rightpassword')
    await page.getByRole('button', { name: 'Se connecter' }).click()
    await expect(page).toHaveURL(/\/groups$/)
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
