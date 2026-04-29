import { test, expect, type Page } from '@playwright/test'

async function mockRegister(page: Page, status: number, body: unknown) {
  await page.route('**/api/v1/auth/register', async (route) => {
    await route.fulfill({
      status,
      contentType: 'application/json',
      body: JSON.stringify(body),
    })
  })
}

async function mockLoginSuccess(page: Page) {
  await page.route('**/api/v1/auth/login', async (route) => {
    await route.fulfill({
      status: 200,
      contentType: 'application/json',
      body: JSON.stringify({
        token: 'fake-jwt-token',
        organizer: { id: 'org-1', email: 'alex@example.fr', display_name: 'Alex L.', created_at: '2026-01-01T00:00:00Z' },
      }),
    })
  })
}

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

  test('submitting a complete form propagates the API error', async ({ page }) => {
    await mockRegister(page, 409, { error: 'email already exists' })
    await page.goto('/register')
    await page.getByRole('textbox', { name: /Comment te présenter/ }).fill('Alex L.')
    await page.getByRole('textbox', { name: 'Email' }).fill('alex@example.fr')
    await page.getByRole('textbox', { name: 'Mot de passe' }).fill('hunter2hunter2')
    await page.getByRole('checkbox').check({ force: true })

    await page.getByRole('button', { name: 'Créer mon compte' }).click()
    await expect(page.getByRole('alert')).toContainText('email already exists')
  })

  test('successful registration auto-logs-in and redirects to /groups', async ({ page }) => {
    await mockRegister(page, 201, { id: 'org-1', email: 'alex@example.fr', display_name: 'Alex L.', created_at: '2026-01-01T00:00:00Z' })
    await mockLoginSuccess(page)

    await page.goto('/register')
    await page.getByRole('textbox', { name: /Comment te présenter/ }).fill('Alex L.')
    await page.getByRole('textbox', { name: 'Email' }).fill('alex@example.fr')
    await page.getByRole('textbox', { name: 'Mot de passe' }).fill('hunter2hunter2')
    await page.getByRole('checkbox').check({ force: true })

    await page.getByRole('button', { name: 'Créer mon compte' }).click()
    await expect(page).toHaveURL(/\/groups$/)
  })

  test('clicking "Se connecter" goes to /login', async ({ page }) => {
    await page.goto('/register')
    await expect(page.getByRole('link', { name: 'Se connecter' })).toHaveAttribute('href', '/login')
  })
})
