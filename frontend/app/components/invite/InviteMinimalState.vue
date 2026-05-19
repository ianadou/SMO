<script setup lang="ts">
import InviteIcon from './InviteIcon.vue'

defineProps<{
  icon: 'clock' | 'unlink' | 'users'
  iconTone: 'warn' | 'muted'
  title: string
  subtitle: string
  legal: string
  actionLabel?: string
}>()

const emit = defineEmits<{ action: [] }>()
</script>

<template>
  <main class="min-h-dvh max-w-[500px] mx-auto flex flex-col px-5 pt-6 pb-10">
    <header class="flex justify-center pt-2 pb-8">
      <Wordmark :size="32" />
    </header>

    <section class="my-auto flex flex-col items-center text-center gap-3">
      <span
        data-testid="invite-icon"
        class="inline-flex"
        :class="iconTone === 'warn' ? 'text-warn' : 'text-fg-muted'"
      >
        <InviteIcon :name="icon" :size="60" />
      </span>
      <h1 class="text-[28px] leading-[1.2] font-semibold tracking-[-0.01em] text-fg-default m-0">
        {{ title }}
      </h1>
      <p class="text-[15px] leading-[1.5] text-fg-muted m-0 max-w-[34ch]">
        {{ subtitle }}
      </p>
      <button
        v-if="actionLabel"
        type="button"
        class="mt-2 inline-flex items-center justify-center w-full h-[52px] rounded-[var(--radius-md)] bg-bg-elevated border border-border-default text-fg-default text-base font-medium cursor-pointer transition-colors duration-150 hover:bg-bg-subtle"
        @click="emit('action')"
      >
        {{ actionLabel }}
      </button>
    </section>

    <p class="mt-auto pt-8 text-[11px] leading-[1.5] text-fg-muted text-center">
      {{ legal }}
    </p>
  </main>
</template>
