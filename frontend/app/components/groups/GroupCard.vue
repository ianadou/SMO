<script setup lang="ts">
import { ChevronRight, Webhook } from 'lucide-vue-next'

defineProps<{
  name: string
  hasWebhook: boolean
  createdAt: string
}>()

const formatter = new Intl.DateTimeFormat('fr-FR', {
  day: 'numeric',
  month: 'long',
  year: 'numeric',
})

function formatDate(iso: string): string {
  const date = new Date(iso)
  if (Number.isNaN(date.getTime())) return ''
  return formatter.format(date)
}
</script>

<template>
  <button
    type="button"
    class="flex flex-col gap-3 w-full text-left bg-bg-elevated border-0 rounded-[var(--radius-lg)] p-4 text-fg-default font-sans cursor-pointer transition-colors duration-150 hover:bg-[#1F252D] active:bg-[#161B22] focus-visible:outline-none focus-visible:[box-shadow:0_0_0_2px_var(--color-bg-base),0_0_0_4px_rgba(32,128,255,0.45)]"
  >
    <div class="flex items-center justify-between gap-3">
      <div class="text-[18px] leading-[1.25] font-semibold tracking-[-0.01em] text-fg-default truncate">
        {{ name }}
      </div>
      <ChevronRight :size="18" class="text-fg-muted flex-shrink-0" />
    </div>

    <div class="flex items-center gap-3 text-sm text-fg-muted">
      <span>Créé le {{ formatDate(createdAt) }}</span>
      <span
        v-if="hasWebhook"
        class="inline-flex items-center gap-1.5 text-fg-default"
        title="Notifications Discord activées"
      >
        <Webhook :size="14" />
        <span class="text-xs">Discord</span>
      </span>
    </div>
  </button>
</template>
