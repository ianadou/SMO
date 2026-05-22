<script setup lang="ts">
import { computed } from 'vue'
import TeamMatchupCard from '~/components/matches/TeamMatchupCard.vue'
import type { MatchStatus } from '~/types/matches'

const props = defineProps<{
  id: string
  scheduledAt: string
  status: MatchStatus
  scoreA?: number | null
  scoreB?: number | null
}>()

const cardStatus = computed<'upcoming' | 'live' | 'finished' | 'closed'>(() => {
  switch (props.status) {
    case 'in_progress':
      return 'live'
    case 'completed':
      return 'finished'
    case 'closed':
      return 'closed'
    default:
      return 'upcoming'
  }
})

const winner = computed<'red' | 'green' | undefined>(() => {
  if (props.scoreA == null || props.scoreB == null) return undefined
  if (props.scoreA > props.scoreB) return 'red'
  if (props.scoreA < props.scoreB) return 'green'
  return undefined
})

const scheduled = computed(() => new Date(props.scheduledAt))

const dateLabel = computed(() => {
  const raw = scheduled.value.toLocaleDateString('fr-FR', {
    weekday: 'short',
    day: 'numeric',
    month: 'short',
  })
  return raw.charAt(0).toUpperCase() + raw.slice(1)
})

const timeLabel = computed(() =>
  scheduled.value.toLocaleTimeString('fr-FR', {
    hour: '2-digit',
    minute: '2-digit',
  }),
)
</script>

<template>
  <NuxtLink :to="`/matches/${id}`" class="match-card-link" :aria-label="`Match du ${dateLabel}`">
    <TeamMatchupCard
      :status="cardStatus"
      :date-label="dateLabel"
      :time-label="timeLabel"
      :winner="winner"
    />
  </NuxtLink>
</template>

<style scoped>
.match-card-link {
  display: block;
  text-decoration: none;
  color: inherit;
  border-radius: var(--radius-lg);
  transition: transform var(--motion-fast) var(--motion-easing);
}
.match-card-link:hover :deep(.mc) { background: #1F252D; }
.match-card-link:active { transform: scale(0.995); }
.match-card-link:focus-visible {
  outline: none;
  box-shadow: 0 0 0 2px var(--color-bg-base), 0 0 0 4px rgba(32, 128, 255, 0.45);
}
</style>
