<script setup lang="ts">
import { Plus, ArrowLeft, Webhook } from 'lucide-vue-next'
import CardSkeleton from '~/components/groups/CardSkeleton.vue'
import MatchCard from '~/components/groups/MatchCard.vue'
import CreateMatchModal from '~/components/groups/CreateMatchModal.vue'
import Wordmark from '~/components/Wordmark.vue'
import type { GroupDTO } from '~/types/groups'
import type { MatchDTO } from '~/types/matches'
import { ApiError } from '~/composables/useApi'

definePageMeta({ layout: false, middleware: 'auth' })

const route = useRoute()
const auth = useAuthStore()

const groupId = computed(() => route.params.id as string)

const group = ref<GroupDTO | null>(null)
const matches = ref<MatchDTO[]>([])
const loading = ref(true)
const error = ref('')
const notFound = ref(false)
const createOpen = ref(false)

useHead(() => ({
  title: group.value ? `${group.value.name} — SMO` : 'Groupe — SMO',
}))

async function loadAll() {
  error.value = ''
  notFound.value = false
  loading.value = true
  try {
    const api = useApi()
    const [g, m] = await Promise.all([
      api.get<GroupDTO>(`/groups/${groupId.value}`),
      api.get<MatchDTO[]>(`/groups/${groupId.value}/matches`),
    ])
    group.value = g
    matches.value = m
  } catch (e) {
    if (e instanceof ApiError && e.status === 401) {
      auth.logout()
      await navigateTo('/login', { replace: true })
      return
    }
    if (e instanceof ApiError && e.status === 404) {
      notFound.value = true
      return
    }
    error.value = e instanceof ApiError ? e.publicMessage : 'Impossible de charger le groupe.'
  } finally {
    loading.value = false
  }
}

async function onMatchCreated(created: MatchDTO) {
  createOpen.value = false
  matches.value = [created, ...matches.value]
}

onMounted(loadAll)
</script>

<template>
  <div class="relative w-full max-w-[600px] mx-auto min-h-dvh flex flex-col pb-24">
    <header class="flex items-center justify-between px-5 pt-5">
      <NuxtLink
        to="/groups"
        class="w-10 h-10 inline-flex items-center justify-center bg-transparent text-fg-default rounded-[var(--radius-pill)] no-underline transition-colors duration-150 hover:bg-white/5 active:bg-white/10 focus-visible:outline-none focus-visible:[box-shadow:0_0_0_2px_var(--color-bg-base),0_0_0_4px_rgba(32,128,255,0.45)]"
        aria-label="Retour aux groupes"
      >
        <ArrowLeft :size="18" />
      </NuxtLink>
      <Wordmark class="leading-none" />
      <div class="w-10 h-10" aria-hidden="true" />
    </header>

    <div v-if="notFound" class="flex flex-col items-center text-center px-6 py-14 gap-3">
      <h1 class="text-xl font-semibold tracking-[-0.01em] text-fg-default">
        Ce groupe n'existe pas
      </h1>
      <p class="text-fg-muted max-w-[320px]">
        Le groupe a peut-être été supprimé. Retourne à la liste pour en choisir un autre.
      </p>
      <NuxtLink
        to="/groups"
        class="mt-2 inline-flex items-center px-4 h-10 rounded-md bg-action-primary text-fg-emphasis no-underline font-medium transition-colors duration-150 hover:bg-action-primary-hover"
      >
        Mes groupes
      </NuxtLink>
    </div>

    <template v-else>
      <div class="px-5 pt-6 pb-5">
        <h1 class="text-3xl font-semibold tracking-[-0.01em] text-fg-default m-0">
          <template v-if="loading">
            <span class="inline-block h-8 w-2/3 rounded-md bg-bg-elevated animate-pulse align-middle" />
          </template>
          <template v-else-if="group">{{ group.name }}</template>
        </h1>
        <p
          v-if="group?.has_webhook"
          class="mt-2 inline-flex items-center gap-1.5 text-fg-muted text-sm"
        >
          <Webhook :size="14" />
          Notifications Discord activées
        </p>
      </div>

      <div class="px-5 pb-3">
        <h2 class="text-sm font-semibold uppercase tracking-wider text-fg-muted m-0">
          Matchs
        </h2>
      </div>

      <div class="flex flex-col gap-3 px-5">
        <template v-if="loading">
          <CardSkeleton v-for="n in 3" :key="n" />
        </template>

        <div
          v-else-if="error"
          role="alert"
          class="rounded-[var(--radius-md)] bg-bg-elevated text-fg-default px-4 py-3 text-sm"
        >
          {{ error }}
          <button
            type="button"
            class="ml-2 underline text-action-primary cursor-pointer bg-transparent border-0"
            @click="loadAll"
          >
            Réessayer
          </button>
        </div>

        <div
          v-else-if="matches.length === 0"
          class="flex flex-col items-center text-center px-6 py-10 gap-2"
        >
          <p class="text-fg-default font-medium">Pas encore de match</p>
          <p class="text-fg-muted text-sm max-w-[280px]">
            Crée le premier match pour ce groupe et invite tes joueurs.
          </p>
        </div>

        <MatchCard
          v-for="match in matches"
          v-else
          :id="match.id"
          :key="match.id"
          :title="match.title"
          :venue="match.venue"
          :scheduled-at="match.scheduled_at"
          :status="match.status"
        />
      </div>

      <button
        v-if="group"
        type="button"
        class="fixed right-[max(1.25rem,calc((100vw-600px)/2+1.25rem))] bottom-6 w-14 h-14 border-0 rounded-full bg-action-primary text-fg-emphasis inline-flex items-center justify-center cursor-pointer shadow-elevated transition-colors duration-150 hover:bg-action-primary-hover active:bg-action-primary-pressed focus-visible:outline-none focus-visible:[box-shadow:0_8px_24px_-8px_rgba(0,0,0,0.6),0_0_0_2px_var(--color-bg-base),0_0_0_4px_rgba(32,128,255,0.45)]"
        aria-label="Créer un match"
        @click="createOpen = true"
      >
        <Plus :size="22" />
      </button>

      <CreateMatchModal
        :open="createOpen"
        :group-id="groupId"
        @close="createOpen = false"
        @created="onMatchCreated"
      />
    </template>
  </div>
</template>
