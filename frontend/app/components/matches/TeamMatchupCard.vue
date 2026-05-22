<script setup lang="ts">
import { XCircle } from 'lucide-vue-next'
import Jersey from '~/components/ui/Jersey.vue'

const props = withDefaults(
  defineProps<{
    status: 'upcoming' | 'live' | 'finished' | 'closed' | 'cancelled'
    dateLabel: string
    timeLabel?: string
    winner?: 'red' | 'green'
    votesLabel?: string
    redLabel?: string
    greenLabel?: string
    jerseySize?: number
    interactive?: boolean
    ariaLabel?: string
  }>(),
  {
    redLabel: 'Équipe rouge',
    greenLabel: 'Équipe verte',
    jerseySize: 52,
    interactive: false,
  },
)

defineEmits<{ click: [] }>()

const trophySize = Math.round(props.jerseySize * (32 / 52))
</script>

<template>
  <component
    :is="interactive ? 'button' : 'div'"
    :type="interactive ? 'button' : undefined"
    :class="['mc', { 'is-cancelled': status === 'cancelled', 'is-static': !interactive }]"
    :aria-label="ariaLabel"
    @click="interactive && $emit('click')"
  >
    <div class="mc-context">
      <template v-if="status === 'live'">
        <span class="mc-pill mc-pill-live">
          <span class="mc-dot" aria-hidden="true" />
          <span>{{ dateLabel }}</span>
          <span class="mc-context-sep">·</span>
          <span>En cours</span>
        </span>
      </template>
      <template v-else-if="status === 'cancelled'">
        <span class="mc-pill mc-pill-cancelled">
          <span>{{ dateLabel }}</span>
          <span class="mc-context-sep">·</span>
          <span>Annulé</span>
        </span>
      </template>
      <template v-else>
        <span>{{ dateLabel }}</span>
        <span class="mc-context-sep">·</span>
        <span v-if="status === 'upcoming'" class="num">{{ timeLabel }}</span>
        <span v-else-if="status === 'finished'">Terminé</span>
        <span v-else-if="status === 'closed'">Clôturé</span>
      </template>
    </div>

    <div class="mc-vs">
      <div class="mc-team">
        <Jersey color="red" :size="jerseySize" :aria-label="redLabel" />
        <span class="mc-team-label">{{ redLabel }}</span>
      </div>

      <div class="mc-center">
        <template v-if="status === 'finished' || status === 'closed'">
          <svg
            :class="['mc-center-trophy', winner === 'green' ? 'is-green' : 'is-red']"
            :width="trophySize"
            :height="trophySize"
            viewBox="0 0 24 24"
            xmlns="http://www.w3.org/2000/svg"
            role="img"
            :aria-label="`Vainqueur : équipe ${winner === 'green' ? 'verte' : 'rouge'}`"
          >
            <path d="M5 4 L5 8 C5 9.5 6 10.7 7.4 11.1 L7.4 9 C6.6 8.7 6.2 8.1 6.2 7.4 L6.2 5.2 L7.5 5.2 L7.5 4 Z" fill="currentColor" />
            <path d="M19 4 L19 8 C19 9.5 18 10.7 16.6 11.1 L16.6 9 C17.4 8.7 17.8 8.1 17.8 7.4 L17.8 5.2 L16.5 5.2 L16.5 4 Z" fill="currentColor" />
            <path d="M7 3 L17 3 L17 9 C17 11.76 14.76 14 12 14 C9.24 14 7 11.76 7 9 Z" fill="currentColor" />
            <rect x="11" y="14" width="2" height="4" fill="currentColor" />
            <path d="M7 19 L17 19 L17 21 L7 21 Z" fill="currentColor" />
            <rect x="9" y="17.5" width="6" height="2" rx="0.5" fill="currentColor" />
          </svg>
        </template>
        <template v-else-if="status === 'cancelled'">
          <XCircle :size="trophySize" class="mc-center-cancelled" aria-hidden="true" />
        </template>
        <template v-else>
          <span class="mc-center-vs">VS</span>
        </template>
      </div>

      <div class="mc-team">
        <Jersey color="green" :size="jerseySize" :aria-label="greenLabel" />
        <span class="mc-team-label">{{ greenLabel }}</span>
      </div>
    </div>

    <div v-if="status === 'closed' && votesLabel" class="mc-foot">
      <span>{{ votesLabel }}</span>
    </div>
  </component>
