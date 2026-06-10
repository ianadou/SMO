<script setup lang="ts">
import { Star, TriangleAlert, Check } from 'lucide-vue-next'
import BaseModal from '~/components/BaseModal.vue'
import AvatarRing from '~/components/ui/AvatarRing.vue'
import type { VoteContextTeammate } from '~/types/vote'

defineProps<{
  open: boolean
  busy?: boolean
  teammates: VoteContextTeammate[]
  drafts: Record<string, number>
}>()

const emit = defineEmits<{ confirm: []; cancel: [] }>()
</script>

<template>
  <BaseModal
    :open="open"
    title="Confirmer vos votes ?"
    variant="confirm"
    :close-disabled="busy"
    @close="emit('cancel')"
  >
    <span class="self-center text-warn inline-flex -mt-2">
      <TriangleAlert :size="32" />
    </span>
    <p class="text-sm text-fg-muted text-center leading-[1.5] m-0">
      Cette action est définitive. Vous ne pourrez plus modifier vos notes.
    </p>

    <div class="flex flex-col gap-2 bg-bg-base rounded-[var(--radius-md)] p-3">
      <div
        v-for="teammate in teammates"
        :key="teammate.player_id"
        class="flex items-center gap-3"
      >
        <AvatarRing :initials="teammate.initials" :size="28" />
        <span class="flex-1 text-sm text-fg-default truncate">{{ teammate.name }}</span>
        <span class="inline-flex gap-px text-warn" :aria-label="`${drafts[teammate.player_id] ?? 0} étoiles`">
          <Star
            v-for="star in 5"
            :key="star"
            :size="14"
            :class="star <= (drafts[teammate.player_id] ?? 0) ? '' : 'text-fg-muted'"
            :fill="star <= (drafts[teammate.player_id] ?? 0) ? 'currentColor' : 'none'"
          />
        </span>
      </div>
    </div>

    <div class="flex flex-col items-center gap-2">
      <button
        type="button"
        data-testid="confirm-votes"
        class="inline-flex items-center justify-center gap-2 w-full h-[52px] rounded-[var(--radius-md)] border-0 text-base font-medium cursor-pointer bg-action-primary text-fg-emphasis transition-[background,transform] duration-150 hover:bg-action-primary-hover active:translate-y-px disabled:opacity-60 disabled:cursor-not-allowed"
        :disabled="busy"
        @click="emit('confirm')"
      >
        <Check :size="18" />
        <span>Confirmer</span>
      </button>
      <button
        type="button"
        data-testid="cancel-votes"
        class="bg-transparent border-0 cursor-pointer text-fg-muted text-sm px-4 py-3 underline underline-offset-[3px] hover:text-fg-default"
        :disabled="busy"
        @click="emit('cancel')"
      >
        Modifier
      </button>
    </div>
  </BaseModal>
</template>
