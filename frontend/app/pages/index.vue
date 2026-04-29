<script setup lang="ts">
definePageMeta({ layout: false })

const router = useRouter()

onMounted(() => {
  router.replace('/login')
})
</script>

<template>
  <div />
</template>
