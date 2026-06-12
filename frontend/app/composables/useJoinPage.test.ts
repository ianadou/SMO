import { describe, it, expect, vi } from 'vitest'
import { useJoinPage } from './useJoinPage'
import { ApiError } from './useApi'
import type { ShareLinkContext } from '~/types/shareLink'

function shareLinkContext(): ShareLinkContext {
  return {
    match_id: 'm-1',
    organizer_name: 'Alex L.',
    group_name: 'Foot du jeudi',
    match_title: 'Match',
    venue: 'Salle Pierre Mendès, Lyon',
    scheduled_at: '2026-06-18T19:30:00+02:00',
    match_status: 'open',
    capacity: '10 (5v5)',
    confirmed_count: 4,
    max_participants: 10,
    confirmed_initials: ['IR', 'TB'],
    roster: [
      { player_id: 'p-1', player_name: 'Inès', state: 'responded' },
      { player_id: 'p-2', player_name: 'Marc', state: 'claimable' },
      { player_id: 'p-3', player_name: 'Théo', state: 'claimed' },
    ],
  }
}

function fakeApi() {
  return {
    get: vi.fn(async () => shareLinkContext()),
    post: vi.fn(async () => ({ invitation_token: 'tok-perso' })),
  }
}

describe('useJoinPage', () => {
  it('loads the context and exposes an ok outcome', async () => {
    const api = fakeApi()
    const page = useJoinPage('tok-share', api)

    await page.load()

    expect(api.get).toHaveBeenCalledWith('/share/tok-share')
    expect(page.outcome.value?.kind).toBe('ok')
    expect(page.loading.value).toBe(false)
  })

  it('maps a 404 to the invalid outcome', async () => {
    const api = fakeApi()
    api.get.mockRejectedValueOnce(new ApiError(404, 'share link not found'))
    const page = useJoinPage('tok-share', api)

    await page.load()

    expect(page.outcome.value?.kind).toBe('invalid')
  })

  it('maps a 500 to the error outcome', async () => {
    const api = fakeApi()
    api.get.mockRejectedValueOnce(new ApiError(500, 'boom'))
    const page = useJoinPage('tok-share', api)

    await page.load()

    expect(page.outcome.value?.kind).toBe('error')
  })

  it('returns the personal token when the claim succeeds', async () => {
    const api = fakeApi()
    const page = useJoinPage('tok-share', api)

    const result = await page.claim('p-2')

    expect(api.post).toHaveBeenCalledWith('/share/tok-share/claim', { player_id: 'p-2' })
    expect(result).toEqual({ kind: 'claimed', invitationToken: 'tok-perso' })
  })

  it('maps a claim 409 to the race outcome', async () => {
    const api = fakeApi()
    api.post.mockRejectedValueOnce(new ApiError(409, 'invitation already claimed'))
    const page = useJoinPage('tok-share', api)

    const result = await page.claim('p-2')

    expect(result).toEqual({ kind: 'race' })
  })

  it('maps a claim 423 to the locked outcome', async () => {
    const api = fakeApi()
    api.post.mockRejectedValueOnce(new ApiError(423, 'attendance locked'))
    const page = useJoinPage('tok-share', api)

    const result = await page.claim('p-2')

    expect(result).toEqual({ kind: 'locked' })
  })

  it('maps a claim 500 to the failed outcome', async () => {
    const api = fakeApi()
    api.post.mockRejectedValueOnce(new ApiError(500, 'boom'))
    const page = useJoinPage('tok-share', api)

    const result = await page.claim('p-2')

    expect(result).toEqual({ kind: 'failed' })
  })

  it('returns the personal token when the join succeeds', async () => {
    const api = fakeApi()
    const page = useJoinPage('tok-share', api)

    const result = await page.join('Nadia')

    expect(api.post).toHaveBeenCalledWith('/share/tok-share/join', { player_name: 'Nadia' })
    expect(result).toEqual({ kind: 'joined', invitationToken: 'tok-perso' })
  })

  it('maps a join 409 to the name-taken outcome', async () => {
    const api = fakeApi()
    api.post.mockRejectedValueOnce(new ApiError(409, 'player already invited to this match'))
    const page = useJoinPage('tok-share', api)

    const result = await page.join('Marc')

    expect(result).toEqual({ kind: 'name-taken' })
  })

  it('maps a join 423 to the locked outcome', async () => {
    const api = fakeApi()
    api.post.mockRejectedValueOnce(new ApiError(423, 'attendance locked'))
    const page = useJoinPage('tok-share', api)

    const result = await page.join('Nadia')

    expect(result).toEqual({ kind: 'locked' })
  })
})
