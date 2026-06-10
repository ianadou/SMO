<script setup lang="ts">
import AvatarRing from '~/components/ui/AvatarRing.vue'
import type { VoteSelfResult } from '~/types/vote'

defineProps<{
  initials: string
  self: VoteSelfResult
}>()
</script>

<template>
  <div class="bg-bg-elevated rounded-[var(--radius-lg)] px-4 py-5 flex flex-col items-center gap-3 text-center">
    <AvatarRing :initials="initials" :size="88" :score="self.average" :stroke="4" />
    <div class="text-2xl font-mono font-semibold tabular-nums text-fg-default">
      <template v-if="self.average !== null">
        <span>{{ self.average.toFixed(1) }}</span>
        <span class="text-fg-muted font-normal text-base">/5</span>
      </template>
      <template v-else>
        <span class="text-fg-muted text-base font-normal">Aucun vote reçu</span>
      </template>
    </div>
    <div class="text-[13px] text-fg-muted">
      <span class="font-mono tabular-nums text-fg-default">{{ self.votes_count }}</span>
      vote{{ self.votes_count > 1 ? 's' : '' }} reçu{{ self.votes_count > 1 ? 's' : '' }}
    </div>
  </div>
</template>
