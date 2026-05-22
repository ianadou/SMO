<script setup lang="ts">
import { computed, onBeforeUnmount, onMounted, ref } from 'vue'
import PlayerPion from './PlayerPion.vue'
import type { DragState } from '~/composables/useTeamDrag'
import type { TeamMemberDTO } from '~/types/matches'

const props = defineProps<{
  teamA: TeamMemberDTO[]
  teamB: TeamMemberDTO[]
  mode: 'edit' | 'view'
  drag?: DragState | null
}>()

const emit = defineEmits<{
  pointerdown: [player: TeamMemberDTO, evt: PointerEvent]
}>()

const fieldEl = ref<HTMLElement | null>(null)
const orientation = ref<'portrait' | 'landscape'>('portrait')
let observer: ResizeObserver | null = null

onMounted(() => {
  const el = fieldEl.value
  if (!el || typeof ResizeObserver === 'undefined') return
  observer = new ResizeObserver(() => {
    const r = el.getBoundingClientRect()
    orientation.value = r.width > r.height ? 'landscape' : 'portrait'
  })
  observer.observe(el)
})
onBeforeUnmount(() => observer?.disconnect())

function buildSlots(side: 'red' | 'green') {
  const rows = [
    { depth: 0.13, cross: 0 },
    { depth: 0.3, cross: -0.18 },
    { depth: 0.3, cross: 0.18 },
    { depth: 0.45, cross: -0.22 },
    { depth: 0.45, cross: 0.22 },
  ]
  return rows.map(({ depth, cross }) => {
    if (orientation.value === 'portrait') {
      const y = side === 'red' ? depth * 50 : 100 - depth * 50
      return { x: 50 + cross * 100, y }
    }
    const x = side === 'red' ? depth * 50 : 100 - depth * 50
    return { x, y: 50 + cross * 100 }
  })
}

const redSlots = computed(() => buildSlots('red'))
const greenSlots = computed(() => buildSlots('green'))

function offsetFor(id: string) {
  return props.drag && props.drag.id === id
    ? { dx: props.drag.dx, dy: props.drag.dy }
    : null
}

function onPointerDown(player: TeamMemberDTO, evt: PointerEvent) {
  emit('pointerdown', player, evt)
}
</script>

<template>
  <div ref="fieldEl" class="md-field">
    <div class="md-field-line is-pbox-top" />
    <div class="md-field-line is-pbox-bottom" />
    <div class="md-field-line is-pspot-top" />
    <div class="md-field-line is-pspot-bottom" />
    <div class="md-field-line is-halfway" />
    <div class="md-field-line is-circle" />
    <div class="md-field-line is-center-spot" />

    <PlayerPion
      v-for="(p, i) in teamA.slice(0, 5)"
      :key="p.player_id"
      :player="p"
      team="red"
      :x="redSlots[i]!.x"
      :y="redSlots[i]!.y"
      :mode="mode"
      :is-dragging="drag?.id === p.player_id"
      :is-drop-target="drag?.dropTargetId === p.player_id"
      :drag-offset="offsetFor(p.player_id)"
      @pointerdown="onPointerDown"
    />
    <PlayerPion
      v-for="(p, i) in teamB.slice(0, 5)"
      :key="p.player_id"
      :player="p"
      team="green"
      :x="greenSlots[i]!.x"
      :y="greenSlots[i]!.y"
      :mode="mode"
      :is-dragging="drag?.id === p.player_id"
      :is-drop-target="drag?.dropTargetId === p.player_id"
      :drag-offset="offsetFor(p.player_id)"
      @pointerdown="onPointerDown"
    />
  </div>
</template>
