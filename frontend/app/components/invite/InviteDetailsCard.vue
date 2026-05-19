<script setup lang="ts">
import InviteIcon from './InviteIcon.vue'
import { formatMatchDate, formatMatchTime, parseCapacity } from '~/utils/inviteFormat'

const props = withDefaults(
  defineProps<{
    scheduledAt: string
    venue: string
    capacity: string
    compact?: boolean
  }>(),
  { compact: false },
)

const date = computed(() => formatMatchDate(props.scheduledAt))
const time = computed(() => formatMatchTime(props.scheduledAt))
const cap = computed(() => parseCapacity(props.capacity))
</script>

<template>
  <div
    data-testid="details-card"
    class="bg-bg-elevated rounded-[var(--radius-lg)] flex flex-col"
    :class="compact ? 'px-4 py-3 gap-2' : 'px-4 py-4 gap-3'"
  >
    <div class="flex items-center gap-3">
      <span class="shrink-0 w-5 h-5 inline-flex text-fg-muted"><InviteIcon name="calendar" :size="20" /></span>
      <span class="text-[15px] leading-[1.4] text-fg-default">{{ date }}</span>
    </div>
    <div class="flex items-center gap-3">
      <span class="shrink-0 w-5 h-5 inline-flex text-fg-muted"><InviteIcon name="clock" :size="20" /></span>
      <span class="text-[15px] leading-[1.4] text-fg-default font-mono tabular-nums">{{ time }}</span>
    </div>
    <div class="flex items-center gap-3">
      <span class="shrink-0 w-5 h-5 inline-flex text-fg-muted"><InviteIcon name="map-pin" :size="20" /></span>
      <span class="text-[15px] leading-[1.4] text-fg-default">{{ venue }}</span>
    </div>
    <div class="flex items-center gap-3">
      <span class="shrink-0 w-5 h-5 inline-flex text-fg-muted"><InviteIcon name="users" :size="20" /></span>
      <span class="text-[15px] leading-[1.4] text-fg-default">
        <span class="font-mono tabular-nums">{{ cap.count }}</span>
        <span> joueurs </span>
        <span v-if="cap.format" class="text-fg-muted">({{ cap.format }})</span>
      </span>
    </div>
  </div>
</template>
