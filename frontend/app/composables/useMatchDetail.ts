import { computed, ref } from 'vue'
import { ApiError, useApi } from './useApi'
import { deriveScreen } from '~/utils/teamComposition'
import type { MatchDTO, MatchScreen, TeamMemberDTO } from '~/types/matches'

export interface MatchApi {
  get: <T>(path: string) => Promise<T>
  post: <T>(path: string, body: unknown) => Promise<T>
  put: <T>(path: string, body: unknown) => Promise<T>
}

export function useMatchDetail(matchId: string, api: MatchApi = useApi()) {
  const match = ref<MatchDTO | null>(null)
  const members = ref<TeamMemberDTO[]>([])
  const loading = ref(false)
  const error = ref<string | null>(null)

  const screen = computed<MatchScreen>(() =>
    match.value
      ? deriveScreen(match.value.status, members.value.length > 0)
      : 'setup-draft',
  )

  async function load() {
    loading.value = true
    error.value = null
    try {
      const [loaded, team] = await Promise.all([
        api.get<MatchDTO>(`/matches/${matchId}`),
        api.get<TeamMemberDTO[]>(`/matches/${matchId}/teams`),
      ])
      match.value = loaded
      members.value = team ?? []
    } catch (err) {
      error.value =
        err instanceof ApiError ? err.publicMessage : 'Une erreur est survenue.'
    } finally {
      loading.value = false
    }
  }

  async function mutate(run: () => Promise<unknown>) {
    loading.value = true
    error.value = null
    try {
      await run()
    } catch (err) {
      error.value =
        err instanceof ApiError ? err.publicMessage : 'Une erreur est survenue.'
      loading.value = false
      return
    }
    await load()
  }

  function openMatch() {
    return mutate(() => api.post(`/matches/${matchId}/open`, {}))
  }

  function generate(strategy: 'random' | 'ranking') {
    return mutate(() =>
      api.post(`/matches/${matchId}/teams/generate`, { strategy }),
    )
  }

  function validateTeams(teamA: string[], teamB: string[]) {
    return mutate(async () => {
      await api.put(`/matches/${matchId}/teams`, {
        team_a: teamA,
        team_b: teamB,
      })
      await api.post(`/matches/${matchId}/teams-ready`, {})
    })
  }

  return {
    match,
    members,
    screen,
    loading,
    error,
    load,
    openMatch,
    generate,
    validateTeams,
  }
}
