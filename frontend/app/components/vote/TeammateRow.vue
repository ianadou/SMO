<script setup lang="ts">
import AvatarRing from '~/components/ui/AvatarRing.vue'
import StarRow from '~/components/vote/StarRow.vue'
import type { VoteContextTeammate } from '~/types/vote'

defineProps<{
  teammate: VoteContextTeammate
  modelValue: number
}>()

const emit = defineEmits<{ 'update:modelValue': [value: number] }>()
</script>

<template>
  <div
    class="bg-bg-elevated rounded-[var(--radius-md)] px-4 py-3 flex items-center gap-3 max-[419px]:flex-wrap"
    :data-testid="`teammate-${teammate.player_id}`"
  >
    <AvatarRing :initials="teammate.initials" :size="44" />
    <div class="flex-1 min-w-0 flex flex-col gap-0.5">
      <div class="text-[15px] font-semibold text-fg-default truncate">
        {{ teammate.name }}
        <span
          v-if="teammate.your_score !== null"
          class="text-team-green text-[11px] font-normal ml-1"
        >✓ voté</span>
      </div>
      <div class="text-xs text-fg-muted">
        <span class="font-mono tabular-nums text-fg-default">{{ teammate.matches_together }}</span>
        matchs joués ensemble
      </div>
    </div>
    <StarRow
      class="max-[419px]:basis-full max-[419px]:pl-[56px]"
      :model-value="teammate.your_score ?? modelValue"
      :locked="teammate.your_score !== null"
      @update:model-value="emit('update:modelValue', $event)"
    />
  </div>
</template>
