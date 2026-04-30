<script setup lang="ts">
import { X } from 'lucide-vue-next'
import TextField from '~/components/login/TextField.vue'
import PrimaryButton from '~/components/login/PrimaryButton.vue'
import InlineError from '~/components/login/InlineError.vue'
import type { MatchDTO, CreateMatchPayload } from '~/types/matches'
import { ApiError } from '~/composables/useApi'

const props = defineProps<{
  open: boolean
  groupId: string
}>()
const emit = defineEmits<{
  close: []
  created: [match: MatchDTO]
}>()

const TITLE_MAX = 100
const VENUE_MAX = 200
const REQUEST_TIMEOUT_MS = 10_000

const title = ref('')
const venue = ref('')
const scheduledAtLocal = ref('')
const submitting = ref(false)
const error = ref('')
const dateError = ref('')

const dialogRef = ref<HTMLDivElement | null>(null)
let previouslyFocused: HTMLElement | null = null

const trimmedTitle = computed(() => title.value.trim())
const trimmedVenue = computed(() => venue.value.trim())
const titleTooLong = computed(() => title.value.length > TITLE_MAX)
const venueTooLong = computed(() => venue.value.length > VENUE_MAX)
const canSubmit = computed(
  () =>
    trimmedTitle.value.length > 0
    && trimmedVenue.value.length > 0
    && scheduledAtLocal.value !== ''
    && !titleTooLong.value
    && !venueTooLong.value
    && dateError.value === ''
    && !submitting.value,
)

function reset() {
  title.value = ''
  venue.value = ''
  scheduledAtLocal.value = ''
  error.value = ''
  dateError.value = ''
  submitting.value = false
}

function close() {
  if (submitting.value) return
  emit('close')
}

function trimTitle() {
  title.value = title.value.trim()
}

function trimVenue() {
  venue.value = venue.value.trim()
}

function validateDate() {
  if (scheduledAtLocal.value === '') {
    dateError.value = ''
    return
  }
  const parsed = new Date(scheduledAtLocal.value)
  if (Number.isNaN(parsed.getTime())) {
    dateError.value = 'Date invalide.'
    return
  }
  if (parsed.getTime() <= Date.now()) {
    dateError.value = 'La date doit être dans le futur.'
    return
  }
  dateError.value = ''
}

async function submit() {
  if (!canSubmit.value) return
  error.value = ''
  submitting.value = true

  const controller = new AbortController()
  const timeoutId = window.setTimeout(() => controller.abort(), REQUEST_TIMEOUT_MS)

  try {
    const api = useApi()
    const payload: CreateMatchPayload = {
      group_id: props.groupId,
      title: trimmedTitle.value,
      venue: trimmedVenue.value,
      scheduled_at: new Date(scheduledAtLocal.value).toISOString(),
    }
    const created = await api.post<MatchDTO>('/matches', payload, { signal: controller.signal })
    emit('created', created)
    reset()
  } catch (e) {
    if (e instanceof DOMException && e.name === 'AbortError') {
      error.value = 'La requête a pris trop de temps. Réessaie.'
    } else if (e instanceof ApiError) {
      error.value = e.publicMessage
    } else if (typeof navigator !== 'undefined' && navigator.onLine === false) {
      error.value = 'Tu es hors connexion. Vérifie ton réseau.'
    } else {
      error.value = 'Connexion au serveur impossible. Réessaie dans un instant.'
    }
  } finally {
    window.clearTimeout(timeoutId)
    submitting.value = false
  }
}

function getFocusable(): HTMLElement[] {
  if (!dialogRef.value) return []
  return Array.from(
    dialogRef.value.querySelectorAll<HTMLElement>(
      'a[href], button:not([disabled]), input:not([disabled]), [tabindex]:not([tabindex="-1"])',
    ),
  )
}

function onKeydown(event: KeyboardEvent) {
  if (!props.open) return
  if (event.key === 'Escape') {
    event.preventDefault()
    close()
    return
  }
  if (event.key !== 'Tab') return
  const focusable = getFocusable()
  if (focusable.length === 0) return
  const first = focusable[0]
  const last = focusable[focusable.length - 1]
  const active = document.activeElement as HTMLElement | null
  if (event.shiftKey && active === first) {
    event.preventDefault()
    last.focus()
  } else if (!event.shiftKey && active === last) {
    event.preventDefault()
    first.focus()
  }
}

