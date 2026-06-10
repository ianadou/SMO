import type { MatchStatus } from '~/types/matches'

export interface InvitationContext {
  organizer_name: string
  group_name: string
  match_title: string
  venue: string
  scheduled_at: string
  match_status: MatchStatus
  capacity: string
  confirmed_count: number
  max_participants: number
  confirmed_initials: string[]
  response: 'pending' | 'yes' | 'no'
  expires_at: string
  state: 'respondable' | 'locked' | 'expired'
}

export interface RespondResult {
  response: 'yes' | 'no'
  responded_at: string
}

export type InviteView =
  | 'initial'
  | 'result'
  | 'expired'
  | 'invalid'
  | 'locked-pending'
  | 'error'

export type LoadOutcome =
  | { kind: 'ok'; context: InvitationContext }
  | { kind: 'invalid' }
  | { kind: 'error' }

export interface MatchInvitationDTO {
  id: string
  match_id: string
  player_id: string
  expires_at: string
  response: 'pending' | 'yes' | 'no'
  responded_at: string | null
  created_at: string
}

export interface CreatedInvitationDTO extends MatchInvitationDTO {
  plain_token: string
}

export type InviteRowStatus = 'not-invited' | 'fresh' | 'pending' | 'yes' | 'no'

export interface InviteRow {
  playerId: string
  playerName: string
  status: InviteRowStatus
  shareUrl: string | null
}
