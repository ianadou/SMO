<script setup lang="ts">
import AccountMenu from '~/components/app/AccountMenu.vue'
import Wordmark from '~/components/Wordmark.vue'

defineProps<{ displayName: string, email: string }>()

const emit = defineEmits<{ logout: [] }>()
</script>

<template>
  <header class="sticky top-0 z-40 bg-bg-base/95 backdrop-blur border-b border-white/5">
    <div class="w-full max-w-[600px] lg:max-w-[1040px] mx-auto px-5 h-14 flex items-center justify-between">
      <NuxtLink
        to="/groups"
        class="inline-flex focus-visible:outline-none focus-visible:rounded-sm focus-visible:[box-shadow:0_0_0_2px_var(--color-bg-base),0_0_0_4px_rgba(32,128,255,0.45)]"
        aria-label="Mes groupes"
      >
        <Wordmark :size="28" class="leading-none" />
      </NuxtLink>

      <AccountMenu
        :display-name="displayName"
        :email="email"
        @logout="emit('logout')"
      />
    </div>
  </header>
</template>