watch(
  () => props.open,
  async (isOpen) => {
    if (typeof window === 'undefined') return
    if (isOpen) {
      previouslyFocused = document.activeElement as HTMLElement | null
      document.body.style.overflow = 'hidden'
      await nextTick()
      const firstInput = dialogRef.value?.querySelector<HTMLInputElement>('input')
      firstInput?.focus()
    } else {
      document.body.style.overflow = ''
      previouslyFocused?.focus()
      reset()
    }
  },
  { immediate: false },
)

onMounted(() => {
  if (typeof document !== 'undefined') {
    document.addEventListener('keydown', onKeydown)
  }
})

onUnmounted(() => {
  if (typeof document !== 'undefined') {
    document.removeEventListener('keydown', onKeydown)
    document.body.style.overflow = ''
  }
})
</script>

<template>
  <Teleport to="body">
    <Transition
      enter-active-class="transition-opacity duration-150"
      enter-from-class="opacity-0"
      leave-active-class="transition-opacity duration-150"
      leave-to-class="opacity-0"
    >
      <div
        v-if="open"
        class="fixed inset-0 z-50 flex items-center justify-center bg-black/60 px-4"
        @click.self="close"
      >
        <div
          ref="dialogRef"
          role="dialog"
          aria-modal="true"
          aria-labelledby="create-match-title"
          class="w-full max-w-[480px] bg-bg-base border border-border-default rounded-[var(--radius-lg)] p-6 shadow-elevated"
        >
          <header class="flex items-start justify-between gap-3 mb-5">
            <h2
              id="create-match-title"
              class="text-xl font-semibold tracking-[-0.01em] text-fg-default m-0"
            >
              Nouveau match
            </h2>
            <button
              type="button"
              class="w-9 h-9 inline-flex items-center justify-center bg-transparent border-0 text-fg-muted rounded-[var(--radius-pill)] cursor-pointer transition-colors duration-150 hover:bg-white/5 hover:text-fg-default focus-visible:outline-none focus-visible:[box-shadow:0_0_0_2px_var(--color-bg-base),0_0_0_4px_rgba(32,128,255,0.45)]"
              aria-label="Fermer"
              :disabled="submitting"
              @click="close"
            >
              <X :size="18" />
            </button>
          </header>

          <form class="flex flex-col gap-4" @submit.prevent="submit">
            <div class="flex flex-col gap-2">
              <TextField
                id="match-title"
                v-model="title"
                label="Titre"
                placeholder="Match du jeudi soir"
                autocomplete="off"
                :has-error="titleTooLong"
                @blur="trimTitle"
              />
              <div class="flex justify-between text-xs text-fg-muted">
                <span v-if="titleTooLong" class="text-team-red">
                  Trop long (max {{ TITLE_MAX }} caractères)
                </span>
                <span v-else>Le nom que verront les joueurs invités</span>
                <span :class="titleTooLong && 'text-team-red'">
                  {{ title.length }}/{{ TITLE_MAX }}
                </span>
              </div>
            </div>

            <div class="flex flex-col gap-2">
              <TextField
                id="match-venue"
                v-model="venue"
                label="Lieu"
                placeholder="Stade municipal, terrain 3"
                autocomplete="off"
                :has-error="venueTooLong"
                @blur="trimVenue"
              />
              <div v-if="venueTooLong" class="text-xs text-team-red">
                Trop long (max {{ VENUE_MAX }} caractères)
              </div>
            </div>

            <div class="flex flex-col gap-2">
              <label
                for="match-scheduled"
                class="text-sm font-medium text-fg-default"
              >
                Date et heure
              </label>
              <input
                id="match-scheduled"
                v-model="scheduledAtLocal"
                type="datetime-local"
                class="h-12 px-3 bg-bg-elevated border border-border-default rounded-md text-fg-default font-sans text-base focus:outline-none focus:border-action-primary"
                :class="dateError && 'border-team-red'"
                @change="validateDate"
                @blur="validateDate"
              >
              <p v-if="dateError" class="text-xs text-team-red">{{ dateError }}</p>
            </div>

            <InlineError v-if="error">{{ error }}</InlineError>

            <div class="flex gap-3 mt-2">
              <button
                type="button"
                class="flex-1 h-12 inline-flex items-center justify-center bg-transparent border border-border-default rounded-md text-fg-default font-medium cursor-pointer transition-colors duration-150 hover:bg-white/5 disabled:opacity-60 disabled:cursor-not-allowed"
                :disabled="submitting"
                @click="close"
              >
                Annuler
              </button>
              <PrimaryButton
                type="submit"
                :loading="submitting"
                :disabled="!canSubmit"
                loading-label="Création…"
              >
                Créer le match
              </PrimaryButton>
            </div>
          </form>
        </div>
      </div>
    </Transition>
  </Teleport>
</template>
