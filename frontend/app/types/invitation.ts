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
