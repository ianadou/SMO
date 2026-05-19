<script setup lang="ts">
import { computed } from 'vue'
import type { TeamMemberDTO } from '~/types/matches'

const props = defineProps<{
  red: TeamMemberDTO[]
  green: TeamMemberDTO[]
}>()

const total = computed(() => props.red.length + props.green.length)

function initials(name: string) {
  const parts = name.trim().split(/\s+/)
  const letters =
    parts.length > 1
      ? parts[0]![0]! + parts[parts.length - 1]![0]!
      : name.slice(0, 2)
  return letters.toUpperCase()
}
</script>

<template>
  <section class="md-presents">
    <div class="md-presents-head">
      <span class="md-presents-title">Présents</span>
      <span class="md-presents-count">{{ total }}</span>
    </div>
    <div class="md-chips">
      <span v-for="p in red" :key="p.player_id" class="md-chip is-red">
        <span class="md-chip-avatar">{{ initials(p.player_name) }}</span>
        <span>{{ p.player_name }}</span>
      </span>
      <span v-for="p in green" :key="p.player_id" class="md-chip is-green">
        <span class="md-chip-avatar">{{ initials(p.player_name) }}</span>
        <span>{{ p.player_name }}</span>
      </span>
    </div>
  </section>
</template>
