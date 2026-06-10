import { describe, it, expect } from 'vitest'
import { deriveScreen, splitTeams, toTeamArrays } from './teamComposition'
import type { TeamMemberDTO } from '~/types/matches'

describe('deriveScreen', () => {
  const kickoff = '2026-06-01T19:00:00Z'
  const wellBeforeLock = new Date('2026-06-01T18:00:00Z')
  const exactlyAtLock = new Date('2026-06-01T18:50:00Z')
  const pastLock = new Date('2026-06-01T18:55:00Z')

  it('maps draft to the open-match setup screen', () => {
    expect(deriveScreen('draft', false, wellBeforeLock, kickoff)).toBe('setup-draft')
    expect(deriveScreen('draft', true, wellBeforeLock, kickoff)).toBe('setup-draft')
  })

  it('maps open without teams to the generate setup screen', () => {
    expect(deriveScreen('open', false, wellBeforeLock, kickoff)).toBe('setup-generate')
  })

  it('maps open with teams to the editable composition screen before the lock', () => {
    expect(deriveScreen('open', true, wellBeforeLock, kickoff)).toBe('composition')
  })

  it('maps teams_ready to the editable composition screen before the lock', () => {
    expect(deriveScreen('teams_ready', true, wellBeforeLock, kickoff)).toBe('composition')
  })

  it('locks the composition exactly ten minutes before kickoff', () => {
    expect(deriveScreen('open', true, exactlyAtLock, kickoff)).toBe('locked')
  })

  it('locks the composition past the lock threshold for both editable statuses', () => {
    expect(deriveScreen('open', true, pastLock, kickoff)).toBe('locked')
    expect(deriveScreen('teams_ready', true, pastLock, kickoff)).toBe('locked')
  })

  it('maps in_progress to the read-only finished screen', () => {
    expect(deriveScreen('in_progress', true, wellBeforeLock, kickoff)).toBe('finished')
  })

  it('maps completed and closed to the scored read-only screen', () => {
    expect(deriveScreen('completed', true, wellBeforeLock, kickoff)).toBe('closed')
    expect(deriveScreen('closed', true, wellBeforeLock, kickoff)).toBe('closed')
  })
})

describe('splitTeams', () => {
  const members: TeamMemberDTO[] = [
    { player_id: 'p3', player_name: 'C', team: 'A', slot: 1 },
    { player_id: 'p1', player_name: 'A', team: 'A', slot: 0 },
    { player_id: 'p2', player_name: 'B', team: 'B', slot: 0 },
  ]

  it('groups members by side and orders each side by slot', () => {
    const { teamA, teamB } = splitTeams(members)
    expect(teamA.map((m) => m.player_id)).toEqual(['p1', 'p3'])
    expect(teamB.map((m) => m.player_id)).toEqual(['p2'])
  })

  it('returns empty sides when there are no members', () => {
    const { teamA, teamB } = splitTeams([])
    expect(teamA).toEqual([])
    expect(teamB).toEqual([])
  })
})

describe('toTeamArrays', () => {
  it('serializes ordered sides into the PUT /teams payload shape', () => {
    const teamA: TeamMemberDTO[] = [
      { player_id: 'p1', player_name: 'A', team: 'A', slot: 0 },
      { player_id: 'p3', player_name: 'C', team: 'A', slot: 1 },
    ]
    const teamB: TeamMemberDTO[] = [
      { player_id: 'p2', player_name: 'B', team: 'B', slot: 0 },
    ]
    expect(toTeamArrays(teamA, teamB)).toEqual({
      team_a: ['p1', 'p3'],
      team_b: ['p2'],
    })
  })
})
