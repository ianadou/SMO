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
    delete: vi.fn(async () => undefined),
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
      { playerId: 'p-1', playerName: 'Inès R.', status: 'yes', shareUrl: null, claimed: false },
      { playerId: 'p-2', playerName: 'Théo B.', status: 'pending', shareUrl: null, claimed: false },
      { playerId: 'p-3', playerName: 'Marc R.', status: 'not-invited', shareUrl: null, claimed: false },
    ])
  })

  it('flags a row as claimed when its invitation carries claimed_at', async () => {
    const claimedInvitation: MatchInvitationDTO = {
      id: 'inv-1', match_id: 'm-1', player_id: 'p-1', expires_at: '2026-06-15T10:00:00Z',
      response: 'yes', responded_at: '2026-06-09T10:00:00Z', created_at: '2026-06-08T10:00:00Z',
      claimed_at: '2026-06-09T12:00:00Z',
    }
    const api = fakeApi()
    api.get.mockImplementation(async (path: string) =>
      path.includes('/players') ? players : [claimedInvitation],
    )
    const panel = useMatchInvitations('m-1', api)

    await panel.load('g-1')

    expect(panel.rows.value.find(row => row.playerId === 'p-1')?.claimed).toBe(true)
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

  it('generates the match share link and keeps the full url in memory', async () => {
    const api = fakeApi()
    api.post.mockResolvedValueOnce({
      token: 'tok-share', url: 'http://x/join/tok-share', expires_at: '2026-06-17T10:00:00Z',
    })
    const panel = useMatchInvitations('m-1', api)

    const generated = await panel.generateShareLink()

    expect(api.post).toHaveBeenCalledWith('/matches/m-1/share-link', {})
    expect(generated).toBe(true)
    expect(panel.shareLink.value?.url).toBe('http://x/join/tok-share')
  })

  it('derives the join url from the token when the response omits it', async () => {
    const api = fakeApi()
    api.post.mockResolvedValueOnce({
      token: 'tok-share', url: '', expires_at: '2026-06-17T10:00:00Z',
    })
    const panel = useMatchInvitations('m-1', api)

    await panel.generateShareLink()

    expect(panel.shareLink.value?.url).toContain('/join/tok-share')
  })

  it('returns false and keeps no link when the generation fails', async () => {
    const api = fakeApi()
    api.post.mockRejectedValueOnce(new Error('boom'))
    const panel = useMatchInvitations('m-1', api)

    const generated = await panel.generateShareLink()

    expect(generated).toBe(false)
    expect(panel.shareLink.value).toBeNull()
  })

  it('revokes the match share link and forgets it', async () => {
    const api = fakeApi()
    api.post.mockResolvedValueOnce({
      token: 'tok-share', url: 'http://x/join/tok-share', expires_at: '2026-06-17T10:00:00Z',
    })
    const panel = useMatchInvitations('m-1', api)
    await panel.generateShareLink()

    const revoked = await panel.revokeShareLink()

    expect(api.delete).toHaveBeenCalledWith('/matches/m-1/share-link')
    expect(revoked).toBe(true)
    expect(panel.shareLink.value).toBeNull()
  })

  it('keeps the link when the revocation fails', async () => {
    const api = fakeApi()
    api.post.mockResolvedValueOnce({
      token: 'tok-share', url: 'http://x/join/tok-share', expires_at: '2026-06-17T10:00:00Z',
    })
    api.delete.mockRejectedValueOnce(new Error('boom'))
    const panel = useMatchInvitations('m-1', api)
    await panel.generateShareLink()

    const revoked = await panel.revokeShareLink()

    expect(revoked).toBe(false)
    expect(panel.shareLink.value?.url).toBe('http://x/join/tok-share')
  })
})
