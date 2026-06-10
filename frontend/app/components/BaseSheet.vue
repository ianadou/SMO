<script setup lang="ts">
const props = defineProps<{
  open: boolean
  title: string
}>()

const emit = defineEmits<{ close: [] }>()

const titleId = useId()
const dialog = ref<HTMLDialogElement | null>(null)

function onCancel(event: Event) {
  event.preventDefault()
  emit('close')
}

watch(
  () => props.open,
  (isOpen) => {
    const el = dialog.value
    if (!el) return
    if (isOpen) {
      if (!el.open) el.showModal()
      document.body.style.overflow = 'hidden'
    }
    else {
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
    class="bg-transparent p-0"
    @cancel="onCancel"
    @click.self="emit('close')"
  >
    <div
      v-if="open"
      class="w-full max-w-[600px] mx-auto bg-bg-elevated border border-border-default border-b-0 rounded-t-[var(--radius-lg)] px-4 pt-2 pb-[calc(theme(spacing.4)+env(safe-area-inset-bottom,0px))] max-h-[80dvh] overflow-y-auto"
    >
      <div class="w-9 h-1 rounded-full bg-border-default mx-auto mb-3" aria-hidden="true" />
      <h2
        :id="titleId"
        class="text-[17px] font-semibold tracking-[-0.01em] text-fg-default m-0 mb-3"
      >
        {{ title }}
      </h2>
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
  align-items: flex-end;
  justify-content: center;
  opacity: 1;
}

dialog[open] > div {
  animation: sheet-up 200ms var(--motion-easing, ease);
}

@keyframes sheet-up {
  from { transform: translateY(24px); opacity: 0.6; }
  to { transform: translateY(0); opacity: 1; }
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
