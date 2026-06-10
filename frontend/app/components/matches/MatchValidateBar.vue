<script setup lang="ts">
defineProps<{ busy?: boolean }>()
const emit = defineEmits<{ validate: [] }>()
</script>

<template>
  <div class="md-bottombar">
    <div class="md-bottombar-inner">
      <button class="md-bottombar-btn" :disabled="busy" @click="emit('validate')">
        Valider les équipes
      </button>
    </div>
  </div>
</template>
