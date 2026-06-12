import { describe, it, expect } from 'vitest'
import { resolveJoinView } from './shareLinkView'
import type { ShareLinkContext } from '~/types/shareLink'
import type { MatchStatus } from '~/types/matches'

function contextWithStatus(status: MatchStatus): ShareLinkContext {
  return {
    match_id: 'm-1',
    organizer_name: 'Alex L.',
    group_name: 'Foot du jeudi',
    match_title: 'Match',
    venue: 'Stade',
    scheduled_at: '2026-06-18T19:30:00+02:00',
    match_status: status,
    capacity: '10 (5v5)',
    confirmed_count: 0,
    max_participants: 10,
    confirmed_initials: [],
    roster: [],
  }
}

describe('resolveJoinView', () => {
  it('returns invalid for an invalid outcome', () => {
    expect(resolveJoinView({ kind: 'invalid' })).toBe('invalid')
  })

  it('returns error for an error outcome', () => {
    expect(resolveJoinView({ kind: 'error' })).toBe('error')
  })

  it('returns roster while the match is open', () => {
    expect(resolveJoinView({ kind: 'ok', context: contextWithStatus('open') })).toBe('roster')
  })

  it('returns roster while the match is a draft', () => {
    expect(resolveJoinView({ kind: 'ok', context: contextWithStatus('draft') })).toBe('roster')
  })

  it('returns locked once the teams are formed', () => {
    expect(resolveJoinView({ kind: 'ok', context: contextWithStatus('teams_ready') })).toBe('locked')
  })
})
