<script setup lang="ts">
import { computed, onMounted, ref, watch } from 'vue'
import { MapPin } from 'lucide-vue-next'
import TeamMatchupCard from '~/components/matches/TeamMatchupCard.vue'
import { useApi } from '~/composables/useApi'
import type { MatchDTO } from '~/types/matches'

const props = defineProps<{ match: MatchDTO }>()

const api = useApi()
const confirmedCount = ref<number | null>(null)
const maxParticipants = 10

async function loadParticipants(matchId: string) {
  try {
    const list = await api.get<Array<{ player_id: string }>>(
      `/matches/${matchId}/participants`,
    )
    confirmedCount.value = list.length
  } catch {
    confirmedCount.value = null
  }
}

onMounted(() => loadParticipants(props.match.id))
watch(() => props.match.id, (id) => loadParticipants(id))

const scheduled = computed(() => new Date(props.match.scheduled_at))

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
  <section class="next-match">
    <div class="next-match-eyebrow">Prochain match</div>
    <NuxtLink :to="`/matches/${match.id}`" class="next-match-link" aria-label="Voir le match">
      <TeamMatchupCard
        status="upcoming"
        :date-label="dateLabel"
        :time-label="timeLabel"
        :jersey-size="56"
      />
      <div class="next-match-foot">
        <span class="next-match-venue">
          <MapPin :size="14" />
          <span>{{ match.venue }}</span>
        </span>
        <span v-if="confirmedCount != null" class="next-match-count">
          <strong>{{ confirmedCount }}</strong>/{{ maxParticipants }} confirmés
        </span>
      </div>
    </NuxtLink>
  </section>
</template>

<style scoped>
.next-match {
  display: flex;
  flex-direction: column;
  gap: var(--space-2);
}
.next-match-eyebrow {
  font-family: var(--font-mono);
  font-size: 11px;
  letter-spacing: 0.08em;
  text-transform: uppercase;
  color: var(--color-fg-muted);
  padding: 0 var(--space-1);
}

.next-match-link {
  display: block;
  text-decoration: none;
  color: inherit;
  border-radius: var(--radius-lg);
  transition: transform var(--motion-fast) var(--motion-easing);
  background: var(--color-bg-elevated);
  padding-bottom: var(--space-3);
}
.next-match-link:hover :deep(.mc) { background: transparent; }
.next-match-link:hover { background: #1F252D; }
.next-match-link:active { transform: scale(0.997); }
.next-match-link:focus-visible {
  outline: none;
  box-shadow: 0 0 0 2px var(--color-bg-base), 0 0 0 4px rgba(32, 128, 255, 0.45);
}
.next-match-link :deep(.mc) {
  background: transparent;
  border-radius: var(--radius-lg) var(--radius-lg) 0 0;
}

.next-match-foot {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: var(--space-3);
  padding: 0 var(--space-4);
  font-size: 13px;
  color: var(--color-fg-muted);
}
.next-match-venue { display: inline-flex; align-items: center; gap: 6px; }
.next-match-venue :deep(svg) { color: var(--color-fg-subtle); }
.next-match-count {
  font-family: var(--font-mono);
  font-variant-numeric: tabular-nums;
  letter-spacing: 0.02em;
}
.next-match-count strong {
  color: var(--color-fg-default);
  font-weight: 600;
}
</style>
