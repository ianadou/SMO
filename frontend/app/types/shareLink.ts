import type { MatchStatus } from '~/types/matches'

export type ShareRosterState = 'claimable' | 'claimed' | 'responded'

export interface ShareRosterEntry {
  player_id: string
  player_name: string
  state: ShareRosterState
}

export interface ShareLinkContext {
  match_id: string
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
  roster: ShareRosterEntry[]
}

export interface ClaimedInvitationDTO {
  invitation_token: string
}

export interface MatchShareLinkDTO {
  token: string
  url: string
  expires_at: string
}

export type ShareLinkLoadOutcome =
  | { kind: 'ok'; context: ShareLinkContext }
  | { kind: 'invalid' }
  | { kind: 'error' }

export type JoinView = 'roster' | 'locked' | 'invalid' | 'error'

export type ClaimOutcome =
  | { kind: 'claimed'; invitationToken: string }
  | { kind: 'race' }
  | { kind: 'locked' }
  | { kind: 'failed' }

export type JoinOutcome =
  | { kind: 'joined'; invitationToken: string }
  | { kind: 'name-taken' }
  | { kind: 'locked' }
  | { kind: 'failed' }
