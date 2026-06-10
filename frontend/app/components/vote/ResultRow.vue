<script setup lang="ts">
import { TrendingUp, TrendingDown } from 'lucide-vue-next'
import AvatarRing from '~/components/ui/AvatarRing.vue'
import type { VoteTeammateResult } from '~/types/vote'

defineProps<{ result: VoteTeammateResult }>()
</script>

<template>
  <div class="bg-bg-elevated rounded-[var(--radius-md)] px-4 py-3 flex items-center gap-3">
    <AvatarRing :initials="result.initials" :size="48" :score="result.average" />
    <div class="flex-1 min-w-0 flex flex-col gap-0.5">
      <div class="text-[15px] font-semibold text-fg-default truncate">{{ result.name }}</div>
      <div class="text-xs text-fg-muted">
        <span class="font-mono tabular-nums text-fg-default">{{ result.votes_count }}</span>
        vote{{ result.votes_count > 1 ? 's' : '' }} reçu{{ result.votes_count > 1 ? 's' : '' }}
      </div>
      <span
        v-if="result.delta !== null && result.delta !== 0"
        class="inline-flex items-center gap-1 text-[11px] font-mono tabular-nums"
        :class="result.delta > 0 ? 'text-team-green' : 'text-team-red'"
      >
        <TrendingUp v-if="result.delta > 0" :size="12" />
        <TrendingDown v-else :size="12" />
        <span>{{ result.delta > 0 ? '+' : '' }}{{ result.delta.toFixed(1) }} vs précédent</span>
      </span>
    </div>
    <div class="min-w-[56px] text-right text-sm font-mono tabular-nums text-fg-default">
      <span>{{ result.average.toFixed(1) }}</span>
      <span class="text-fg-muted">/5</span>
    </div>
  </div>
</template>
