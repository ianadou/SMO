<script setup lang="ts">
import { X } from 'lucide-vue-next'
import TextField from '~/components/login/TextField.vue'
import PrimaryButton from '~/components/login/PrimaryButton.vue'
import InlineError from '~/components/login/InlineError.vue'
import type { GroupDTO } from '~/types/groups'
import { ApiError } from '~/composables/useApi'

const props = defineProps<{ open: boolean }>()
const emit = defineEmits<{
  close: []
  created: [group: GroupDTO]
}>()

const NAME_MAX = 100
const REQUEST_TIMEOUT_MS = 10_000
const DISCORD_WEBHOOK_RE = /^https:\/\/discord\.com\/api\/webhooks\/\d+\/[\w-]+$/

const name = ref('')
const webhookURL = ref('')
const submitting = ref(false)
const error = ref('')
const webhookError = ref('')

const dialogRef = ref<HTMLDivElement | null>(null)
let previouslyFocused: HTMLElement | null = null

const trimmedName = computed(() => name.value.trim())
const nameTooLong = computed(() => name.value.length > NAME_MAX)
const canSubmit = computed(
  () => trimmedName.value.length > 0 && !nameTooLong.value && !submitting.value && webhookError.value === '',
)

function reset() {
  name.value = ''
  webhookURL.value = ''
  error.value = ''
  webhookError.value = ''
  submitting.value = false
}

function close() {
  if (submitting.value) return
  emit('close')
}

function trimName() {
  name.value = name.value.trim()
}

function validateWebhook() {
  const value = webhookURL.value.trim()
  if (value === '') {
    webhookError.value = ''
    return
  }
  if (!DISCORD_WEBHOOK_RE.test(value)) {
    webhookError.value = 'Format attendu : https://discord.com/api/webhooks/<id>/<token>'
    return
  }
  webhookError.value = ''
}

async function submit() {
  if (!canSubmit.value) return
  error.value = ''
  submitting.value = true

  const controller = new AbortController()
  const timeoutId = window.setTimeout(() => controller.abort(), REQUEST_TIMEOUT_MS)

  try {
    const api = useApi()
    const payload: Record<string, string> = { name: trimmedName.value }
    const trimmedWebhook = webhookURL.value.trim()
    if (trimmedWebhook !== '') {
      payload.discord_webhook_url = trimmedWebhook
    }
    const created = await api.post<GroupDTO>('/groups', payload, { signal: controller.signal })
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
          aria-labelledby="create-group-title"
          class="w-full max-w-[480px] bg-bg-base border border-border-default rounded-[var(--radius-lg)] p-6 shadow-elevated"
        >
          <header class="flex items-start justify-between gap-3 mb-5">
            <h2
              id="create-group-title"
              class="text-xl font-semibold tracking-[-0.01em] text-fg-default m-0"
            >
              Nouveau groupe
            </h2>
            <button
              type="button"
              class="w-9 h-9 inline-flex items-center justify-center bg-transparent border-0 text-fg-muted rounded-[var(--radius-pill)] cursor-pointer transition-colors duration-150 hover:bg-white/5 hover:text-fg-default focus-visible:outline-none focus-visible:[box-shadow:0_0_0_2px_var(--color-bg-base),0_0_0_4px_rgba(32,128,255,0.45)]"
              :aria-label="'Fermer'"
              :disabled="submitting"
              @click="close"
            >
              <X :size="18" />
            </button>
          </header>

          <form class="flex flex-col gap-4" @submit.prevent="submit">
            <div class="flex flex-col gap-2">
              <TextField
                id="group-name"
                v-model="name"
                label="Nom du groupe"
                placeholder="Foot du jeudi"
                autocomplete="off"
                :has-error="nameTooLong"
                @blur="trimName"
              />
              <div class="flex justify-between text-xs text-fg-muted">
                <span v-if="nameTooLong" class="text-team-red">
                  Trop long (max {{ NAME_MAX }} caractères)
                </span>
                <span v-else>Visible par tes joueurs sur l'invitation</span>
                <span :class="nameTooLong && 'text-team-red'">
                  {{ name.length }}/{{ NAME_MAX }}
                </span>
              </div>
            </div>

            <div class="flex flex-col gap-2">
              <TextField
                id="group-webhook"
                v-model="webhookURL"
                label="Webhook Discord (optionnel)"
                type="url"
                inputmode="url"
                placeholder="https://discord.com/api/webhooks/..."
                autocomplete="off"
                :has-error="webhookError !== ''"
                @blur="validateWebhook"
              />
              <p class="text-xs text-fg-muted leading-[1.4]">
                Colle l'URL du webhook Discord du salon où SMO postera quand les équipes
                sont prêtes. Crée-la dans Discord → Paramètres du salon → Intégrations.
              </p>
              <p v-if="webhookError" class="text-xs text-team-red">{{ webhookError }}</p>
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
                Créer le groupe
              </PrimaryButton>
            </div>
          </form>
        </div>
      </div>
    </Transition>
  </Teleport>
</template>
