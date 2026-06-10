<script setup lang="ts">
import { computed } from 'vue'
import { playerInitials } from '~/utils/playerInitials'
import type { TeamMemberDTO } from '~/types/matches'

const props = defineProps<{
  red: TeamMemberDTO[]
  green: TeamMemberDTO[]
}>()

const placed = computed(() => props.red.length + props.green.length)
</script>

<template>
  <section class="md-presents">
    <div class="md-presents-head">
      <span class="md-presents-title">Présents</span>
      <span class="md-presents-count">
        <span class="num">{{ placed }}</span>
        <span style="color: var(--fg-subtle)"> / </span>
        <span class="num" style="color: var(--fg-subtle)">{{ placed }}</span>
      </span>
    </div>
    <div class="md-chips">
      <span v-for="p in red" :key="p.player_id" class="md-chip is-red">
        <span class="md-chip-avatar">{{ playerInitials(p.player_name) }}</span>
        <span>{{ p.player_name }}</span>
      </span>
      <span v-for="p in green" :key="p.player_id" class="md-chip is-green">
        <span class="md-chip-avatar">{{ playerInitials(p.player_name) }}</span>
        <span>{{ p.player_name }}</span>
      </span>
    </div>
  </section>
</template>
