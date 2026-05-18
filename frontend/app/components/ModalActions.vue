<script setup lang="ts">
import PrimaryButton from '~/components/login/PrimaryButton.vue'
import InlineError from '~/components/login/InlineError.vue'

withDefaults(
  defineProps<{
    submitting: boolean
    canSubmit: boolean
    submitLabel: string
    error?: string
    loadingLabel?: string
  }>(),
  { error: '', loadingLabel: 'Création…' },
)

defineEmits<{ cancel: [] }>()
</script>

<template>
  <InlineError v-if="error">{{ error }}</InlineError>

  <div class="flex gap-3 mt-2">
    <button
      type="button"
      class="flex-1 h-12 inline-flex items-center justify-center bg-transparent border border-border-default rounded-md text-fg-default font-medium cursor-pointer transition-colors duration-150 hover:bg-white/5 disabled:opacity-60 disabled:cursor-not-allowed"
      :disabled="submitting"
      @click="$emit('cancel')"
    >
      Annuler
    </button>
    <PrimaryButton
      type="submit"
      :loading="submitting"
      :disabled="!canSubmit"
      :loading-label="loadingLabel"
    >
      {{ submitLabel }}
    </PrimaryButton>
  </div>
</template>
