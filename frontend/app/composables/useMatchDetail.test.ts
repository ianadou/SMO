import { describe, it, expect, vi } from 'vitest'
import { nextTick } from 'vue'
import { useMatchDetail } from './useMatchDetail'
import { ApiError } from './useApi'
import type { MatchDTO, TeamMemberDTO } from '~/types/matches'

function matchDTO(over: Partial<MatchDTO> = {}): MatchDTO {
  return {
    id: 'm1',
    group_id: 'g1',
    title: 'Match',
    venue: 'Stade',
    scheduled_at: '2026-06-01T19:00:00Z',
    status: 'open',
    score_a: null,
    score_b: null,
    created_at: '2026-05-01T10:00:00Z',
    ...over,
  }
}

function fakeApi(match: MatchDTO, members: TeamMemberDTO[]) {
  return {
    get: vi.fn(async (path: string) =>
      path.endsWith('/teams') ? members : match,
    ),
    post: vi.fn(async () => ({})),
    put: vi.fn(async () => ({})),
  }
}

async function flush() {
  await Promise.resolve()
  await nextTick()
}

describe('useMatchDetail', () => {
  it('loads the match and its teams, deriving the screen', async () => {
    const api = fakeApi(matchDTO({ status: 'open' }), [
      { player_id: 'p1', player_name: 'A', team: 'A', slot: 0 },
    ])
    const md = useMatchDetail('m1', api)

    await md.load()

    expect(api.get).toHaveBeenCalledWith('/matches/m1')
    expect(api.get).toHaveBeenCalledWith('/matches/m1/teams')
    expect(md.match.value?.id).toBe('m1')
    expect(md.screen.value).toBe('composition')
    expect(md.loading.value).toBe(false)
    expect(md.error.value).toBeNull()
  })

  it('derives the generate screen when open without teams', async () => {
    const api = fakeApi(matchDTO({ status: 'open' }), [])
    const md = useMatchDetail('m1', api)
    await md.load()
    expect(md.screen.value).toBe('setup-generate')
  })

  it('surfaces the api error message and stops loading', async () => {
    const api = fakeApi(matchDTO(), [])
    api.get.mockRejectedValueOnce(new ApiError(404, 'match not found'))
    const md = useMatchDetail('m1', api)

    await md.load()

    expect(md.error.value).toBe('match not found')
    expect(md.loading.value).toBe(false)
  })

  it('opens the match then refetches', async () => {
    const api = fakeApi(matchDTO({ status: 'draft' }), [])
    const md = useMatchDetail('m1', api)
    await md.load()

    await md.openMatch()

    expect(api.post).toHaveBeenCalledWith('/matches/m1/open', {})
    expect(api.get).toHaveBeenCalledTimes(4)
  })

  it('generates teams with the chosen strategy then refetches', async () => {
    const api = fakeApi(matchDTO({ status: 'open' }), [])
    const md = useMatchDetail('m1', api)
    await md.load()

    await md.generate('ranking')

    expect(api.post).toHaveBeenCalledWith('/matches/m1/teams/generate', {
      strategy: 'ranking',
    })
  })

  it('validates teams: PUT composition then POST teams-ready', async () => {
    const api = fakeApi(matchDTO({ status: 'open' }), [])
    const md = useMatchDetail('m1', api)
    await md.load()

    await md.validateTeams(['p1', 'p2'], ['p3', 'p4'])

    expect(api.put).toHaveBeenCalledWith('/matches/m1/teams', {
      team_a: ['p1', 'p2'],
      team_b: ['p3', 'p4'],
    })
    expect(api.post).toHaveBeenCalledWith('/matches/m1/teams-ready', {})
  })

  it('reports the error and stops loading when a mutation fails', async () => {
    const api = fakeApi(matchDTO({ status: 'draft' }), [])
    api.post.mockRejectedValueOnce(new ApiError(409, 'cannot open'))
    const md = useMatchDetail('m1', api)
    await md.load()

    await md.openMatch()

    expect(md.error.value).toBe('cannot open')
    expect(md.loading.value).toBe(false)
  })
})
