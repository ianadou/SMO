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

test.describe('Create group modal', () => {
  async function openModal(page: Page) {
    await authenticate(page)
    await mockGroups(page, 200, [])
    await page.goto('/groups')
    await page.getByRole('button', { name: 'Créer un groupe' }).click()
    await expect(page.getByRole('dialog')).toBeVisible()
  }

  test('opens when clicking the floating action button', async ({ page }) => {
    await openModal(page)
    await expect(page.getByRole('heading', { name: 'Nouveau groupe' })).toBeVisible()
    await expect(page.getByRole('textbox', { name: 'Nom du groupe' })).toBeVisible()
    await expect(page.getByRole('textbox', { name: 'Webhook Discord (optionnel)' })).toBeVisible()
  })

  test('Esc closes the modal', async ({ page }) => {
    await openModal(page)
    await page.keyboard.press('Escape')
    await expect(page.getByRole('dialog')).not.toBeVisible()
  })

  test('clicking the overlay closes the modal', async ({ page }) => {
    await openModal(page)
    await page.locator('.fixed.inset-0.z-50').click({ position: { x: 5, y: 5 } })
    await expect(page.getByRole('dialog')).not.toBeVisible()
  })

  test('Annuler button closes the modal', async ({ page }) => {
    await openModal(page)
    await page.getByRole('button', { name: 'Annuler' }).click()
    await expect(page.getByRole('dialog')).not.toBeVisible()
  })

  test('submit button is disabled when name is empty', async ({ page }) => {
    await openModal(page)
    await expect(page.getByRole('button', { name: 'Créer le groupe' })).toBeDisabled()
  })

  test('submit button enables when a non-empty name is typed', async ({ page }) => {
    await openModal(page)
    await page.getByRole('textbox', { name: 'Nom du groupe' }).fill('Foot du jeudi')
    await expect(page.getByRole('button', { name: 'Créer le groupe' })).toBeEnabled()
  })

  test('shows length counter and rejects names over 100 chars', async ({ page }) => {
    await openModal(page)
    await page.getByRole('textbox', { name: 'Nom du groupe' }).fill('A'.repeat(101))
    await expect(page.getByText('101/100')).toBeVisible()
    await expect(page.getByText(`Trop long (max 100 caractères)`)).toBeVisible()
    await expect(page.getByRole('button', { name: 'Créer le groupe' })).toBeDisabled()
  })

  test('rejects an invalid Discord webhook URL', async ({ page }) => {
    await openModal(page)
    await page.getByRole('textbox', { name: 'Nom du groupe' }).fill('Foot')
    await page.getByRole('textbox', { name: 'Webhook Discord (optionnel)' }).fill('https://malicious.com/api/webhooks/1/abc')
    await page.getByRole('textbox', { name: 'Webhook Discord (optionnel)' }).blur()
    await expect(page.getByText(/Format attendu/)).toBeVisible()
    await expect(page.getByRole('button', { name: 'Créer le groupe' })).toBeDisabled()
  })

  test('submitting valid input creates the group, closes the modal and refreshes the list', async ({ page }) => {
    await authenticate(page)
    let listCalls = 0
    await page.route('**/api/v1/groups', async (route) => {
      const method = route.request().method()
      if (method === 'POST') {
        await route.fulfill({
          status: 201,
          contentType: 'application/json',
          body: JSON.stringify({
            id: 'g-new',
            name: 'Foot du jeudi',
            organizer_id: 'org-1',
            has_webhook: false,
            created_at: '2026-04-30T10:00:00Z',
          }),
        })
        return
      }
      listCalls++
      const body = listCalls === 1
        ? []
        : [
            {
              id: 'g-new',
              name: 'Foot du jeudi',
              organizer_id: 'org-1',
              has_webhook: false,
              created_at: '2026-04-30T10:00:00Z',
            },
          ]
      await route.fulfill({
        status: 200,
        contentType: 'application/json',
        body: JSON.stringify(body),
      })
    })
    await page.goto('/groups')
    await page.getByRole('button', { name: 'Créer un groupe' }).click()
    await page.getByRole('textbox', { name: 'Nom du groupe' }).fill('Foot du jeudi')
    await page.getByRole('button', { name: 'Créer le groupe' }).click()
    await expect(page.getByRole('dialog')).not.toBeVisible()
    await expect(page.getByText('Foot du jeudi')).toBeVisible()
  })

  test('shows the backend error inline when the API returns 400', async ({ page }) => {
    await authenticate(page)
    await page.route('**/api/v1/groups', async (route) => {
      if (route.request().method() === 'POST') {
        await route.fulfill({
          status: 400,
          contentType: 'application/json',
          body: JSON.stringify({ error: 'invalid name' }),
        })
        return
      }
      await route.fulfill({ status: 200, contentType: 'application/json', body: '[]' })
    })
    await page.goto('/groups')
    await page.getByRole('button', { name: 'Créer un groupe' }).click()
    await page.getByRole('textbox', { name: 'Nom du groupe' }).fill('Foot')
    await page.getByRole('button', { name: 'Créer le groupe' }).click()
    await expect(page.getByRole('alert')).toContainText('invalid name')
    await expect(page.getByRole('dialog')).toBeVisible()
  })

  test('shows a network error message when the request fails', async ({ page }) => {
    await authenticate(page)
    await page.route('**/api/v1/groups', async (route) => {
      if (route.request().method() === 'POST') {
        await route.abort('failed')
        return
      }
      await route.fulfill({ status: 200, contentType: 'application/json', body: '[]' })
    })
    await page.goto('/groups')
    await page.getByRole('button', { name: 'Créer un groupe' }).click()
    await page.getByRole('textbox', { name: 'Nom du groupe' }).fill('Foot')
    await page.getByRole('button', { name: 'Créer le groupe' }).click()
    await expect(page.getByRole('alert')).toContainText(/Connexion au serveur impossible|hors connexion/)
    await expect(page.getByRole('dialog')).toBeVisible()
  })

  test('trims whitespace around the name on blur', async ({ page }) => {
    await openModal(page)
    const input = page.getByRole('textbox', { name: 'Nom du groupe' })
    await input.fill('  Foot  ')
    await input.blur()
    await expect(input).toHaveValue('Foot')
  })
})
