import { describe, it, expect, vi, afterEach } from 'vitest'
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

describe('useTeamDrag.onPointerDown', () => {
  afterEach(() => {
    vi.restoreAllMocks()
    delete (document as unknown as { elementsFromPoint?: unknown })
      .elementsFromPoint
  })

  function stubElementsFromPoint(result: Element[]) {
    ;(
      document as unknown as { elementsFromPoint: () => Element[] }
    ).elementsFromPoint = () => result
  }

  function pointerEvent(type: string, x: number, y: number): PointerEvent {
    const e = new Event(type) as PointerEvent
    Object.assign(e, { clientX: x, clientY: y, pointerId: 1 })
    return e
  }

  it('swaps on drop over another pion and removes its window listeners', () => {
    const { teamA, teamB, drag } = setup()
    const removeSpy = vi.spyOn(window, 'removeEventListener')

    const dropPion = document.createElement('div')
    dropPion.className = 'md-pion'
    dropPion.dataset.pionId = 'b1'
    stubElementsFromPoint([dropPion])

    const handle = document.createElement('div')
    const down = {
      preventDefault: () => {},
      currentTarget: handle,
      pointerId: 1,
    } as unknown as PointerEvent

    drag.onPointerDown(teamA.value[0]!, down)
    expect(drag.drag.value?.id).toBe('a1')

    window.dispatchEvent(pointerEvent('pointermove', 5, 5))
    expect(drag.drag.value?.dropTargetId).toBe('b1')

    window.dispatchEvent(pointerEvent('pointerup', 5, 5))

    expect(teamA.value.map((m) => m.player_id)).toEqual(['b1', 'a2'])
    expect(teamB.value.map((m) => m.player_id)).toEqual(['a1', 'b2'])
    expect(drag.drag.value).toBeNull()
    expect(removeSpy).toHaveBeenCalledWith('pointermove', expect.any(Function))
    expect(removeSpy).toHaveBeenCalledWith('pointerup', expect.any(Function))
    expect(removeSpy).toHaveBeenCalledWith('pointercancel', expect.any(Function))
  })

  it('snaps back without swapping when released over empty space', () => {
    const { teamA, teamB, drag } = setup()
    stubElementsFromPoint([])

    const handle = document.createElement('div')
    const down = {
      preventDefault: () => {},
      currentTarget: handle,
      pointerId: 1,
    } as unknown as PointerEvent

    drag.onPointerDown(teamA.value[0]!, down)
    window.dispatchEvent(pointerEvent('pointermove', 5, 5))
    window.dispatchEvent(pointerEvent('pointerup', 5, 5))

    expect(teamA.value.map((m) => m.player_id)).toEqual(['a1', 'a2'])
    expect(teamB.value.map((m) => m.player_id)).toEqual(['b1', 'b2'])
    expect(drag.drag.value).toBeNull()
  })
})
