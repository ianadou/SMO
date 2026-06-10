import type { MatchScreen, MatchStatus, TeamMemberDTO } from '~/types/matches'

export const TEAM_LOCK_LEAD_TIME_MS = 10 * 60 * 1000

export function deriveScreen(
  status: MatchStatus,
  hasTeams: boolean,
  now: Date,
  scheduledAt: string,
): MatchScreen {
  if (status === 'draft') return 'setup-draft'
  if (status === 'open' && !hasTeams) return 'setup-generate'
  if (status === 'open' || status === 'teams_ready') {
    return isTeamLockReached(now, scheduledAt) ? 'locked' : 'composition'
  }
  if (status === 'in_progress') return 'finished'
  return 'closed'
}

function isTeamLockReached(now: Date, scheduledAt: string): boolean {
  return now.getTime() >= new Date(scheduledAt).getTime() - TEAM_LOCK_LEAD_TIME_MS
}

export function splitTeams(members: TeamMemberDTO[]): {
  teamA: TeamMemberDTO[]
  teamB: TeamMemberDTO[]
} {
  const bySlot = (a: TeamMemberDTO, b: TeamMemberDTO) => a.slot - b.slot
  return {
    teamA: members.filter((m) => m.team === 'A').sort(bySlot),
    teamB: members.filter((m) => m.team === 'B').sort(bySlot),
  }
}

export function toTeamArrays(
  teamA: TeamMemberDTO[],
  teamB: TeamMemberDTO[],
): { team_a: string[]; team_b: string[] } {
  return {
    team_a: teamA.map((m) => m.player_id),
    team_b: teamB.map((m) => m.player_id),
  }
}
