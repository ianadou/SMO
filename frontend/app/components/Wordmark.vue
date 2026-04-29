<script setup lang="ts">
const props = withDefaults(defineProps<{ size?: number }>(), { size: 40 })

const dotPx = computed(() => Math.max(4, Math.round(props.size * 0.13)))
const liftPx = computed(() => Math.round(props.size * 0.07))
const gapPx = computed(() => Math.max(3, Math.round(props.size * 0.1)))
</script>

<template>
  <span
    class="inline-flex items-baseline select-none"
    aria-label="SMO"
  >
    <span
      class="font-bold tracking-tight text-fg-emphasis leading-none font-sans"
      :style="{ fontSize: `${size}px` }"
    >SMO</span>
    <span
      class="inline-block rounded-full bg-action-primary"
      :style="{
        width: `${dotPx}px`,
        height: `${dotPx}px`,
        marginLeft: `${gapPx}px`,
        transform: `translateY(-${liftPx}px)`,
      }"
    />
  </span>
</template>
