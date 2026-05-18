<script setup lang="ts">
import { X } from 'lucide-vue-next'

const props = defineProps<{
  open: boolean
  title: string
  closeDisabled?: boolean
}>()

const emit = defineEmits<{ close: [] }>()

const titleId = useId()
const dialog = ref<HTMLDialogElement | null>(null)

function requestClose() {
  if (props.closeDisabled) return
  emit('close')
}

function onCancel(event: Event) {
  event.preventDefault()
  requestClose()
}

function onBackdropClick() {
  requestClose()
}

watch(
  () => props.open,
  async (isOpen) => {
    const el = dialog.value
    if (!el) return
    if (isOpen) {
      if (!el.open) el.showModal()
      document.body.style.overflow = 'hidden'
      await nextTick()
      el.querySelector<HTMLInputElement>('input, textarea, select')?.focus()
    } else {
      if (el.open) el.close()
      document.body.style.overflow = ''
    }
  },
)

onUnmounted(() => {
  document.body.style.overflow = ''
})
</script>

<template>
  <dialog
    ref="dialog"
    :aria-labelledby="titleId"
    class="bg-transparent p-4"
    @cancel="onCancel"
    @click.self="onBackdropClick"
  >
    <div
      v-if="open"
      class="w-full max-w-[480px] bg-bg-base border border-border-default rounded-[var(--radius-lg)] p-6 shadow-elevated"
    >
      <header class="flex items-start justify-between gap-3 mb-5">
        <h2
          :id="titleId"
          class="text-xl font-semibold tracking-[-0.01em] text-fg-default m-0"
        >
          {{ title }}
        </h2>
        <button
          type="button"
          class="w-9 h-9 inline-flex items-center justify-center bg-transparent border-0 text-fg-muted rounded-[var(--radius-pill)] cursor-pointer transition-colors duration-150 hover:bg-white/5 hover:text-fg-default focus-visible:outline-none focus-visible:[box-shadow:0_0_0_2px_var(--color-bg-base),0_0_0_4px_rgba(32,128,255,0.45)] disabled:opacity-60 disabled:cursor-not-allowed"
          aria-label="Fermer"
          :disabled="closeDisabled"
          @click="requestClose"
        >
          <X :size="18" />
        </button>
      </header>

      <slot />
    </div>
  </dialog>
</template>

<style scoped>
dialog {
  margin: 0;
  width: 100%;
  height: 100%;
  max-width: none;
  max-height: none;
  opacity: 0;
  transition: opacity 150ms ease;
}

dialog[open] {
  display: flex;
  align-items: center;
  justify-content: center;
  opacity: 1;
}

dialog::backdrop {
  background: rgb(0 0 0 / 0.6);
  opacity: 0;
  transition: opacity 150ms ease;
}

dialog[open]::backdrop {
  opacity: 1;
}
</style>
