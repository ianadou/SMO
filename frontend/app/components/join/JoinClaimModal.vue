<script setup lang="ts">
import BaseModal from '~/components/BaseModal.vue'
import InviteIcon from '~/components/invite/InviteIcon.vue'

defineProps<{ open: boolean, busy: boolean, playerName: string }>()

const emit = defineEmits<{ confirm: [], cancel: [] }>()
</script>

<template>
  <BaseModal
    :open="open"
    variant="confirm"
    :title="`Vous êtes ${playerName} ?`"
    :close-disabled="busy"
    @close="emit('cancel')"
  >
    <p class="text-[13px] leading-[1.5] text-fg-muted text-center m-0">
      Ce prénom sera définitivement le vôtre pour ce match — votre lien personnel sera créé.
    </p>
    <button
      type="button"
      data-testid="claim-confirm"
      class="inline-flex items-center justify-center gap-2 w-full h-[52px] rounded-[var(--radius-md)] border-0 text-base font-medium cursor-pointer bg-team-green text-[#0E1014] transition-[background,transform] duration-150 hover:brightness-110 active:translate-y-px disabled:opacity-60 disabled:cursor-not-allowed"
      :disabled="busy"
      @click="emit('confirm')"
    >
      <InviteIcon name="check" :size="18" />
      <span>Oui, c'est moi</span>
    </button>
    <button
      type="button"
      data-testid="claim-cancel"
      class="self-center bg-transparent border-0 cursor-pointer text-fg-muted text-sm underline underline-offset-[3px] px-4 py-3 hover:text-fg-default disabled:opacity-60"
      :disabled="busy"
      @click="emit('cancel')"
    >
      Annuler
    </button>
  </BaseModal>
</template>
