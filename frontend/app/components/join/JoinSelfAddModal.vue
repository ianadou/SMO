<script setup lang="ts">
import BaseModal from '~/components/BaseModal.vue'
import TextField from '~/components/login/TextField.vue'
import InlineError from '~/components/login/InlineError.vue'
import InviteIcon from '~/components/invite/InviteIcon.vue'

const props = defineProps<{ open: boolean, busy: boolean, error?: string }>()

const emit = defineEmits<{ confirm: [firstName: string], cancel: [] }>()

const firstName = ref('')

watch(
  () => props.open,
  (isOpen) => {
    if (isOpen) firstName.value = ''
  },
)

const canSubmit = computed(() => firstName.value.trim().length > 0)

function submit() {
  if (!canSubmit.value || props.busy) return
  emit('confirm', firstName.value.trim())
}
</script>

<template>
  <BaseModal
    :open="open"
    variant="confirm"
    title="Je ne suis pas dans la liste"
    :close-disabled="busy"
    @close="emit('cancel')"
  >
    <form class="flex flex-col gap-4" @submit.prevent="submit">
      <TextField
        id="self-add-name"
        v-model="firstName"
        label="Votre prénom"
        placeholder="Prénom"
        autocomplete="given-name"
        :has-error="Boolean(error)"
      />
      <InlineError v-if="error" data-testid="self-add-error">{{ error }}</InlineError>
      <p class="text-[13px] leading-[1.5] text-fg-muted text-center m-0">
        Ce prénom sera définitivement le vôtre pour ce match — votre lien personnel sera créé.
      </p>
      <button
        type="submit"
        data-testid="self-add-confirm"
        class="inline-flex items-center justify-center gap-2 w-full h-[52px] rounded-[var(--radius-md)] border-0 text-base font-medium cursor-pointer bg-team-green text-[#0E1014] transition-[background,transform] duration-150 hover:brightness-110 active:translate-y-px disabled:opacity-60 disabled:cursor-not-allowed"
        :disabled="busy || !canSubmit"
      >
        <InviteIcon name="check" :size="18" />
        <span>Oui, c'est moi</span>
      </button>
      <button
        type="button"
        data-testid="self-add-cancel"
        class="self-center bg-transparent border-0 cursor-pointer text-fg-muted text-sm underline underline-offset-[3px] px-4 py-3 hover:text-fg-default disabled:opacity-60"
        :disabled="busy"
        @click="emit('cancel')"
      >
        Annuler
      </button>
    </form>
  </BaseModal>
</template>
