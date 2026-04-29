<script setup lang="ts">
import { passwordStrengthLabel, type StrengthLevel } from '~/utils/password'

const props = defineProps<{ level: StrengthLevel }>()

const SEGMENTS = [1, 2, 3, 4] as const

const segmentColor = computed(() => {
  switch (props.level) {
    case 1: return 'bg-team-red'
    case 2: return 'bg-warn'
    case 3: return 'bg-team-green'
    case 4: return 'bg-team-green'
    default: return 'bg-bg-elevated'
  }
})

const labelColor = computed(() => {
  switch (props.level) {
    case 1: return 'text-team-red'
    case 2: return 'text-warn'
    case 3:
    case 4: return 'text-team-green'
    default: return 'text-fg-muted'
  }
})
</script>

<template>
  <div class="mt-2">
    <div
      class="grid grid-cols-4 gap-1 h-1"
      aria-hidden="true"
    >
      <span
        v-for="i in SEGMENTS"
        :key="i"
        :class="[
          'h-1 rounded-full transition-colors duration-150',
          i <= level ? segmentColor : 'bg-bg-elevated',
        ]"
      />
    </div>
    <div class="flex items-center justify-between mt-2 text-[12px] leading-[1.3]">
      <span class="text-fg-muted">Au moins 8 caractères</span>
      <span :class="[labelColor, 'font-mono']">{{ passwordStrengthLabel(level) }}</span>
    </div>
  </div>
</template>
