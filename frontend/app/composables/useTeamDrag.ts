import { ref, type Ref } from 'vue'
import type { TeamMemberDTO } from '~/types/matches'

export interface DragState {
  id: string
  dx: number
  dy: number
  dropTargetId: string | null
}

interface Located {
  side: 'A' | 'B'
  index: number
}

export function useTeamDrag(
  teamA: Ref<TeamMemberDTO[]>,
  teamB: Ref<TeamMemberDTO[]>,
) {
  const drag = ref<DragState | null>(null)

  function locate(id: string): Located | null {
    const a = teamA.value.findIndex((m) => m.player_id === id)
    if (a !== -1) return { side: 'A', index: a }
    const b = teamB.value.findIndex((m) => m.player_id === id)
    if (b !== -1) return { side: 'B', index: b }
    return null
  }

  function swap(idA: string, idB: string) {
    if (idA === idB) return
    const a = locate(idA)
    const b = locate(idB)
    if (!a || !b) return

    if (a.side === b.side) {
      const list = a.side === 'A' ? teamA.value.slice() : teamB.value.slice()
      ;[list[a.index], list[b.index]] = [list[b.index], list[a.index]]
      if (a.side === 'A') teamA.value = list
      else teamB.value = list
      return
    }

    const nextA = teamA.value.slice()
    const nextB = teamB.value.slice()
    if (a.side === 'A') {
      nextA[a.index] = teamB.value[b.index]
      nextB[b.index] = teamA.value[a.index]
    } else {
      nextB[a.index] = teamA.value[b.index]
      nextA[b.index] = teamB.value[a.index]
    }
    teamA.value = nextA
    teamB.value = nextB
  }

  function beginDrag(id: string) {
    drag.value = { id, dx: 0, dy: 0, dropTargetId: null }
  }

  function finishDrag(dropTargetId: string | null) {
    const current = drag.value
    drag.value = null
    if (current && dropTargetId && dropTargetId !== current.id) {
      swap(current.id, dropTargetId)
    }
  }

  function onPointerDown(player: TeamMemberDTO, evt: PointerEvent) {
    evt.preventDefault()
    const target = evt.currentTarget as HTMLElement
    target.setPointerCapture?.(evt.pointerId)
    const rect = target.getBoundingClientRect()
    beginDrag(player.player_id)

    const onMove = (e: PointerEvent) => {
      const dx = e.clientX - rect.left - rect.width / 2
      const dy = e.clientY - rect.top - rect.height / 2
      let dropTargetId: string | null = null
      for (const el of document.elementsFromPoint(e.clientX, e.clientY)) {
        const candidate = (el as HTMLElement).closest?.('.md-pion') as
          | HTMLElement
          | null
        if (candidate && candidate.dataset.pionId !== player.player_id) {
          dropTargetId = candidate.dataset.pionId ?? null
          break
        }
      }
      if (drag.value) drag.value = { ...drag.value, dx, dy, dropTargetId }
    }
    const onUp = () => {
      const dropTargetId = drag.value?.dropTargetId ?? null
      finishDrag(dropTargetId)
      window.removeEventListener('pointermove', onMove)
      window.removeEventListener('pointerup', onUp)
      window.removeEventListener('pointercancel', onUp)
    }
    window.addEventListener('pointermove', onMove)
    window.addEventListener('pointerup', onUp)
    window.addEventListener('pointercancel', onUp)
  }

  return { drag, swap, beginDrag, finishDrag, onPointerDown }
}
