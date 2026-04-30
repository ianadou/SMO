<script setup lang="ts">
import { Users } from 'lucide-vue-next'
</script>

<template>
  <div class="flex flex-col items-center text-center px-6 py-14 gap-3">
    <div class="w-24 h-24 rounded-full bg-bg-elevated text-fg-muted inline-flex items-center justify-center mb-4">
      <Users :size="36" />
    </div>
    <h2 class="text-xl font-semibold tracking-[-0.01em] text-fg-default">
      Pas encore de groupe
    </h2>
    <p class="text-fg-muted max-w-[320px]">
      Crée ton premier groupe pour inviter tes potes et organiser ton premier match.
    </p>
  </div>
</template>
