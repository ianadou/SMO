import { computed, ref } from 'vue'
import { ApiError, useApi } from './useApi'
import { useToast } from './useToast'
import { deriveScreen } from '~/utils/teamComposition'
import { toFriendlyError } from '~/utils/errorMessages'
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
  const toast = useToast()

  const screen = computed<MatchScreen>(() =>
    match.value
      ? deriveScreen(
          match.value.status,
          members.value.length > 0,
          new Date(),
          match.value.scheduled_at,
        )
      : 'setup-draft',
  )

  function reportError(err: unknown, fallbackTitle: string) {
    const friendly = toFriendlyError(err)
    toast.error(friendly.title ?? fallbackTitle, friendly.message)
    error.value =
      err instanceof ApiError ? err.publicMessage : friendly.message ?? fallbackTitle
  }

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
      reportError(err, 'Chargement impossible')
    } finally {
      loading.value = false
    }
  }

  async function mutate(run: () => Promise<unknown>, fallbackTitle = 'Action impossible') {
    loading.value = true
    error.value = null
    try {
      await run()
    } catch (err) {
      reportError(err, fallbackTitle)
      loading.value = false
      return false
    }
    await load()
    return true
  }

  async function autoGenerate() {
    try {
      await api.post(`/matches/${matchId}/teams/generate`, { strategy: 'ranking' })
    } catch {
      await api.post(`/matches/${matchId}/teams/generate`, { strategy: 'random' })
    }
  }

  function openMatch() {
    return mutate(async () => {
      await api.post(`/matches/${matchId}/open`, {})
      await autoGenerate()
    })
  }

  function generate() {
    return mutate(autoGenerate)
  }

  async function validateTeams(teamA: string[], teamB: string[]) {
    const ok = await mutate(async () => {
      await api.put(`/matches/${matchId}/teams`, {
        team_a: teamA,
        team_b: teamB,
      })
      await api.post(`/matches/${matchId}/teams-ready`, {})
    })
    if (ok) {
      toast.success(
        'Équipes validées',
        "Modifiable jusqu'à 10 min avant le coup d'envoi.",
      )
    }
  }

  async function saveTeams(teamA: string[], teamB: string[]) {
    const ok = await mutate(() =>
      api.put(`/matches/${matchId}/teams`, {
        team_a: teamA,
        team_b: teamB,
      }),
    )
    if (ok) {
      toast.success(
        'Composition mise à jour',
        "Modifiable jusqu'à 10 min avant le coup d'envoi.",
      )
    }
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
    saveTeams,
  }
}
