import { describe, it, expect, vi } from 'vitest'
import { useMatchInvitations } from './useMatchInvitations'
import type { PlayerDTO } from '~/types/groups'
import type { MatchInvitationDTO } from '~/types/invitation'

const players: PlayerDTO[] = [
  { id: 'p-1', group_id: 'g-1', name: 'Inès R.', ranking: 1200 },
  { id: 'p-2', group_id: 'g-1', name: 'Théo B.', ranking: 1100 },
  { id: 'p-3', group_id: 'g-1', name: 'Marc R.', ranking: 1000 },
]

const invitations: MatchInvitationDTO[] = [
  {
    id: 'inv-1', match_id: 'm-1', player_id: 'p-1', expires_at: '2026-06-15T10:00:00Z',
    response: 'yes', responded_at: '2026-06-09T10:00:00Z', created_at: '2026-06-08T10:00:00Z',
  },
  {
    id: 'inv-2', match_id: 'm-1', player_id: 'p-2', expires_at: '2026-06-15T10:00:00Z',
    response: 'pending', responded_at: null, created_at: '2026-06-08T10:00:00Z',
  },
]

function fakeApi() {
  return {
    get: vi.fn(async (path: string) =>
      path.includes('/players') ? players : invitations,
    ),
    post: vi.fn(async () => ({
      id: 'inv-3', match_id: 'm-1', player_id: 'p-3', expires_at: '2026-06-15T10:00:00Z',
      response: 'pending', responded_at: null, created_at: '2026-06-10T10:00:00Z',
      plain_token: 'tok-secret',
    })),
  }
}

describe('useMatchInvitations', () => {
  it('merges group players with their invitation status', async () => {
    const api = fakeApi()
    const panel = useMatchInvitations('m-1', api)

    await panel.load('g-1')

    expect(api.get).toHaveBeenCalledWith('/groups/g-1/players')
    expect(api.get).toHaveBeenCalledWith('/matches/m-1/invitations')
    expect(panel.rows.value).toEqual([
      { playerId: 'p-1', playerName: 'Inès R.', status: 'yes', shareUrl: null },
      { playerId: 'p-2', playerName: 'Théo B.', status: 'pending', shareUrl: null },
      { playerId: 'p-3', playerName: 'Marc R.', status: 'not-invited', shareUrl: null },
    ])
  })

  it('counts confirmed players', async () => {
    const panel = useMatchInvitations('m-1', fakeApi())

    await panel.load('g-1')

    expect(panel.confirmedCount.value).toBe(1)
  })

  it('creates an invitation and flips the row to fresh with the share url', async () => {
    const api = fakeApi()
    const panel = useMatchInvitations('m-1', api)
    await panel.load('g-1')

    const url = await panel.invite('p-3')

    expect(api.post).toHaveBeenCalledWith('/invitations', { match_id: 'm-1', player_id: 'p-3' })
    expect(url).toContain('/invite/tok-secret')
    const row = panel.rows.value.find(r => r.playerId === 'p-3')
    expect(row?.status).toBe('fresh')
    expect(row?.shareUrl).toContain('/invite/tok-secret')
  })

  it('returns null and keeps the row intact when creation fails', async () => {
    const api = fakeApi()
    api.post.mockRejectedValueOnce(new Error('boom'))
    const panel = useMatchInvitations('m-1', api)
    await panel.load('g-1')

    const url = await panel.invite('p-3')

    expect(url).toBeNull()
    expect(panel.rows.value.find(r => r.playerId === 'p-3')?.status).toBe('not-invited')
  })

  it('flags the error state when loading fails', async () => {
    const api = fakeApi()
    api.get.mockRejectedValueOnce(new Error('boom'))
    const panel = useMatchInvitations('m-1', api)

    await panel.load('g-1')

    expect(panel.error.value).toBe(true)
  })
})
