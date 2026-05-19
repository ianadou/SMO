<script setup lang="ts">
const props = withDefaults(
  defineProps<{
    name:
      | 'calendar' | 'clock' | 'map-pin' | 'users'
      | 'check' | 'x' | 'check-circle'
      | 'check-circle-filled' | 'x-circle-filled' | 'unlink'
    size?: number
  }>(),
  { size: 20 },
)

const filled = computed(
  () => props.name === 'check-circle-filled' || props.name === 'x-circle-filled',
)
</script>

<template>
  <svg
    xmlns="http://www.w3.org/2000/svg"
    :width="size"
    :height="size"
    viewBox="0 0 24 24"
    :fill="filled ? 'currentColor' : 'none'"
    :stroke="filled ? undefined : 'currentColor'"
    :stroke-width="filled ? undefined : 1.8"
    stroke-linecap="round"
    stroke-linejoin="round"
    aria-hidden="true"
  >
    <template v-if="name === 'calendar'">
      <rect x="3" y="4" width="18" height="18" rx="2" />
      <line x1="16" y1="2" x2="16" y2="6" />
      <line x1="8" y1="2" x2="8" y2="6" />
      <line x1="3" y1="10" x2="21" y2="10" />
    </template>
    <template v-else-if="name === 'clock'">
      <circle cx="12" cy="12" r="10" />
      <polyline points="12 6 12 12 16 14" />
    </template>
    <template v-else-if="name === 'map-pin'">
      <path d="M20 10c0 6-8 12-8 12s-8-6-8-12a8 8 0 0 1 16 0Z" />
      <circle cx="12" cy="10" r="3" />
    </template>
    <template v-else-if="name === 'users'">
      <path d="M16 21v-2a4 4 0 0 0-4-4H6a4 4 0 0 0-4 4v2" />
      <circle cx="9" cy="7" r="4" />
      <path d="M22 21v-2a4 4 0 0 0-3-3.87" />
      <path d="M16 3.13a4 4 0 0 1 0 7.75" />
    </template>
    <template v-else-if="name === 'check'">
      <polyline points="20 6 9 17 4 12" />
    </template>
    <template v-else-if="name === 'x'">
      <line x1="18" y1="6" x2="6" y2="18" />
      <line x1="6" y1="6" x2="18" y2="18" />
    </template>
    <template v-else-if="name === 'check-circle'">
      <path d="M22 11.08V12a10 10 0 1 1-5.93-9.14" />
      <polyline points="22 4 12 14.01 9 11.01" />
    </template>
    <template v-else-if="name === 'check-circle-filled'">
      <path
        fill-rule="evenodd"
        d="M12 2a10 10 0 1 0 10 10A10 10 0 0 0 12 2zm-1.4 14.6L5.8 11.8l1.4-1.4 3.4 3.4 6.2-6.2 1.4 1.4-7.6 7.6z"
      />
    </template>
    <template v-else-if="name === 'x-circle-filled'">
      <path
        fill-rule="evenodd"
        d="M12 2a10 10 0 1 0 10 10A10 10 0 0 0 12 2zm4.3 13l-1.3 1.3-3-3-3 3-1.3-1.3 3-3-3-3 1.3-1.3 3 3 3-3 1.3 1.3-3 3z"
      />
    </template>
    <template v-else-if="name === 'unlink'">
      <path d="M9 17H7A5 5 0 0 1 7 7h2" />
      <path d="M15 7h2a5 5 0 0 1 4 8" />
      <line x1="8" y1="12" x2="12" y2="12" />
      <line x1="2" y1="2" x2="22" y2="22" />
    </template>
  </svg>
</template>
