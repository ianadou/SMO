<script setup lang="ts">
import { computed } from 'vue'
import { ArrowLeft, MoreVertical } from 'lucide-vue-next'
import type { MatchDTO } from '~/types/matches'

const props = defineProps<{ match: MatchDTO }>()
const emit = defineEmits<{ back: [] }>()

const hasScore = computed(
  () => props.match.score_a != null && props.match.score_b != null,
)

const meta = computed(() => {
  const d = new Date(props.match.scheduled_at)
  const date = d.toLocaleDateString('fr-FR', {
    weekday: 'short',
    day: 'numeric',
    month: 'short',
  })
  const time = d.toLocaleTimeString('fr-FR', {
    hour: '2-digit',
    minute: '2-digit',
  })
  return `${date} · ${time} · ${props.match.venue}`
})
</script>

<template>
  <div class="md-header">
    <button class="md-icon-btn" aria-label="Retour" @click="emit('back')">
      <ArrowLeft :size="22" />
    </button>
    <div class="md-header-mid">
      <div class="md-vscard">
        <div class="md-vscard-team">
          <div class="md-vscard-jersey is-red" />
          <span class="md-vscard-label">Équipe rouge</span>
        </div>
        <div class="md-vscard-center">
          <template v-if="hasScore">{{ match.score_a }} – {{ match.score_b }}</template>
          <template v-else>VS</template>
        </div>
        <div class="md-vscard-team">
          <div class="md-vscard-jersey is-green" />
          <span class="md-vscard-label">Équipe verte</span>
        </div>
      </div>
      <div class="md-vscard-meta">{{ match.title }} · {{ meta }}</div>
    </div>
    <button class="md-icon-btn" aria-label="Plus d'options">
      <MoreVertical :size="22" />
    </button>
  </div>
</template>
