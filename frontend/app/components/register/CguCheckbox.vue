<script setup lang="ts">
import { Check } from 'lucide-vue-next'

defineProps<{ hasError?: boolean }>()

const checked = defineModel<boolean>({ required: true })
</script>

<template>
  <label class="flex items-start gap-3 cursor-pointer select-none">
    <input
      v-model="checked"
      type="checkbox"
      class="sr-only peer"
      :aria-invalid="hasError ? 'true' : 'false'"
    >
    <span
      class="relative flex items-center justify-center w-5 h-5 mt-px shrink-0 rounded-sm bg-bg-elevated transition-colors duration-150 peer-checked:bg-action-primary peer-focus-visible:[box-shadow:0_0_0_2px_var(--color-bg-base),0_0_0_4px_rgba(32,128,255,0.45)]"
      :class="hasError && !checked && 'shadow-[inset_0_0_0_1px_var(--color-team-red)]'"
    >
      <Check v-if="checked" :size="14" class="text-fg-emphasis" />
    </span>
    <span class="text-[14px] leading-[1.4] text-fg-default">
      J'accepte les
      <a href="#cgu" target="_blank" rel="noopener noreferrer" class="text-fg-default underline underline-offset-2 decoration-fg-muted hover:decoration-fg-default">conditions générales</a>
      et la
      <a href="#privacy" target="_blank" rel="noopener noreferrer" class="text-fg-default underline underline-offset-2 decoration-fg-muted hover:decoration-fg-default">politique de confidentialité</a>
    </span>
  </label>
</template>
