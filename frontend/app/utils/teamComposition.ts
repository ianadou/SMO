import type { MatchScreen, MatchStatus, TeamMemberDTO } from '~/types/matches'

export function deriveScreen(status: MatchStatus, hasTeams: boolean): MatchScreen {
  if (status === 'draft') return 'setup-draft'
  if (status === 'open') return hasTeams ? 'composition' : 'setup-generate'
  if (status === 'teams_ready' || status === 'in_progress') return 'finished'
  return 'closed'
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
