// @vitest-environment nuxt
import { describe, it, expect, vi, beforeEach } from 'vitest'
import { mountSuspended, mockNuxtImport } from '@nuxt/test-utils/runtime'
import { flushPromises } from '@vue/test-utils'
import VotePage from './[token].vue'
import { ApiError } from '~/composables/useApi'
import type { VoteContext } from '~/types/vote'

const post = vi.fn()

mockNuxtImport('useApi', () => () => ({
  get: vi.fn(),
  post,
  patch: vi.fn(),
  delete: vi.fn(),
}))

mockNuxtImport('useRoute', () => () => ({ params: { token: 'tok-123' } }))

function voteContext(over: Partial<VoteContext> = {}): VoteContext {
  return {
    group_name: 'Foot du jeudi',
    match_title: 'Match',
    venue: 'Stade',
    scheduled_at: '2026-06-04T19:00:00Z',
    status: 'completed',
    score_a: 3,
    score_b: 2,
    winner: 'A',
    voter: { player_id: 'p-1', name: 'Alice Martin', initials: 'AM', team: 'A' },
    teammates: [
      { player_id: 'p-2', name: 'Bob Durand', initials: 'BD', matches_together: 7, your_score: null },
      { player_id: 'p-3', name: 'Carol Petit', initials: 'CP', matches_together: 3, your_score: null },
    ],
    voters_done: 1,
    voters_total: 4,
    results: null,
    ...over,
  }
}

beforeEach(() => {
  post.mockReset()
})

describe('Vote page', () => {
  it('shows the rating view with teammates and a disabled CTA', async () => {
    post.mockResolvedValueOnce(voteContext())
    const wrapper = await mountSuspended(VotePage)
    await flushPromises()

    expect(wrapper.text()).toContain('Notez vos coéquipiers')
    expect(wrapper.text()).toContain('Bob Durand')
    expect(wrapper.text()).toContain('7')
    expect(wrapper.text()).toContain('Équipe rouge')
    expect(wrapper.get('[data-testid="submit-votes"]').attributes('disabled')).toBeDefined()
  })

  it('posts the token from the route to the context endpoint', async () => {
    post.mockResolvedValueOnce(voteContext())
    await mountSuspended(VotePage)
    await flushPromises()

    expect(post).toHaveBeenCalledWith('/votes/context', { token: 'tok-123' })
  })

  it('enables the CTA once every teammate is rated then submits through the modal', async () => {
    post.mockResolvedValueOnce(voteContext())
    post.mockResolvedValueOnce({})
    post.mockResolvedValueOnce({})
    post.mockResolvedValueOnce(voteContext({
      teammates: [
        { player_id: 'p-2', name: 'Bob Durand', initials: 'BD', matches_together: 7, your_score: 4 },
        { player_id: 'p-3', name: 'Carol Petit', initials: 'CP', matches_together: 3, your_score: 5 },
      ],
      voters_done: 2,
    }))
    const wrapper = await mountSuspended(VotePage)
    await flushPromises()

    await wrapper.get('[data-testid="teammate-p-2"] [data-testid="star-4"]').trigger('click')
    await wrapper.get('[data-testid="teammate-p-3"] [data-testid="star-5"]').trigger('click')
    expect(wrapper.get('[data-testid="submit-votes"]').attributes('disabled')).toBeUndefined()

    await wrapper.get('[data-testid="submit-votes"]').trigger('click')
    await wrapper.get('[data-testid="confirm-votes"]').trigger('click')
    await flushPromises()

    expect(post).toHaveBeenCalledWith('/votes', { token: 'tok-123', voted_id: 'p-2', score: 4 })
    expect(post).toHaveBeenCalledWith('/votes', { token: 'tok-123', voted_id: 'p-3', score: 5 })
    expect(wrapper.text()).toContain('Merci pour votre vote')
  })

  it('locks already-rated teammates and only requires the remaining ones', async () => {
    post.mockResolvedValueOnce(voteContext({
      teammates: [
        { player_id: 'p-2', name: 'Bob Durand', initials: 'BD', matches_together: 7, your_score: 4 },
        { player_id: 'p-3', name: 'Carol Petit', initials: 'CP', matches_together: 3, your_score: null },
      ],
    }))
    const wrapper = await mountSuspended(VotePage)
    await flushPromises()

    expect(wrapper.get('[data-testid="teammate-p-2"]').text()).toContain('✓ voté')

    await wrapper.get('[data-testid="teammate-p-3"] [data-testid="star-3"]').trigger('click')
    expect(wrapper.get('[data-testid="submit-votes"]').attributes('disabled')).toBeUndefined()
  })

  it('shows the submitted view with collective progress when everything is rated', async () => {
    post.mockResolvedValueOnce(voteContext({
      teammates: [
        { player_id: 'p-2', name: 'Bob Durand', initials: 'BD', matches_together: 7, your_score: 4 },
      ],
      voters_done: 2,
      voters_total: 4,
    }))
    const wrapper = await mountSuspended(VotePage)
    await flushPromises()

    expect(wrapper.text()).toContain('Merci pour votre vote')
    expect(wrapper.text()).toContain('Votes en cours')
  })

  it('shows the results view with averages, delta and the self card when closed', async () => {
    post.mockResolvedValueOnce(voteContext({
      status: 'closed',
      results: {
        teammates: [
          { player_id: 'p-2', name: 'Bob Durand', initials: 'BD', average: 4.2, votes_count: 3, delta: 0.3 },
        ],
        self: { average: 3.7, votes_count: 3 },
      },
    }))
    const wrapper = await mountSuspended(VotePage)
    await flushPromises()

    expect(wrapper.text()).toContain('Notes finales de l\'équipe rouge')
    expect(wrapper.text()).toContain('4.2')
    expect(wrapper.text()).toContain('+0.3 vs précédent')
    expect(wrapper.text()).toContain('Votre note moyenne ce match')
    expect(wrapper.text()).toContain('3.7')
  })

  it('shows the too-early view before the match is completed', async () => {
    post.mockResolvedValueOnce(voteContext({ status: 'open', teammates: [] }))
    const wrapper = await mountSuspended(VotePage)
    await flushPromises()

    expect(wrapper.text()).toContain('Le vote ouvrira après le match')
  })

  it('shows the invalid screen on a 404', async () => {
    post.mockRejectedValueOnce(new ApiError(404, 'invitation not found'))
    const wrapper = await mountSuspended(VotePage)
    await flushPromises()

    expect(wrapper.text()).toContain('Lien invalide')
  })

  it('shows the not-participant screen on a 403', async () => {
    post.mockRejectedValueOnce(new ApiError(403, 'not a confirmed participant for this match'))
    const wrapper = await mountSuspended(VotePage)
    await flushPromises()

    expect(wrapper.text()).toContain('Vous n\'avez pas participé')
  })
})
