import { test, expect, type Page } from './fixtures'

const RATE_CONTEXT = {
  group_name: 'Foot du jeudi',
  match_title: 'Match',
  venue: 'Salle Pierre Mendès, Lyon',
  scheduled_at: '2026-05-07T19:30:00+02:00',
  status: 'completed',
  score_a: 3,
  score_b: 2,
  winner: 'A',
  voter: { player_id: 'p-1', name: 'Alice Martin', initials: 'AM', team: 'A' },
  teammates: [
    { player_id: 'p-2', name: 'Bob Durand', initials: 'BD', matches_together: 7, your_score: null },
  ],
  voters_done: 1,
  voters_total: 4,
  results: null,
}

const CLOSED_CONTEXT = {
  ...RATE_CONTEXT,
  status: 'closed',
  results: {
    teammates: [
      { player_id: 'p-2', name: 'Bob Durand', initials: 'BD', average: 4.2, votes_count: 3, delta: 0.3 },
    ],
    self: { average: 3.7, votes_count: 3 },
  },
}

async function mockVoteContext(page: Page, status: number, body: unknown) {
  await page.route('**/api/v1/votes/context', route =>
    route.fulfill({ status, contentType: 'application/json', body: JSON.stringify(body) }),
  )
}

test.describe('Vote page', () => {
  test('rating view renders the match, team banner and teammates', async ({ page }) => {
    await mockVoteContext(page, 200, RATE_CONTEXT)
    await page.goto('/vote/tok-123')
    await expect(page.getByRole('heading', { name: 'Notez vos coéquipiers' })).toBeVisible()
    await expect(page.getByRole('note')).toContainText('Équipe rouge')
    await expect(page.getByText('Bob Durand')).toBeVisible()
    await expect(page.getByTestId('submit-votes')).toBeDisabled()
  })

  test('rating every teammate enables the CTA and submitting shows the thank-you view', async ({ page }) => {
    let contextCalls = 0
    await page.route('**/api/v1/votes/context', (route) => {
      contextCalls += 1
      const body = contextCalls === 1
        ? RATE_CONTEXT
        : {
            ...RATE_CONTEXT,
            teammates: [{ ...RATE_CONTEXT.teammates[0], your_score: 4 }],
            voters_done: 2,
          }
      return route.fulfill({ status: 200, contentType: 'application/json', body: JSON.stringify(body) })
    })
    await page.route('**/api/v1/votes', route =>
      route.fulfill({
        status: 201,
        contentType: 'application/json',
        body: JSON.stringify({ voted_id: 'p-2', score: 4 }),
      }),
    )

    await page.goto('/vote/tok-123')
    await page.getByTestId('teammate-p-2').getByTestId('star-4').click()
    await expect(page.getByTestId('submit-votes')).toBeEnabled()

    await page.getByTestId('submit-votes').click()
    await expect(page.getByRole('heading', { name: 'Confirmer vos votes ?' })).toBeVisible()
    await page.getByTestId('confirm-votes').click()

    await expect(page.getByRole('heading', { name: 'Merci pour votre vote' })).toBeVisible()
    await expect(page.getByText('Votes en cours')).toBeVisible()
  })

  test('a closed match shows the final results and the self card', async ({ page }) => {
    await mockVoteContext(page, 200, CLOSED_CONTEXT)
    await page.goto('/vote/tok-123')
    await expect(page.getByRole('heading', { name: 'Notes finales de l\'équipe rouge' })).toBeVisible()
    await expect(page.getByText('4.2')).toBeVisible()
    await expect(page.getByText('+0.3 vs précédent')).toBeVisible()
    await expect(page.getByRole('heading', { name: 'Votre note moyenne ce match' })).toBeVisible()
  })

  test('an unknown token shows the invalid screen', async ({ page }) => {
    await mockVoteContext(page, 404, { error: 'invitation not found' })
    await page.goto('/vote/nope')
    await expect(page.getByRole('heading', { name: 'Lien invalide' })).toBeVisible()
  })

  test('the invite link routes to the vote page once the match is completed', async ({ page }) => {
    await page.route('**/api/v1/invitations/context', route =>
      route.fulfill({
        status: 200,
        contentType: 'application/json',
        body: JSON.stringify({
          organizer_name: 'Alex L.',
          group_name: 'Foot du jeudi',
          match_title: 'Match',
          venue: 'Salle Pierre Mendès, Lyon',
          scheduled_at: '2026-05-07T19:30:00+02:00',
          match_status: 'completed',
          capacity: '10 (5v5)',
          confirmed_count: 6,
          max_participants: 10,
          confirmed_initials: ['IR'],
          response: 'yes',
          expires_at: '2026-05-07T18:00:00+02:00',
          state: 'locked',
        }),
      }),
    )
    await mockVoteContext(page, 200, RATE_CONTEXT)

    await page.goto('/invite/tok-123')

    await expect(page).toHaveURL(/\/vote\/tok-123$/)
    await expect(page.getByRole('heading', { name: 'Notez vos coéquipiers' })).toBeVisible()
  })
})
