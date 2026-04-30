import { test, expect, type Page } from '@playwright/test'

const SAMPLE_ORGANIZER = {
  id: 'org-1',
  email: 'alex@example.fr',
  display_name: 'Alex',
  created_at: '2026-01-01T00:00:00Z',
}

async function authenticate(page: Page) {
  await page.goto('/login')
  await page.evaluate(
    ([token, organizer]) => {
      window.localStorage.setItem('smo.auth.token', JSON.stringify(token))
      window.localStorage.setItem('smo.auth.organizer', JSON.stringify(organizer))
    },
    ['fake-jwt-token', SAMPLE_ORGANIZER] as const,
  )
}

async function mockGroups(page: Page, status: number, body: unknown) {
  await page.route('**/api/v1/groups', async (route) => {
    await route.fulfill({
      status,
      contentType: 'application/json',
      body: JSON.stringify(body),
    })
  })
}

test.describe('Groups page', () => {
  test('redirects to /login when unauthenticated', async ({ page }) => {
    await page.goto('/groups')
    await expect(page).toHaveURL(/\/login$/)
  })

  test('renders empty state when the organizer has no groups', async ({ page }) => {
    await authenticate(page)
    await mockGroups(page, 200, [])
    await page.goto('/groups')
    await expect(page.getByRole('heading', { name: 'Pas encore de groupe' })).toBeVisible()
  })

  test('renders one card per group', async ({ page }) => {
    await authenticate(page)
    await mockGroups(page, 200, [
      {
        id: 'g-1',
        name: 'Foot du jeudi',
        organizer_id: 'org-1',
        has_webhook: true,
        created_at: '2026-04-01T00:00:00Z',
      },
      {
        id: 'g-2',
        name: 'Foot du mardi',
        organizer_id: 'org-1',
        has_webhook: false,
        created_at: '2026-04-15T00:00:00Z',
      },
    ])
    await page.goto('/groups')
    await expect(page.getByText('Foot du jeudi')).toBeVisible()
    await expect(page.getByText('Foot du mardi')).toBeVisible()
    await expect(page.getByText('Discord')).toHaveCount(1)
  })

  test('shows the welcome line with the organizer display_name', async ({ page }) => {
    await authenticate(page)
    await mockGroups(page, 200, [])
    await page.goto('/groups')
    await expect(page.getByText('Salut Alex, voici tes groupes.')).toBeVisible()
  })

  test('logs out and redirects when the API returns 401', async ({ page }) => {
    await authenticate(page)
    await mockGroups(page, 401, { error: 'invalid token' })
    await page.goto('/groups')
    await expect(page).toHaveURL(/\/login$/)
    const stored = await page.evaluate(() => window.localStorage.getItem('smo.auth.token'))
    expect(stored).toBeNull()
  })

  test('shows an error and a retry button when the API fails with 500', async ({ page }) => {
    await authenticate(page)
    await mockGroups(page, 500, { error: 'something went wrong' })
    await page.goto('/groups')
    await expect(page.getByRole('alert')).toContainText('something went wrong')
    await expect(page.getByRole('button', { name: 'Réessayer' })).toBeVisible()
  })

  test('logout button clears the session and redirects to /login', async ({ page }) => {
    await authenticate(page)
    await mockGroups(page, 200, [])
    await page.goto('/groups')
    await page.getByRole('button', { name: 'Se déconnecter' }).click()
    await expect(page).toHaveURL(/\/login$/)
    const stored = await page.evaluate(() => window.localStorage.getItem('smo.auth.token'))
    expect(stored).toBeNull()
  })
})
