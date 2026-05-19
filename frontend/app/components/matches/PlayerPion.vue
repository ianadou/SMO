<script setup lang="ts">
import { computed } from 'vue'
import type { TeamMemberDTO } from '~/types/matches'

const props = defineProps<{
  player: TeamMemberDTO
  team: 'red' | 'green'
  x: number
  y: number
  mode: 'edit' | 'view'
  score?: number | null
  isDragging?: boolean
  isDropTarget?: boolean
  dragOffset?: { dx: number; dy: number } | null
}>()

const emit = defineEmits<{
  pointerdown: [player: TeamMemberDTO, evt: PointerEvent]
}>()

const initials = computed(() => {
  const parts = props.player.player_name.trim().split(/\s+/)
  const letters = parts.length > 1 ? parts[0]![0]! + parts[parts.length - 1]![0]! : props.player.player_name.slice(0, 2)
  return letters.toUpperCase()
})

const label = computed(() => {
  const n = props.player.player_name
  return n.length > 9 ? n.slice(0, 8) + '…' : n
})

const interactive = computed(() => props.mode === 'edit' && !props.isDragging)

const style = computed(() => {
  const base: Record<string, string> = { left: `${props.x}%`, top: `${props.y}%` }
  if (props.dragOffset) {
    base.transform = `translate(calc(-50% + ${props.dragOffset.dx}px), calc(-50% + ${props.dragOffset.dy}px))`
  }
  return base
})

const classes = computed(() => [
  'md-pion',
  props.team === 'green' ? 'md-pion-green' : 'md-pion-red',
  interactive.value ? 'is-interactive' : '',
  props.isDragging ? 'is-dragging' : '',
  props.isDropTarget ? 'is-drop-target' : '',
])

function onPointerDown(evt: PointerEvent) {
  if (props.mode !== 'edit') return
  emit('pointerdown', props.player, evt)
}
</script>

<template>
  <div :class="classes" :style="style" :data-pion-id="player.player_id" @pointerdown="onPointerDown">
    <div class="md-pion-disc">{{ initials }}</div>
    <div class="md-pion-name">{{ label }}</div>
    <div v-if="score != null" class="md-pion-score">{{ score.toFixed(1) }}</div>
  </div>
</template>
