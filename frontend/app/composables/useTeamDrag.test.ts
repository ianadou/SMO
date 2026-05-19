import { describe, it, expect } from 'vitest'
import { ref } from 'vue'
import { useTeamDrag } from './useTeamDrag'
import type { TeamMemberDTO } from '~/types/matches'

function member(id: string, team: 'A' | 'B', slot: number): TeamMemberDTO {
  return { player_id: id, player_name: id.toUpperCase(), team, slot }
}

function setup() {
  const teamA = ref<TeamMemberDTO[]>([
    member('a1', 'A', 0),
    member('a2', 'A', 1),
  ])
  const teamB = ref<TeamMemberDTO[]>([
    member('b1', 'B', 0),
    member('b2', 'B', 1),
  ])
  return { teamA, teamB, drag: useTeamDrag(teamA, teamB) }
}

describe('useTeamDrag.swap', () => {
  it('exchanges membership across teams keeping the other slot index', () => {
    const { teamA, teamB, drag } = setup()

    drag.swap('a1', 'b2')

    expect(teamA.value.map((m) => m.player_id)).toEqual(['b2', 'a2'])
    expect(teamB.value.map((m) => m.player_id)).toEqual(['b1', 'a1'])
  })

  it('reorders within the same team when both players share a side', () => {
    const { teamA, teamB, drag } = setup()

    drag.swap('a1', 'a2')

    expect(teamA.value.map((m) => m.player_id)).toEqual(['a2', 'a1'])
    expect(teamB.value.map((m) => m.player_id)).toEqual(['b1', 'b2'])
  })

  it('does nothing when one id is unknown', () => {
    const { teamA, teamB, drag } = setup()

    drag.swap('a1', 'ghost')

    expect(teamA.value.map((m) => m.player_id)).toEqual(['a1', 'a2'])
    expect(teamB.value.map((m) => m.player_id)).toEqual(['b1', 'b2'])
  })
})

describe('useTeamDrag.finishDrag', () => {
  it('snaps back without swapping when dropped on empty space', () => {
    const { teamA, teamB, drag } = setup()
    drag.beginDrag('a1')

    drag.finishDrag(null)

    expect(teamA.value.map((m) => m.player_id)).toEqual(['a1', 'a2'])
    expect(teamB.value.map((m) => m.player_id)).toEqual(['b1', 'b2'])
    expect(drag.drag.value).toBeNull()
  })

  it('swaps when released over another pion then clears drag', () => {
    const { teamA, teamB, drag } = setup()
    drag.beginDrag('a1')

    drag.finishDrag('b1')

    expect(teamA.value.map((m) => m.player_id)).toEqual(['b1', 'a2'])
    expect(teamB.value.map((m) => m.player_id)).toEqual(['a1', 'b2'])
    expect(drag.drag.value).toBeNull()
  })
})
