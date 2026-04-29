<script setup lang="ts">
import { Loader2 } from 'lucide-vue-next'

withDefaults(defineProps<{
  loading?: boolean
  disabled?: boolean
  type?: 'submit' | 'button' | 'reset'
  loadingLabel?: string
}>(), {
  loading: false,
  disabled: false,
  type: 'submit',
  loadingLabel: 'Connexion…',
})
</script>

<template>
  <button
    :type="type"
    :disabled="disabled || loading"
    class="inline-flex items-center justify-center gap-2 w-full h-12 border-0 rounded-md font-sans text-[15px] font-medium leading-none text-fg-emphasis cursor-pointer transition-colors duration-150 bg-action-primary hover:bg-action-primary-hover active:bg-action-primary-pressed focus-visible:outline-none focus-visible:[box-shadow:0_0_0_2px_var(--color-bg-base),0_0_0_4px_rgba(32,128,255,0.45)] disabled:opacity-60 disabled:cursor-not-allowed disabled:hover:bg-action-primary"
  >
    <template v-if="loading">
      <Loader2 :size="16" class="animate-spin" />
      <span>{{ loadingLabel }}</span>
    </template>
    <slot v-else />
  </button>
</template>
