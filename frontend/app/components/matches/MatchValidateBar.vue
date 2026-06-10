<script setup lang="ts">
withDefaults(defineProps<{ busy?: boolean; label?: string }>(), {
  label: 'Valider les équipes',
})
const emit = defineEmits<{ validate: [] }>()
</script>

<template>
  <div class="md-bottombar">
    <div class="md-bottombar-inner">
      <button class="md-bottombar-btn" :disabled="busy" @click="emit('validate')">
        {{ label }}
      </button>
    </div>
  </div>
</template>
