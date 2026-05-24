<script setup lang="ts">
import { computed } from 'vue'
import { ArrowLeft, MapPin, MoreVertical } from 'lucide-vue-next'
import TeamMatchupCard from './TeamMatchupCard.vue'
import type { MatchDTO } from '~/types/matches'

const props = defineProps<{ match: MatchDTO }>()
const emit = defineEmits<{ back: [] }>()

const cardStatus = computed<'upcoming' | 'live' | 'finished' | 'closed'>(() => {
  switch (props.match.status) {
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
  const a = props.match.score_a
  const b = props.match.score_b
  if (a == null || b == null) return undefined
  if (a > b) return 'red'
  if (b > a) return 'green'
  return undefined
})

const scheduled = computed(() => new Date(props.match.scheduled_at))

const dateLabel = computed(() => {
  const d = scheduled.value
  const raw = d.toLocaleDateString('fr-FR', {
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
  <header class="md-header-block">
    <div class="md-header">
      <button class="md-icon-btn" aria-label="Retour" @click="emit('back')">
        <ArrowLeft :size="22" />
      </button>
      <div class="md-header-mid md-match-card">
        <TeamMatchupCard
          :status="cardStatus"
          :date-label="dateLabel"
          :time-label="timeLabel"
          :winner="winner"
        />
      </div>
      <button class="md-icon-btn" aria-label="Plus d'options">
        <MoreVertical :size="22" />
      </button>
    </div>
    <div class="md-header-venue">
      <MapPin :size="14" />
      <span>{{ match.venue }}</span>
    </div>
  </header>
</template>

<style scoped>
.md-header-mid { flex: 1; min-width: 0; }
.md-match-card :deep(.mc) {
  padding: var(--space-3);
  gap: var(--space-2);
}
.md-match-card :deep(.mc-team-label) { font-size: 12px; }
.md-match-card :deep(.mc-center-vs) { font-size: 22px; }
.md-match-card :deep(.sjersey) { width: 36px !important; height: 36px !important; }
.md-match-card :deep(.mc-center-trophy) { width: 28px !important; height: 28px !important; }

.md-header-block { display: flex; flex-direction: column; }
.md-header-venue {
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 6px;
  padding: 0 var(--space-3) var(--space-2);
  font-size: 12px;
  color: var(--color-fg-muted);
}
.md-header-venue :deep(svg) { color: var(--color-fg-subtle); }
</style>
