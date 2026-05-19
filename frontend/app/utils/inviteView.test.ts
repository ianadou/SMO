import { describe, it, expect } from 'vitest'
import { resolveInviteView } from './inviteView'
import type { InvitationContext, LoadOutcome } from '~/types/invitation'

function ctx(over: Partial<InvitationContext>): InvitationContext {
  return {
    organizer_name: 'Alex L.',
    group_name: 'Foot du jeudi',
    match_title: 'Match',
    venue: 'Salle',
    scheduled_at: '2026-05-07T19:30:00+02:00',
    capacity: '10 (5v5)',
    confirmed_count: 6,
    max_participants: 10,
    confirmed_initials: ['IR', 'TB'],
    response: 'pending',
    expires_at: '2026-05-07T18:00:00+02:00',
    state: 'respondable',
    ...over,
  }
}

describe('resolveInviteView', () => {
  it('returns invalid when load failed with an unknown/missing token', () => {
    expect(resolveInviteView({ kind: 'invalid' } as LoadOutcome)).toBe('invalid')
  })

  it('returns error on network/server failure', () => {
    expect(resolveInviteView({ kind: 'error' } as LoadOutcome)).toBe('error')
  })

  it('returns expired whenever state is expired regardless of response', () => {
    expect(resolveInviteView({ kind: 'ok', context: ctx({ state: 'expired', response: 'yes' }) })).toBe('expired')
  })

  it('returns initial when respondable and still pending', () => {
    expect(resolveInviteView({ kind: 'ok', context: ctx({ state: 'respondable', response: 'pending' }) })).toBe('initial')
  })

  it('returns result when respondable and already answered', () => {
    expect(resolveInviteView({ kind: 'ok', context: ctx({ state: 'respondable', response: 'no' }) })).toBe('result')
  })

  it('returns result when locked and an answer exists', () => {
    expect(resolveInviteView({ kind: 'ok', context: ctx({ state: 'locked', response: 'yes' }) })).toBe('result')
  })

  it('returns locked-pending when locked but never answered', () => {
    expect(resolveInviteView({ kind: 'ok', context: ctx({ state: 'locked', response: 'pending' }) })).toBe('locked-pending')
  })
})
