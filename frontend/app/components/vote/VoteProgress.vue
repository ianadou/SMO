<script setup lang="ts">
import { computed } from 'vue'

const props = defineProps<{
  done: number
  total: number
  label: string
}>()

const percent = computed(() =>
  props.total > 0 ? (props.done / props.total) * 100 : 0,
)
</script>

<template>
  <div class="flex flex-col gap-2">
    <div class="flex items-baseline justify-between text-[13px] text-fg-muted">
      <span>
        <span class="font-mono tabular-nums text-fg-default">{{ done }}</span>
        / <span class="font-mono tabular-nums">{{ total }}</span>
        {{ label }}
      </span>
    </div>
    <div
      class="h-1 rounded-full bg-white/[0.08] overflow-hidden"
      aria-hidden="true"
    >
      <div
        class="h-full rounded-full bg-action-primary transition-[width] duration-200"
        :style="{ width: `${percent}%` }"
      />
    </div>
  </div>
</template>
