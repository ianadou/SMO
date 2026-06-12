import { ref } from 'vue'
import { ApiError, useApi } from './useApi'
import type {
  ClaimedInvitationDTO,
  ClaimOutcome,
  JoinOutcome,
  ShareLinkContext,
  ShareLinkLoadOutcome,
} from '~/types/shareLink'

export interface ShareApi {
  get: <T>(path: string) => Promise<T>
  post: <T>(path: string, body: unknown) => Promise<T>
}

export function useJoinPage(token: string, api: ShareApi = useApi()) {
  const outcome = ref<ShareLinkLoadOutcome | null>(null)
  const loading = ref(true)
  const submitting = ref(false)

  async function load() {
    loading.value = true
    try {
      const context = await api.get<ShareLinkContext>(`/share/${token}`)
      outcome.value = { kind: 'ok', context }
    }
    catch (err) {
      if (err instanceof ApiError && (err.status === 404 || err.status === 400)) {
        outcome.value = { kind: 'invalid' }
      }
      else {
        outcome.value = { kind: 'error' }
      }
    }
    finally {
      loading.value = false
    }
  }

  async function claim(playerId: string): Promise<ClaimOutcome> {
    submitting.value = true
    try {
      const claimed = await api.post<ClaimedInvitationDTO>(`/share/${token}/claim`, {
        player_id: playerId,
      })
      return { kind: 'claimed', invitationToken: claimed.invitation_token }
    }
    catch (err) {
      if (err instanceof ApiError && err.status === 409) return { kind: 'race' }
      if (err instanceof ApiError && err.status === 423) return { kind: 'locked' }
      return { kind: 'failed' }
    }
    finally {
      submitting.value = false
    }
  }

  async function join(playerName: string): Promise<JoinOutcome> {
    submitting.value = true
    try {
      const joined = await api.post<ClaimedInvitationDTO>(`/share/${token}/join`, {
        player_name: playerName,
      })
      return { kind: 'joined', invitationToken: joined.invitation_token }
    }
    catch (err) {
      if (err instanceof ApiError && err.status === 409) return { kind: 'name-taken' }
      if (err instanceof ApiError && err.status === 423) return { kind: 'locked' }
      return { kind: 'failed' }
    }
    finally {
      submitting.value = false
    }
  }

  return { outcome, loading, submitting, load, claim, join }
}
