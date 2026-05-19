<script setup lang="ts">
const props = defineProps<{
  confirmedCount: number
  maxParticipants: number
  initials: string[]
}>()

const visible = computed(() => props.initials.slice(0, 6))
</script>

<template>
  <div
    v-if="confirmedCount > 0"
    data-testid="confirmed"
    class="flex flex-col gap-3"
  >
    <div class="flex items-baseline justify-between text-[13px] text-fg-muted">
      <span>Déjà confirmés</span>
      <span>
        <span class="text-fg-default font-mono tabular-nums">{{ confirmedCount }}</span>
        <span> / </span>
        <span class="text-fg-default font-mono tabular-nums">{{ maxParticipants }}</span>
      </span>
    </div>
    <div class="flex items-center">
      <span
        v-for="(value, index) in visible"
        :key="index"
        data-testid="avatar"
        class="w-9 h-9 rounded-full bg-bg-subtle text-fg-default text-xs font-semibold inline-flex items-center justify-center border-2 border-bg-base shrink-0"
        :class="index === 0 ? 'ml-0' : '-ml-2'"
      >{{ value }}</span>
    </div>
  </div>
</template>
