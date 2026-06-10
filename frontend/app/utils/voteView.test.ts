import { describe, it, expect } from 'vitest'
import { resolveVoteView } from './voteView'
import type { VoteContext, VoteLoadOutcome } from '~/types/vote'

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
    voter: { player_id: 'p-1', name: 'Alice', initials: 'AM', team: 'A' },
    teammates: [
      { player_id: 'p-2', name: 'Bob', initials: 'BD', matches_together: 7, your_score: null },
    ],
    voters_done: 0,
    voters_total: 4,
    results: null,
    ...over,
  }
}

function ok(over: Partial<VoteContext> = {}): VoteLoadOutcome {
  return { kind: 'ok', context: voteContext(over) }
}

describe('resolveVoteView', () => {
  it('maps an invalid token outcome to the invalid view', () => {
    expect(resolveVoteView({ kind: 'invalid' })).toBe('invalid')
  })

  it('maps a declined bearer outcome to the not-participant view', () => {
    expect(resolveVoteView({ kind: 'not-participant' })).toBe('not-participant')
  })

  it('maps a load failure to the error view', () => {
    expect(resolveVoteView({ kind: 'error' })).toBe('error')
  })

  it('maps every pre-completed status to the too-early view', () => {
    for (const status of ['draft', 'open', 'teams_ready', 'in_progress'] as const) {
      expect(resolveVoteView(ok({ status }))).toBe('too-early')
    }
  })

  it('maps completed with unrated teammates to the rate view', () => {
    expect(resolveVoteView(ok())).toBe('rate')
  })

  it('maps completed with all teammates rated to the submitted view', () => {
    const outcome = ok({
      teammates: [
        { player_id: 'p-2', name: 'Bob', initials: 'BD', matches_together: 7, your_score: 4 },
      ],
    })
    expect(resolveVoteView(outcome)).toBe('submitted')
  })

  it('maps a closed match to the results view', () => {
    expect(resolveVoteView(ok({ status: 'closed' }))).toBe('results')
  })
})
