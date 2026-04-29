<script setup lang="ts">
import { AlertCircle } from 'lucide-vue-next'
</script>

<template>
  <div
    role="alert"
    class="flex items-start gap-2 mt-3 px-4 py-3 bg-team-red/10 text-[#FF8B97] rounded-md text-sm leading-[1.4]"
  >
    <AlertCircle :size="16" class="shrink-0 mt-px" />
    <span><slot /></span>
  </div>
</template>
