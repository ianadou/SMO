<script setup lang="ts">
// Root component. Holds the wordmark and a small token preview so
// the scaffold visibly proves Tailwind v4, the design tokens, and
// the self-hosted Inter font are wired correctly.
//
// This stub is replaced by NuxtPage/NuxtLayout once the first real
// pages (login, groups, ...) are ported from the SMO Design System
// in follow-up PRs.
const tokens = [
  { name: 'bg-elevated', class: 'bg-bg-elevated' },
  { name: 'bg-subtle', class: 'bg-bg-subtle' },
  { name: 'border-default', class: 'bg-border-default' },
  { name: 'action-primary', class: 'bg-action-primary' },
  { name: 'team-red', class: 'bg-team-red' },
  { name: 'team-green', class: 'bg-team-green' },
  { name: 'warn', class: 'bg-warn' },
]
</script>

<template>
  <div class="min-h-screen flex flex-col items-center justify-center px-4 py-10 gap-10">
    <NuxtRouteAnnouncer />

    <!-- Wordmark per design system: off-white "SMO" + blue dot.
         Real wordmark component lands when login is ported. -->
    <h1 class="text-4xl font-bold tracking-tight text-fg-emphasis">
      SMO<span class="text-action-primary">·</span>
    </h1>

    <p class="text-fg-muted text-sm max-w-md text-center">
      Scaffold prêt — Inter + JetBrains Mono auto-hébergés via @nuxt/fonts,
      Tailwind v4 sur les tokens du design system, dark-only, mobile-first.
    </p>

    <div class="grid grid-cols-2 sm:grid-cols-4 gap-3 max-w-2xl w-full">
      <div
        v-for="token in tokens"
        :key="token.name"
        class="flex flex-col items-center gap-2"
      >
        <div :class="[token.class, 'w-full aspect-square rounded-md']" />
        <span class="text-xs text-fg-muted font-mono">{{ token.name }}</span>
      </div>
    </div>

    <p class="text-fg-subtle text-xs">
      Les pages métier arrivent dans les PR suivantes.
    </p>
  </div>
</template>
