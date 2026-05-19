<script setup lang="ts">
import BaseModal from '~/components/BaseModal.vue'
import InviteIcon from './InviteIcon.vue'

defineProps<{ open: boolean, busy: boolean }>()

const emit = defineEmits<{ answer: ['yes' | 'no'], cancel: [] }>()
</script>

<template>
  <BaseModal
    :open="open"
    variant="confirm"
    title="Vous venez à ce match ?"
    :close-disabled="busy"
    @close="emit('cancel')"
  >
    <div class="flex flex-col gap-3">
      <button
        type="button"
        data-testid="answer-yes"
        class="inline-flex items-center justify-center gap-2 w-full h-[52px] rounded-[var(--radius-md)] border-0 text-base font-medium cursor-pointer bg-team-green text-[#0E1014] transition-[background,transform] duration-150 hover:brightness-110 active:translate-y-px disabled:opacity-60 disabled:cursor-not-allowed"
        :disabled="busy"
        @click="emit('answer', 'yes')"
      >
        <InviteIcon name="check" :size="18" />
        <span>Oui, je viens</span>
      </button>
      <button
        type="button"
        data-testid="answer-no"
        class="inline-flex items-center justify-center gap-2 w-full h-[52px] rounded-[var(--radius-md)] border-0 text-base font-medium cursor-pointer bg-team-red text-fg-emphasis transition-[background,transform] duration-150 hover:brightness-110 active:translate-y-px disabled:opacity-60 disabled:cursor-not-allowed"
        :disabled="busy"
        @click="emit('answer', 'no')"
      >
        <InviteIcon name="x" :size="18" />
        <span>Non, je ne peux pas</span>
      </button>
    </div>
    <button
      type="button"
      data-testid="answer-cancel"
      class="self-center bg-transparent border-0 cursor-pointer text-fg-muted text-sm underline underline-offset-[3px] px-4 py-3 hover:text-fg-default disabled:opacity-60"
      :disabled="busy"
      @click="emit('cancel')"
    >
      Annuler
    </button>
  </BaseModal>
</template>
