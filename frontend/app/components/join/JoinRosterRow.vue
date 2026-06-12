<script setup lang="ts">
import { ChevronRight } from 'lucide-vue-next'
import type { ShareRosterEntry } from '~/types/shareLink'

const props = defineProps<{ entry: ShareRosterEntry }>()

const emit = defineEmits<{ claim: [entry: ShareRosterEntry] }>()

const takenLabel = computed(() =>
  props.entry.state === 'responded' ? '✓ a répondu' : 'déjà réclamé',
)
</script>

<template>
  <button
    v-if="entry.state === 'claimable'"
    type="button"
    :data-testid="`roster-${entry.player_id}`"
    class="w-full h-[52px] px-4 bg-bg-elevated border border-border-default rounded-[var(--radius-md)] flex items-center justify-between gap-3 text-[15px] font-medium text-fg-default cursor-pointer transition-colors duration-150 hover:bg-bg-subtle active:translate-y-px"
    @click="emit('claim', entry)"
  >
    <span class="truncate">{{ entry.player_name }}</span>
    <ChevronRight :size="18" class="text-fg-muted shrink-0" />
  </button>

  <div
    v-else
    :data-testid="`roster-${entry.player_id}`"
    class="w-full h-[52px] px-4 bg-bg-elevated/50 border border-border-default rounded-[var(--radius-md)] flex items-center justify-between gap-3 opacity-60"
  >
    <span class="truncate text-[15px] text-fg-muted">{{ entry.player_name }}</span>
    <span class="text-[12px] text-fg-muted shrink-0">{{ takenLabel }}</span>
  </div>
</template>
