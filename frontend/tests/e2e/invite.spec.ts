import { test, expect, type Page } from './fixtures'

const CONTEXT = {
  organizer_name: 'Alex L.',
  group_name: 'Foot du jeudi',
  match_title: 'Match',
  venue: 'Salle Pierre Mendès, Lyon',
  scheduled_at: '2026-05-07T19:30:00+02:00',
  capacity: '10 (5v5)',
  confirmed_count: 6,
  max_participants: 10,
  confirmed_initials: ['IR', 'TB', 'MR'],
  response: 'pending',
  expires_at: '2026-05-07T18:00:00+02:00',
  state: 'respondable',
}

async function mockContext(page: Page, status: number, body: unknown) {
  await page.route('**/api/v1/invitations/context', route =>
    route.fulfill({ status, contentType: 'application/json', body: JSON.stringify(body) }),
  )
}

async function mockRespond(page: Page, status: number, body: unknown) {
  await page.route('**/api/v1/invitations/respond', route =>
    route.fulfill({ status, contentType: 'application/json', body: JSON.stringify(body) }),
  )
}

test.describe('Invitation page', () => {
  test('initial view renders organizer, group and match details', async ({ page }) => {
    await mockContext(page, 200, CONTEXT)
    await page.goto('/invite/tok-123')
    await expect(page.getByRole('heading', { name: 'Vous êtes invité' })).toBeVisible()
    await expect(page.getByText('Alex L.')).toBeVisible()
    await expect(page.getByText('Foot du jeudi')).toBeVisible()
    await expect(page.getByText('Jeudi 7 mai 2026')).toBeVisible()
  })

  test('answering yes through the modal shows the inscribed result', async ({ page }) => {
    await mockContext(page, 200, CONTEXT)
    await mockRespond(page, 200, { response: 'yes', responded_at: '2026-05-01T10:00:00Z' })
    await page.goto('/invite/tok-123')
    await page.getByTestId('respond-cta').click()
    await expect(page.getByRole('heading', { name: 'Vous venez à ce match ?' })).toBeVisible()
    await page.getByTestId('answer-yes').click()
    await expect(page.getByRole('heading', { name: 'Vous êtes inscrit' })).toBeVisible()
  })

  test('an unknown token shows the invalid screen', async ({ page }) => {
    await mockContext(page, 404, { error: 'invitation not found' })
    await page.goto('/invite/nope')
    await expect(page.getByRole('heading', { name: 'Lien invalide' })).toBeVisible()
  })

  test('an expired invitation shows the expired screen', async ({ page }) => {
    await mockContext(page, 200, { ...CONTEXT, state: 'expired' })
    await page.goto('/invite/tok-123')
    await expect(page.getByRole('heading', { name: 'Invitation expirée' })).toBeVisible()
  })
})
