<script setup lang="ts">
import { ChevronRight, MapPin, Calendar } from 'lucide-vue-next'
import type { MatchStatus } from '~/types/matches'

defineProps<{
  id: string
  title: string
  venue: string
  scheduledAt: string
  status: MatchStatus
}>()

const dateFormatter = new Intl.DateTimeFormat('fr-FR', {
  day: 'numeric',
  month: 'long',
  hour: '2-digit',
  minute: '2-digit',
})

function formatScheduled(iso: string): string {
  const date = new Date(iso)
  if (Number.isNaN(date.getTime())) return ''
  return dateFormatter.format(date)
}

const statusLabels: Record<MatchStatus, string> = {
  draft: 'Brouillon',
  open: 'Ouvert',
  teams_ready: 'Équipes prêtes',
  in_progress: 'En cours',
  completed: 'Terminé',
  closed: 'Clôturé',
}

const statusClasses: Record<MatchStatus, string> = {
  draft: 'bg-bg-base text-fg-muted',
  open: 'bg-action-primary/15 text-action-primary',
  teams_ready: 'bg-action-primary/15 text-action-primary',
  in_progress: 'bg-team-green/15 text-team-green',
  completed: 'bg-bg-base text-fg-muted',
  closed: 'bg-bg-base text-fg-muted',
}
</script>

<template>
  <NuxtLink
    :to="`/matches/${id}`"
    class="flex flex-col gap-3 w-full text-left bg-bg-elevated rounded-[var(--radius-lg)] p-4 text-fg-default font-sans no-underline transition-colors duration-150 hover:bg-[#1F252D] active:bg-[#161B22] focus-visible:outline-none focus-visible:[box-shadow:0_0_0_2px_var(--color-bg-base),0_0_0_4px_rgba(32,128,255,0.45)]"
  >
    <div class="flex items-center justify-between gap-3">
      <div class="text-[18px] leading-[1.25] font-semibold tracking-[-0.01em] text-fg-default truncate">
        {{ title }}
      </div>
      <ChevronRight :size="18" class="text-fg-muted flex-shrink-0" />
    </div>

    <div class="flex items-center gap-4 text-sm text-fg-muted flex-wrap">
      <span class="inline-flex items-center gap-1.5">
        <Calendar :size="14" />
        <span>{{ formatScheduled(scheduledAt) }}</span>
      </span>
      <span class="inline-flex items-center gap-1.5">
        <MapPin :size="14" />
        <span class="truncate">{{ venue }}</span>
      </span>
    </div>

    <span
      class="inline-flex self-start items-center px-2 py-1 rounded-md text-xs font-medium"
      :class="statusClasses[status]"
    >
      {{ statusLabels[status] }}
    </span>
  </NuxtLink>
</template>
