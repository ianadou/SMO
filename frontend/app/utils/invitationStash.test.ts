import { describe, it, expect, beforeEach } from 'vitest'
import { readStashedInvitation, stashInvitation } from './invitationStash'

beforeEach(() => {
  window.localStorage.clear()
})

describe('invitationStash', () => {
  it('returns the stashed invitation for the same match', () => {
    stashInvitation('m-1', { token: 'tok-perso', playerName: 'Marc' })

    const stashed = readStashedInvitation('m-1')

    expect(stashed).toEqual({ token: 'tok-perso', playerName: 'Marc' })
  })

  it('returns null for another match', () => {
    stashInvitation('m-1', { token: 'tok-perso', playerName: 'Marc' })

    expect(readStashedInvitation('m-2')).toBeNull()
  })

  it('returns null when the stored value is not valid JSON', () => {
    window.localStorage.setItem('smo.invitation.m-1', 'not-json')

    expect(readStashedInvitation('m-1')).toBeNull()
  })

  it('returns null when the stored value misses a field', () => {
    window.localStorage.setItem('smo.invitation.m-1', JSON.stringify({ token: 'tok-perso' }))

    expect(readStashedInvitation('m-1')).toBeNull()
  })
})
