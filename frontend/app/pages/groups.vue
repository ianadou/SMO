<script setup lang="ts">
import { Plus, LogOut } from 'lucide-vue-next'
import GroupCard from '~/components/groups/GroupCard.vue'
import EmptyState from '~/components/groups/EmptyState.vue'
import CardSkeleton from '~/components/groups/CardSkeleton.vue'
import Wordmark from '~/components/Wordmark.vue'
import type { GroupDTO } from '~/types/groups'
import { ApiError } from '~/composables/useApi'

definePageMeta({ layout: false, middleware: 'auth' })

useHead({ title: 'Mes groupes — SMO' })

const auth = useAuthStore()

const groups = ref<GroupDTO[]>([])
const loading = ref(true)
const error = ref('')

async function loadGroups() {
  error.value = ''
  loading.value = true
  try {
    const api = useApi()
    groups.value = await api.get<GroupDTO[]>('/groups')
  } catch (e) {
    if (e instanceof ApiError && e.status === 401) {
      auth.logout()
      await navigateTo('/login', { replace: true })
      return
    }
    error.value = e instanceof ApiError ? e.publicMessage : 'Impossible de charger les groupes.'
  } finally {
    loading.value = false
  }
}

async function logout() {
  auth.logout()
  await navigateTo('/login', { replace: true })
}

onMounted(loadGroups)
</script>

<template>
  <div class="relative w-full max-w-[600px] mx-auto min-h-dvh flex flex-col pb-24">
    <header class="flex items-center justify-between px-5 pt-5">
      <Wordmark class="leading-none" />
      <button
        type="button"
        class="w-10 h-10 inline-flex items-center justify-center bg-transparent border-0 text-fg-default rounded-[var(--radius-pill)] cursor-pointer transition-colors duration-150 hover:bg-white/5 active:bg-white/10 focus-visible:outline-none focus-visible:[box-shadow:0_0_0_2px_var(--color-bg-base),0_0_0_4px_rgba(32,128,255,0.45)]"
        title="Se déconnecter"
        @click="logout"
      >
        <LogOut :size="18" />
      </button>
    </header>

    <div class="px-5 pt-6 pb-5">
      <h1 class="text-3xl font-semibold tracking-[-0.01em] text-fg-default m-0">
        Mes groupes
      </h1>
      <p v-if="auth.organizer" class="mt-2 text-fg-muted">
        Salut {{ auth.organizer.display_name }}, voici tes groupes.
      </p>
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
          @click="loadGroups"
        >
          Réessayer
        </button>
      </div>

      <EmptyState v-else-if="groups.length === 0" />

      <GroupCard
        v-for="group in groups"
        v-else
        :key="group.id"
        :name="group.name"
        :has-webhook="group.has_webhook"
        :created-at="group.created_at"
      />
    </div>

    <button
      type="button"
      disabled
      class="fixed right-[max(1.25rem,calc((100vw-600px)/2+1.25rem))] bottom-6 w-14 h-14 border-0 rounded-full bg-action-primary text-fg-emphasis inline-flex items-center justify-center cursor-not-allowed opacity-60 shadow-elevated"
      title="Création de groupe — bientôt"
      aria-label="Créer un groupe (bientôt disponible)"
    >
      <Plus :size="22" />
    </button>
  </div>
</template>
