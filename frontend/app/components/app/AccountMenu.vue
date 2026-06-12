<script setup lang="ts">
import { LogOut } from 'lucide-vue-next'
import AvatarRing from '~/components/ui/AvatarRing.vue'
import { playerInitials } from '~/utils/playerInitials'

const props = defineProps<{ displayName: string, email: string }>()

const emit = defineEmits<{ logout: [] }>()

const open = ref(false)
const root = ref<HTMLElement | null>(null)

const initials = computed(() => playerInitials(props.displayName))

function onDocumentClick(event: MouseEvent) {
  if (!root.value?.contains(event.target as Node)) open.value = false
}

function onDocumentKeydown(event: KeyboardEvent) {
  if (event.key === 'Escape') open.value = false
}

onMounted(() => {
  document.addEventListener('click', onDocumentClick)
  document.addEventListener('keydown', onDocumentKeydown)
})

onBeforeUnmount(() => {
  document.removeEventListener('click', onDocumentClick)
  document.removeEventListener('keydown', onDocumentKeydown)
})
</script>

<template>
  <div ref="root" class="relative">
    <button
      type="button"
      class="inline-flex items-center justify-center bg-transparent border-0 rounded-full cursor-pointer transition-opacity duration-150 hover:opacity-80 focus-visible:outline-none focus-visible:[box-shadow:0_0_0_2px_var(--color-bg-base),0_0_0_4px_rgba(32,128,255,0.45)]"
      aria-haspopup="menu"
      :aria-expanded="open ? 'true' : 'false'"
      aria-label="Menu du compte"
      @click="open = !open"
    >
      <AvatarRing :initials="initials" :size="36" />
    </button>

    <div
      v-if="open"
      role="menu"
      class="absolute right-0 top-[calc(100%+8px)] z-50 min-w-[220px] rounded-[var(--radius-md)] bg-bg-elevated shadow-elevated py-2"
    >
      <div class="px-4 py-2">
        <div class="text-[14px] font-medium text-fg-default truncate">{{ displayName }}</div>
        <div class="text-[12px] text-fg-muted truncate">{{ email }}</div>
      </div>
      <div class="my-1 border-t border-white/5" />
      <button
        type="button"
        role="menuitem"
        data-testid="logout"
        class="w-full flex items-center gap-2 px-4 py-2 bg-transparent border-0 text-left text-[14px] text-fg-default cursor-pointer transition-colors duration-150 hover:bg-white/5 focus-visible:outline-none focus-visible:bg-white/5"
        @click="emit('logout')"
      >
        <LogOut :size="16" />
        Déconnexion
      </button>
    </div>
  </div>
</template>
