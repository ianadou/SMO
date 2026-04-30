import { test, expect, type Page } from '@playwright/test'

const SAMPLE_ORGANIZER = {
  id: 'org-1',
  email: 'alex@example.fr',
  display_name: 'Alex',
  created_at: '2026-01-01T00:00:00Z',
}

const SAMPLE_GROUP = {
  id: 'g-1',
  name: 'Foot du jeudi',
  organizer_id: 'org-1',
  has_webhook: true,
  created_at: '2026-04-01T00:00:00Z',
}

const SAMPLE_MATCH = {
  id: 'm-1',
  group_id: 'g-1',
  title: 'Match d\'ouverture',
  venue: 'Stade municipal',
  scheduled_at: '2026-06-01T18:00:00Z',
  status: 'open' as const,
  created_at: '2026-04-15T10:00:00Z',
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

async function mockGroupAndMatches(page: Page, options: {
  groupStatus?: number
  groupBody?: unknown
  matchesStatus?: number
  matchesBody?: unknown
}) {
  await page.route(url => url.pathname === '/api/v1/groups/g-1/matches', async (route) => {
    await route.fulfill({
      status: options.matchesStatus ?? 200,
      contentType: 'application/json',
      body: JSON.stringify(options.matchesBody ?? []),
    })
  })
  await page.route(url => url.pathname === '/api/v1/groups/g-1', async (route) => {
    await route.fulfill({
      status: options.groupStatus ?? 200,
      contentType: 'application/json',
      body: JSON.stringify(options.groupBody ?? SAMPLE_GROUP),
    })
  })
}

test.describe('Group detail page', () => {
  test('redirects to /login when unauthenticated', async ({ page }) => {
    await page.goto('/groups/g-1')
    await expect(page).toHaveURL(/\/login$/)
  })

  test('renders the group name and the empty matches state', async ({ page }) => {
    await authenticate(page)
    await mockGroupAndMatches(page, {})
    await page.goto('/groups/g-1')
    await expect(page.getByRole('heading', { name: 'Foot du jeudi' })).toBeVisible()
    await expect(page.getByText('Pas encore de match')).toBeVisible()
  })

  test('renders the Discord badge when the group has a webhook', async ({ page }) => {
    await authenticate(page)
    await mockGroupAndMatches(page, {})
    await page.goto('/groups/g-1')
    await expect(page.getByText('Notifications Discord activées')).toBeVisible()
  })

  test('renders one card per match', async ({ page }) => {
    await authenticate(page)
    await mockGroupAndMatches(page, {
      matchesBody: [
        SAMPLE_MATCH,
        { ...SAMPLE_MATCH, id: 'm-2', title: 'Tournoi du dimanche', status: 'draft' },
      ],
    })
    await page.goto('/groups/g-1')
    await expect(page.getByText('Match d\'ouverture')).toBeVisible()
    await expect(page.getByText('Tournoi du dimanche')).toBeVisible()
    await expect(page.getByText('Brouillon', { exact: true })).toBeVisible()
    await expect(page.getByText('Ouvert', { exact: true })).toBeVisible()
  })

  test('renders the not-found state when the group does not exist', async ({ page }) => {
    await authenticate(page)
    await mockGroupAndMatches(page, {
      groupStatus: 404,
      groupBody: { error: 'group not found' },
    })
    await page.goto('/groups/g-1')
    await expect(page.getByRole('heading', { name: "Ce groupe n'existe pas" })).toBeVisible()
  })

  test('logs out and redirects when the API returns 401', async ({ page }) => {
    await authenticate(page)
    await mockGroupAndMatches(page, {
      groupStatus: 401,
      groupBody: { error: 'invalid token' },
    })
    await page.goto('/groups/g-1')
    await expect(page).toHaveURL(/\/login$/)
    const stored = await page.evaluate(() => window.localStorage.getItem('smo.auth.token'))
    expect(stored).toBeNull()
  })

  test('back button navigates to /groups', async ({ page }) => {
    await authenticate(page)
    await mockGroupAndMatches(page, {})
    await page.route('**/api/v1/groups', async (route) => {
      await route.fulfill({ status: 200, contentType: 'application/json', body: '[]' })
    })
    await page.goto('/groups/g-1')
    await page.getByRole('link', { name: 'Retour aux groupes' }).click()
    await expect(page).toHaveURL(/\/groups$/)
  })

  test('GroupCard on /groups navigates to the detail page', async ({ page }) => {
    await authenticate(page)
    await page.route('**/api/v1/groups', async (route) => {
      await route.fulfill({
        status: 200,
        contentType: 'application/json',
        body: JSON.stringify([SAMPLE_GROUP]),
      })
    })
    await mockGroupAndMatches(page, {})
    await page.goto('/groups')
    await page.getByRole('link', { name: /Foot du jeudi/ }).click()
    await expect(page).toHaveURL(/\/groups\/g-1$/)
    await expect(page.getByRole('heading', { name: 'Foot du jeudi' })).toBeVisible()
  })
})

test.describe('Create match modal', () => {
  async function openModal(page: Page) {
    await authenticate(page)
    await mockGroupAndMatches(page, {})
    await page.goto('/groups/g-1')
    await page.getByRole('button', { name: 'Créer un match' }).click()
    await expect(page.getByRole('dialog')).toBeVisible()
  }

  test('opens when clicking the FAB', async ({ page }) => {
    await openModal(page)
    await expect(page.getByRole('heading', { name: 'Nouveau match' })).toBeVisible()
    await expect(page.getByRole('textbox', { name: 'Titre' })).toBeVisible()
    await expect(page.getByRole('textbox', { name: 'Lieu' })).toBeVisible()
  })

  test('Esc closes the modal', async ({ page }) => {
    await openModal(page)
    await page.keyboard.press('Escape')
    await expect(page.getByRole('dialog')).not.toBeVisible()
  })

  test('submit button is disabled until title, venue, and date are set', async ({ page }) => {
    await openModal(page)
    const submit = page.getByRole('button', { name: 'Créer le match' })
    await expect(submit).toBeDisabled()
    await page.getByRole('textbox', { name: 'Titre' }).fill('Match du jeudi')
    await expect(submit).toBeDisabled()
    await page.getByRole('textbox', { name: 'Lieu' }).fill('Stade')
    await expect(submit).toBeDisabled()
    await page.locator('#match-scheduled').fill('2099-01-01T18:00')
    await expect(submit).toBeEnabled()
  })

  test('rejects a past date', async ({ page }) => {
    await openModal(page)
    await page.getByRole('textbox', { name: 'Titre' }).fill('Match du jeudi')
    await page.getByRole('textbox', { name: 'Lieu' }).fill('Stade')
    await page.locator('#match-scheduled').fill('2000-01-01T18:00')
    await expect(page.getByText('La date doit être dans le futur.')).toBeVisible()
    await expect(page.getByRole('button', { name: 'Créer le match' })).toBeDisabled()
  })

  test('submitting valid input creates the match and shows it in the list', async ({ page }) => {
    await authenticate(page)
    await mockGroupAndMatches(page, {})
    await page.route('**/api/v1/matches', async (route) => {
      await route.fulfill({
        status: 201,
        contentType: 'application/json',
        body: JSON.stringify({
          ...SAMPLE_MATCH,
          id: 'm-new',
          title: 'Match de test',
          status: 'draft',
        }),
      })
    })
    await page.goto('/groups/g-1')
    await page.getByRole('button', { name: 'Créer un match' }).click()
    await page.getByRole('textbox', { name: 'Titre' }).fill('Match de test')
    await page.getByRole('textbox', { name: 'Lieu' }).fill('Stade')
    await page.locator('#match-scheduled').fill('2099-01-01T18:00')
    await page.getByRole('button', { name: 'Créer le match' }).click()
    await expect(page.getByRole('dialog')).not.toBeVisible()
    await expect(page.getByText('Match de test')).toBeVisible()
  })

  test('shows the backend error inline when the API returns 400', async ({ page }) => {
    await authenticate(page)
    await mockGroupAndMatches(page, {})
    await page.route('**/api/v1/matches', async (route) => {
      await route.fulfill({
        status: 400,
        contentType: 'application/json',
        body: JSON.stringify({ error: 'invalid name' }),
      })
    })
    await page.goto('/groups/g-1')
    await page.getByRole('button', { name: 'Créer un match' }).click()
    await page.getByRole('textbox', { name: 'Titre' }).fill('Match')
    await page.getByRole('textbox', { name: 'Lieu' }).fill('Stade')
    await page.locator('#match-scheduled').fill('2099-01-01T18:00')
    await page.getByRole('button', { name: 'Créer le match' }).click()
    await expect(page.getByRole('alert')).toContainText('invalid name')
    await expect(page.getByRole('dialog')).toBeVisible()
  })
})
