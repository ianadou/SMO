import { test, expect, type Page } from './fixtures'

const ORGANIZER = {
  id: 'org-1',
  email: 'alex@example.fr',
  display_name: 'Alex',
  created_at: '2026-01-01T00:00:00Z',
}

const MATCH = {
  id: 'm-1',
  group_id: 'g-1',
  title: 'Match du jeudi',
  venue: 'Stade municipal',
  scheduled_at: '2026-06-18T19:00:00Z',
  status: 'open',
  score_a: null,
  score_b: null,
  created_at: '2026-06-01T10:00:00Z',
}

const PLAYERS = [
  { id: 'p-1', group_id: 'g-1', name: 'Inès R.', ranking: 1200 },
  { id: 'p-2', group_id: 'g-1', name: 'Théo B.', ranking: 1100 },
]

const INVITATIONS = [
  {
    id: 'inv-1', match_id: 'm-1', player_id: 'p-1', expires_at: '2026-06-20T10:00:00Z',
    response: 'yes', responded_at: '2026-06-09T10:00:00Z', created_at: '2026-06-08T10:00:00Z',
  },
]

async function authenticate(page: Page) {
  await page.goto('/login')
  await page.evaluate(
    ([token, organizer]) => {
      window.localStorage.setItem('smo.auth.token', JSON.stringify(token))
      window.localStorage.setItem('smo.auth.organizer', JSON.stringify(organizer))
    },
    ['fake-jwt-token', ORGANIZER] as const,
  )
}

async function mockMatchPage(page: Page) {
  await page.route(url => url.pathname === '/api/v1/matches/m-1', route =>
    route.fulfill({ status: 200, contentType: 'application/json', body: JSON.stringify(MATCH) }),
  )
  await page.route(url => url.pathname === '/api/v1/matches/m-1/teams', route =>
    route.fulfill({ status: 200, contentType: 'application/json', body: JSON.stringify([]) }),
  )
  await page.route(url => url.pathname === '/api/v1/groups/g-1/players', route =>
    route.fulfill({ status: 200, contentType: 'application/json', body: JSON.stringify(PLAYERS) }),
  )
  await page.route(url => url.pathname === '/api/v1/matches/m-1/invitations', route =>
    route.fulfill({ status: 200, contentType: 'application/json', body: JSON.stringify(INVITATIONS) }),
  )
}

test.describe('Invite management sheet', () => {
  test('the kebab opens the sheet with one row per group player', async ({ page }) => {
    await authenticate(page)
    await mockMatchPage(page)

    await page.goto('/matches/m-1')
    await page.getByTestId('match-menu').click()

    await expect(page.getByRole('heading', { name: 'Invitations' })).toBeVisible()
    await expect(page.getByTestId('invite-row-p-1')).toContainText('✓ Vient')
    await expect(page.getByTestId('invite-p-2')).toContainText('Inviter')
  })

  test('inviting a player flips the row to the one-time share state', async ({ page }) => {
    await authenticate(page)
    await mockMatchPage(page)
    await page.route(url => url.pathname === '/api/v1/invitations', route =>
      route.fulfill({
        status: 201,
        contentType: 'application/json',
        body: JSON.stringify({
          id: 'inv-2', match_id: 'm-1', player_id: 'p-2',
          expires_at: '2026-06-20T10:00:00Z', response: 'pending',
          responded_at: null, created_at: '2026-06-10T10:00:00Z',
          plain_token: 'tok-fresh-secret',
        }),
      }),
    )

    await page.goto('/matches/m-1')
    await page.getByTestId('match-menu').click()
    await page.getByTestId('invite-p-2').click()

    await expect(page.getByTestId('share-p-2')).toContainText('Partager le lien')
  })
})
