import type { MatchStatus, TeamSide } from '~/types/matches'

export interface VoteContextVoter {
  player_id: string
  name: string
  initials: string
  team: TeamSide | ''
}

export interface VoteContextTeammate {
  player_id: string
  name: string
  initials: string
  matches_together: number
  your_score: number | null
}

export interface VoteTeammateResult {
  player_id: string
  name: string
  initials: string
  average: number
  votes_count: number
  delta: number | null
}

export interface VoteSelfResult {
  average: number | null
  votes_count: number
}

export interface VoteResults {
  teammates: VoteTeammateResult[]
  self: VoteSelfResult
}

export interface VoteContext {
  group_name: string
  match_title: string
  venue: string
  scheduled_at: string
  status: MatchStatus
  score_a: number | null
  score_b: number | null
  winner: TeamSide | null
  voter: VoteContextVoter
  teammates: VoteContextTeammate[]
  voters_done: number
  voters_total: number
  results: VoteResults | null
}

export interface CastVoteResult {
  voted_id: string
  score: number
}

export type VoteView =
  | 'rate'
  | 'submitted'
  | 'results'
  | 'too-early'
  | 'not-participant'
  | 'invalid'
  | 'error'

export type VoteLoadOutcome =
  | { kind: 'ok'; context: VoteContext }
  | { kind: 'invalid' }
  | { kind: 'not-participant' }
  | { kind: 'error' }
