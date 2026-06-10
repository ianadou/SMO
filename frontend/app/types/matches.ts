export type MatchStatus =
  | 'draft'
  | 'open'
  | 'teams_ready'
  | 'in_progress'
  | 'completed'
  | 'closed'

export interface MatchDTO {
  id: string
  group_id: string
  title: string
  venue: string
  scheduled_at: string
  status: MatchStatus
  score_a: number | null
  score_b: number | null
  created_at: string
}

export type TeamSide = 'A' | 'B'

export interface TeamMemberDTO {
  player_id: string
  player_name: string
  team: TeamSide
  slot: number
}

export type MatchScreen =
  | 'setup-draft'
  | 'setup-generate'
  | 'composition'
  | 'finished'
  | 'closed'

export interface CreateMatchPayload {
  group_id: string
  title: string
  venue: string
  scheduled_at: string
}
