<script setup lang="ts">
import TextField from '~/components/login/TextField.vue'
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

watch(
  () => props.open,
  (isOpen) => {
    if (!isOpen) reset()
  },
)
</script>

<template>
  <BaseModal :open="open" title="Nouveau match" :close-disabled="submitting" @close="close">
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

      <ModalActions
        :error="error"
        :submitting="submitting"
        :can-submit="canSubmit"
        submit-label="Créer le match"
        @cancel="close"
      />
    </form>
  </BaseModal>
</template>
