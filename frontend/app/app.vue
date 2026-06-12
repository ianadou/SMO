<script setup lang="ts">
import ToastStack from '~/components/ui/ToastStack.vue'
</script>

<template>
  <div>
    <NuxtRouteAnnouncer />
    <NuxtLayout>
      <NuxtPage />
    </NuxtLayout>
    <ToastStack />
  </div>
</template>
