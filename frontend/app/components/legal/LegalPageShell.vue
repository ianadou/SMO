<script setup lang="ts">
defineProps<{ title: string, updatedAt: string }>()
</script>

<template>
  <main class="min-h-dvh flex flex-col">
    <div class="flex-1 w-full max-w-[560px] mx-auto px-6 py-10">
      <NuxtLink to="/" class="inline-block focus-visible:outline-none focus-visible:rounded-sm focus-visible:[box-shadow:0_0_0_2px_var(--color-bg-base),0_0_0_4px_rgba(32,128,255,0.45)]">
        <Wordmark :size="32" />
      </NuxtLink>

      <h1 class="mt-8 text-[26px] leading-[1.2] font-semibold tracking-[-0.01em] text-fg-default">
        {{ title }}
      </h1>
      <p class="mt-2 text-[13px] leading-[1.4] text-fg-muted">
        Dernière mise à jour : {{ updatedAt }}
      </p>

      <div class="mt-8 space-y-8 text-[15px] leading-[1.6] text-fg-default [&_h2]:text-[17px] [&_h2]:font-semibold [&_h2]:mb-2 [&_p]:text-fg-muted [&_li]:text-fg-muted [&_ul]:list-disc [&_ul]:pl-5 [&_ul]:space-y-1 [&_a]:text-fg-default [&_a]:underline [&_a]:underline-offset-2 [&_a]:decoration-fg-muted hover:[&_a]:decoration-fg-default">
        <slot />
      </div>

      <nav class="mt-12 pt-6 border-t border-bg-elevated text-[13px] leading-[1.4] text-fg-muted flex flex-wrap gap-x-4 gap-y-2">
        <NuxtLink to="/legal" class="underline underline-offset-2 decoration-fg-muted hover:decoration-fg-default">Mentions légales</NuxtLink>
        <NuxtLink to="/privacy" class="underline underline-offset-2 decoration-fg-muted hover:decoration-fg-default">Confidentialité</NuxtLink>
        <NuxtLink to="/terms" class="underline underline-offset-2 decoration-fg-muted hover:decoration-fg-default">CGU</NuxtLink>
      </nav>
    </div>
  </main>
</template>
