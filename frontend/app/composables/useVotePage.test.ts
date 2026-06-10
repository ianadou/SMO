import { describe, it, expect, vi } from 'vitest'
import { useVotePage } from './useVotePage'
import { ApiError } from './useApi'
import type { VoteContext } from '~/types/vote'

function voteContext(): VoteContext {
  return {
    group_name: 'Foot du jeudi',
    match_title: 'Match',
    venue: 'Stade',
    scheduled_at: '2026-06-04T19:00:00Z',
    status: 'completed',
    score_a: 3,
    score_b: 2,
    winner: 'A',
    voter: { player_id: 'p-1', name: 'Alice', initials: 'AM', team: 'A' },
    teammates: [
      { player_id: 'p-2', name: 'Bob', initials: 'BD', matches_together: 7, your_score: null },
    ],
    voters_done: 0,
    voters_total: 4,
    results: null,
  }
}

describe('useVotePage', () => {
  it('loads the context and exposes an ok outcome', async () => {
    const api = { post: vi.fn(async () => voteContext()) }
    const page = useVotePage('tok-1', api)

    await page.load()

    expect(api.post).toHaveBeenCalledWith('/votes/context', { token: 'tok-1' })
    expect(page.outcome.value?.kind).toBe('ok')
    expect(page.loading.value).toBe(false)
  })

  it('maps a 403 to the not-participant outcome', async () => {
    const api = { post: vi.fn(async () => { throw new ApiError(403, 'not a confirmed participant') }) }
    const page = useVotePage('tok-1', api)

    await page.load()

    expect(page.outcome.value?.kind).toBe('not-participant')
  })

  it('maps a 404 to the invalid outcome', async () => {
    const api = { post: vi.fn(async () => { throw new ApiError(404, 'invitation not found') }) }
    const page = useVotePage('tok-1', api)

    await page.load()

    expect(page.outcome.value?.kind).toBe('invalid')
  })

  it('maps a 500 to the error outcome', async () => {
    const api = { post: vi.fn(async () => { throw new ApiError(500, 'boom') }) }
    const page = useVotePage('tok-1', api)

    await page.load()

    expect(page.outcome.value?.kind).toBe('error')
  })

  it('submits each vote with the token then reloads the context', async () => {
    const api = { post: vi.fn(async () => voteContext()) }
    const page = useVotePage('tok-1', api)

    const failed = await page.submitVotes([
      { voted_id: 'p-2', score: 4 },
      { voted_id: 'p-3', score: 5 },
    ])

    expect(failed).toEqual([])
    expect(api.post).toHaveBeenCalledWith('/votes', { token: 'tok-1', voted_id: 'p-2', score: 4 })
    expect(api.post).toHaveBeenCalledWith('/votes', { token: 'tok-1', voted_id: 'p-3', score: 5 })
    expect(api.post).toHaveBeenLastCalledWith('/votes/context', { token: 'tok-1' })
  })

  it('treats an already-voted 409 as success', async () => {
    const api = {
      post: vi.fn(async (path: string) => {
        if (path === '/votes') {
          throw new ApiError(409, 'already voted for this player in this match')
        }
        return voteContext()
      }),
    }
    const page = useVotePage('tok-1', api)

    const failed = await page.submitVotes([{ voted_id: 'p-2', score: 4 }])

    expect(failed).toEqual([])
  })

  it('reports the ids whose vote failed for another reason', async () => {
    const api = {
      post: vi.fn(async (path: string, body: unknown) => {
        const voted = (body as { voted_id?: string }).voted_id
        if (path === '/votes' && voted === 'p-3') {
          throw new ApiError(500, 'boom')
        }
        return voteContext()
      }),
    }
    const page = useVotePage('tok-1', api)

    const failed = await page.submitVotes([
      { voted_id: 'p-2', score: 4 },
      { voted_id: 'p-3', score: 5 },
    ])

    expect(failed).toEqual(['p-3'])
  })
})
