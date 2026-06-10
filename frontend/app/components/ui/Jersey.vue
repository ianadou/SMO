<script setup lang="ts">
withDefaults(
  defineProps<{
    color?: 'red' | 'green'
    size?: number
    ariaLabel?: string
  }>(),
  { color: 'red', size: 52 },
)
</script>

<template>
  <svg
    :class="['sjersey', color === 'green' ? 'is-green' : 'is-red']"
    :width="size"
    :height="size"
    viewBox="0 0 64 64"
    xmlns="http://www.w3.org/2000/svg"
    :role="ariaLabel ? 'img' : 'presentation'"
    :aria-label="ariaLabel"
    :aria-hidden="ariaLabel ? undefined : 'true'"
  >
    <path
      d="
        M 22 6
        L 12 12
        L 6 22
        L 14 28
        L 18 24
        L 18 56
        C 18 57.66 19.34 59 21 59
        L 43 59
        C 44.66 59 46 57.66 46 56
        L 46 24
        L 50 28
        L 58 22
        L 52 12
        L 42 6
        L 36 12
        C 35 14.5 33.66 15.5 32 15.5
        C 30.34 15.5 29 14.5 28 12
        L 22 6
        Z
      "
      fill="currentColor"
    />
    <path
      d="
        M 46 24
        L 46 56
        C 46 57.66 44.66 59 43 59
        L 36 59
        L 36 22
        L 46 24
        Z
      "
      fill="rgba(255,255,255,0.08)"
    />
    <path
      d="M 28 12 C 29 14.5 30.34 15.5 32 15.5 C 33.66 15.5 35 14.5 36 12"
      fill="none"
      stroke="rgba(0,0,0,0.25)"
      stroke-width="1.2"
      stroke-linecap="round"
    />
  </svg>
</template>

<style scoped>
.sjersey { display: block; flex-shrink: 0; }
.sjersey.is-red   { color: var(--color-team-red); }
.sjersey.is-green { color: var(--color-team-green); }
</style>
