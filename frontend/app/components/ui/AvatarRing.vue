<script setup lang="ts">
import { computed } from 'vue'

const props = withDefaults(
  defineProps<{
    initials: string
    size?: number
    score?: number | null
    max?: number
    stroke?: number
  }>(),
  { size: 44, score: null, max: 5, stroke: 3 },
)

const ratio = computed(() => {
  if (props.score === null) return 0
  return Math.max(0, Math.min(1, props.score / props.max))
})
const radius = computed(() => (props.size - props.stroke) / 2)
const circumference = computed(() => 2 * Math.PI * radius.value)
const dash = computed(() => circumference.value * ratio.value)
const innerSize = computed(() => props.size - props.stroke * 2 - 4)
</script>

<template>
  <span
    class="relative inline-flex items-center justify-center shrink-0"
    :style="{ width: `${size}px`, height: `${size}px` }"
  >
    <svg
      :width="size"
      :height="size"
      class="absolute inset-0 -rotate-90"
      aria-hidden="true"
    >
      <circle
        :cx="size / 2"
        :cy="size / 2"
        :r="radius"
        fill="none"
        stroke="rgba(255,255,255,0.08)"
        :stroke-width="stroke"
      />
      <circle
        v-if="dash > 0"
        :cx="size / 2"
        :cy="size / 2"
        :r="radius"
        fill="none"
        stroke="var(--color-action-primary)"
        :stroke-width="stroke"
        stroke-linecap="round"
        :stroke-dasharray="`${dash} ${circumference - dash}`"
      />
    </svg>
    <span
      class="inline-flex items-center justify-center rounded-full bg-bg-subtle border border-border-default text-fg-default font-medium select-none"
      :style="{
        width: `${innerSize}px`,
        height: `${innerSize}px`,
        fontSize: `${Math.max(10, Math.round(innerSize * 0.34))}px`,
      }"
    >
      {{ initials }}
    </span>
  </span>
</template>
