import { computed, ref } from 'vue'
import { useApi } from './useApi'
import type { PlayerDTO } from '~/types/groups'
import type { CreatedInvitationDTO, InviteRow, MatchInvitationDTO } from '~/types/invitation'
import type { MatchShareLinkDTO } from '~/types/shareLink'

export interface InvitationApi {
  get: <T>(path: string) => Promise<T>
  post: <T>(path: string, body: unknown) => Promise<T>
  delete: <T>(path: string) => Promise<T>
}

export function useMatchInvitations(matchId: string, api: InvitationApi = useApi()) {
  const players = ref<PlayerDTO[]>([])
  const invitations = ref<MatchInvitationDTO[]>([])
  const freshLinks = ref<Record<string, string>>({})
  const loading = ref(false)
  const invitingId = ref<string | null>(null)
  const error = ref<boolean>(false)
  const shareLink = ref<MatchShareLinkDTO | null>(null)
  const linkBusy = ref(false)

  const rows = computed<InviteRow[]>(() =>
    players.value.map((player) => {
      const invitation = invitations.value.find(inv => inv.player_id === player.id)
      const claimed = Boolean(invitation?.claimed_at)
      const freshUrl = freshLinks.value[player.id]
      if (freshUrl) {
        return { playerId: player.id, playerName: player.name, status: 'fresh' as const, shareUrl: freshUrl, claimed }
      }
      return {
        playerId: player.id,
        playerName: player.name,
        status: invitation ? invitation.response : 'not-invited',
        shareUrl: null,
        claimed,
      }
    }),
  )

  const confirmedCount = computed(
    () => invitations.value.filter(inv => inv.response === 'yes').length,
  )

  async function load(groupId: string) {
    loading.value = true
    error.value = false
    try {
      const [groupPlayers, matchInvitations] = await Promise.all([
        api.get<PlayerDTO[]>(`/groups/${groupId}/players`),
        api.get<MatchInvitationDTO[]>(`/matches/${matchId}/invitations`),
      ])
      players.value = groupPlayers ?? []
      invitations.value = matchInvitations ?? []
    }
    catch {
      error.value = true
    }
    finally {
      loading.value = false
    }
  }

  async function invite(playerId: string): Promise<string | null> {
    invitingId.value = playerId
    try {
      const created = await api.post<CreatedInvitationDTO>('/invitations', {
        match_id: matchId,
        player_id: playerId,
      })
      const shareUrl = `${window.location.origin}/invite/${created.plain_token}`
      freshLinks.value = { ...freshLinks.value, [playerId]: shareUrl }
      invitations.value = [...invitations.value, created]
      return shareUrl
    }
    catch {
      return null
    }
    finally {
      invitingId.value = null
    }
  }

  async function generateShareLink(): Promise<boolean> {
    linkBusy.value = true
    try {
      const created = await api.post<MatchShareLinkDTO>(`/matches/${matchId}/share-link`, {})
      const url = created.url || `${window.location.origin}/join/${created.token}`
      shareLink.value = { ...created, url }
      return true
    }
    catch {
      return false
    }
    finally {
      linkBusy.value = false
    }
  }

  async function revokeShareLink(): Promise<boolean> {
    linkBusy.value = true
    try {
      await api.delete(`/matches/${matchId}/share-link`)
      shareLink.value = null
      return true
    }
    catch {
      return false
    }
    finally {
      linkBusy.value = false
    }
  }

  return {
    rows,
    confirmedCount,
    loading,
    invitingId,
    error,
    shareLink,
    linkBusy,
    load,
    invite,
    generateShareLink,
    revokeShareLink,
  }
}
