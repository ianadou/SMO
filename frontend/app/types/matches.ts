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
  created_at: string
}

export interface CreateMatchPayload {
  group_id: string
  title: string
  venue: string
  scheduled_at: string
}
