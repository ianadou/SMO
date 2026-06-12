<script setup lang="ts">
import AppTopBar from '~/components/app/AppTopBar.vue'
import LegalFooter from '~/components/app/LegalFooter.vue'

const auth = useAuthStore()

async function logout() {
  auth.logout()
  await navigateTo('/login', { replace: true })
}
</script>

<template>
  <div class="min-h-dvh flex flex-col">
    <AppTopBar
      :display-name="auth.organizer?.display_name ?? ''"
      :email="auth.organizer?.email ?? ''"
      @logout="logout"
    />

    <div class="flex-1">
      <slot />
    </div>

    <footer class="py-6 px-5">
      <LegalFooter />
    </footer>
  </div>
</template>