</template>

<style scoped>
.mc {
  appearance: none;
  background: var(--color-bg-elevated);
  color: var(--color-fg-default);
  border: 0;
  width: 100%;
  padding: var(--space-4);
  border-radius: var(--radius-lg);
  font-family: var(--font-sans);
  text-align: left;
  display: flex;
  flex-direction: column;
  gap: var(--space-3);
  transition: background var(--motion-default) var(--motion-easing),
    transform var(--motion-fast) var(--motion-easing);
}
.mc.is-static { cursor: default; }
button.mc { cursor: pointer; }
button.mc:hover { background: #1F252D; }
button.mc:active { background: #1F252D; transform: scale(0.995); }
button.mc:focus-visible {
  outline: none;
  box-shadow: 0 0 0 2px var(--color-bg-base), 0 0 0 4px rgba(32, 128, 255, 0.45);
}

.mc-context {
  display: flex; align-items: center; justify-content: center;
  font-size: 13px; line-height: 1.2;
  color: var(--color-fg-muted);
}
.mc-context .num { font-family: var(--font-mono); }
.mc-context-sep {
  display: inline-block; margin: 0 8px;
  color: var(--color-fg-subtle);
}

.mc-pill {
  display: inline-flex; align-items: center; gap: 6px;
  padding: 4px 10px;
  border-radius: 999px;
  font-size: 13px; line-height: 1.2;
  white-space: nowrap;
}
.mc-pill-live { background: rgba(32, 128, 255, 0.15); color: #9CC4FF; }
.mc-pill-live .mc-dot {
  width: 6px; height: 6px; border-radius: 999px;
  background: var(--color-action-primary, #2080ff);
  box-shadow: 0 0 0 0 rgba(32, 128, 255, 0.6);
  animation: mc-pulse 1.6s ease-out infinite;
}
@keyframes mc-pulse {
  0% { box-shadow: 0 0 0 0 rgba(32, 128, 255, 0.55); }
  70% { box-shadow: 0 0 0 6px rgba(32, 128, 255, 0); }
  100% { box-shadow: 0 0 0 0 rgba(32, 128, 255, 0); }
}
.mc-pill-cancelled {
  background: rgba(255, 255, 255, 0.06);
  color: var(--color-fg-subtle);
  text-decoration: line-through;
  text-decoration-thickness: 1px;
  text-decoration-color: var(--color-fg-subtle);
}

.mc-vs {
  display: grid;
  grid-template-columns: 1fr auto 1fr;
  align-items: center;
  gap: var(--space-3);
}
.mc-team {
  display: flex; flex-direction: column; align-items: center; gap: 8px;
  min-width: 0;
}
.mc-team-label {
  font-size: 13px; font-weight: 500;
  color: var(--color-fg-default);
  letter-spacing: -0.005em;
  text-align: center;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
  max-width: 100%;
}

.mc-center {
  display: flex; flex-direction: column; align-items: center; justify-content: center;
  gap: 4px;
  min-width: 64px;
}
.mc-center-vs {
  font-family: var(--font-sans);
  font-size: 26px; font-weight: 700; letter-spacing: -0.02em;
  color: var(--color-fg-default);
  line-height: 1;
}
.mc-center-trophy { display: block; }
.mc-center-trophy.is-red { color: var(--color-team-red); }
.mc-center-trophy.is-green { color: var(--color-team-green); }
.mc-center-cancelled { color: var(--color-fg-subtle); display: block; }

.mc-foot {
  display: flex; justify-content: center;
  font-size: 12px;
  color: var(--color-fg-muted);
  margin-top: -2px;
}

.mc.is-cancelled :deep(.sjersey) { opacity: 0.4; }
.mc.is-cancelled .mc-team-label { color: var(--color-fg-subtle); }
</style>
