const STASH_KEY_PREFIX = 'smo.invitation.'

export interface StashedInvitation {
  token: string
  playerName: string
}

export function stashInvitation(matchId: string, invitation: StashedInvitation) {
  if (typeof window === 'undefined') return
  window.localStorage.setItem(STASH_KEY_PREFIX + matchId, JSON.stringify(invitation))
}

export function readStashedInvitation(matchId: string): StashedInvitation | null {
  if (typeof window === 'undefined') return null
  const raw = window.localStorage.getItem(STASH_KEY_PREFIX + matchId)
  if (!raw) return null
  try {
    const parsed = JSON.parse(raw) as Partial<StashedInvitation>
    if (typeof parsed.token !== 'string' || typeof parsed.playerName !== 'string') return null
    return { token: parsed.token, playerName: parsed.playerName }
  }
  catch {
    return null
  }
}
