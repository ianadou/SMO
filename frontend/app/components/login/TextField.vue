<script setup lang="ts">
defineProps<{
  id: string
  label: string
  type?: string
  placeholder?: string
  autocomplete?: string
  inputmode?: 'email' | 'numeric' | 'tel' | 'text' | 'url' | 'search'
  hasError?: boolean
}>()

defineEmits<{ blur: [event: FocusEvent] }>()

const model = defineModel<string>({ required: true })
</script>

<template>
  <div class="flex flex-col gap-2">
    <label
      :for="id"
      class="text-[13px] leading-[1.3] font-medium text-fg-default"
    >
      {{ label }}
    </label>

    <div
      class="relative flex items-center bg-bg-elevated rounded-md transition-shadow duration-150 focus-within:shadow-[inset_0_0_0_2px_var(--color-action-primary)]"
      :class="hasError && 'shadow-[inset_0_0_0_2px_var(--color-team-red)]'"
    >
      <input
        :id="id"
        v-model="model"
        :type="type ?? 'text'"
        :placeholder="placeholder"
        :autocomplete="autocomplete"
        :inputmode="inputmode"
        :aria-invalid="hasError ? 'true' : 'false'"
        class="flex-1 w-full h-12 bg-transparent border-0 outline-none text-fg-default px-4 text-base leading-[1.4] rounded-md placeholder:text-fg-muted"
        @blur="$emit('blur', $event)"
      >
      <slot name="right" />
    </div>
  </div>
</template>
