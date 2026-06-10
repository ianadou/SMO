<script setup lang="ts">
import { Star } from 'lucide-vue-next'

const props = withDefaults(
  defineProps<{
    modelValue: number
    locked?: boolean
    size?: number
  }>(),
  { locked: false, size: 20 },
)

const emit = defineEmits<{ 'update:modelValue': [value: number] }>()

function select(value: number) {
  if (props.locked) return
  emit('update:modelValue', value === props.modelValue ? 0 : value)
}
</script>

<template>
  <div
    class="inline-flex items-center gap-0.5"
    role="radiogroup"
    aria-label="Note de 1 à 5 étoiles"
  >
    <button
      v-for="star in 5"
      :key="star"
      type="button"
      :data-testid="`star-${star}`"
      class="w-8 h-8 inline-flex items-center justify-center bg-transparent border-0 p-0 rounded-[var(--radius-sm)] transition-[color,transform] duration-150"
      :class="[
        star <= modelValue ? 'text-warn' : 'text-fg-muted',
        locked ? 'cursor-default' : 'cursor-pointer hover:text-warn/60 active:scale-90',
      ]"
      role="radio"
      :aria-checked="star <= modelValue"
      :aria-label="`${star} étoile${star > 1 ? 's' : ''}`"
      :disabled="locked"
      @click="select(star)"
    >
      <Star
        :size="size"
        :fill="star <= modelValue ? 'currentColor' : 'none'"
      />
    </button>
  </div>
</template>
