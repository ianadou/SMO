<script setup lang="ts">
import { ArrowRight } from 'lucide-vue-next'
import InviteDetailsCard from '~/components/invite/InviteDetailsCard.vue'
import InviteConfirmed from '~/components/invite/InviteConfirmed.vue'
import InviteMinimalState from '~/components/invite/InviteMinimalState.vue'
import JoinRosterRow from '~/components/join/JoinRosterRow.vue'
import JoinClaimModal from '~/components/join/JoinClaimModal.vue'
import JoinSelfAddModal from '~/components/join/JoinSelfAddModal.vue'
import LegalFooter from '~/components/app/LegalFooter.vue'
import { useJoinPage } from '~/composables/useJoinPage'
import { useToast } from '~/composables/useToast'
import { resolveJoinView } from '~/utils/shareLinkView'
import { readStashedInvitation, stashInvitation } from '~/utils/invitationStash'
import type { StashedInvitation } from '~/utils/invitationStash'
import type { ShareLinkContext, ShareRosterEntry } from '~/types/shareLink'

definePageMeta({ layout: false })
useHead({ title: 'Rejoindre le match — SMO' })

const route = useRoute()
const tokenParam = route.params.token
const token = (Array.isArray(tokenParam) ? tokenParam[0] : tokenParam) ?? ''

const { outcome, loading, submitting, load, claim, join } = useJoinPage(token)
const toast = useToast()

const context = computed<ShareLinkContext | null>(() =>
  outcome.value?.kind === 'ok' ? outcome.value.context : null,
)
const view = computed(() => (outcome.value ? resolveJoinView(outcome.value) : 'error'))

const stashed = computed<StashedInvitation | null>(() =>
  context.value ? readStashedInvitation(context.value.match_id) : null,
)

const selectedEntry = ref<ShareRosterEntry | null>(null)
const selfAddOpen = ref(false)
const selfAddError = ref('')

function openClaim(entry: ShareRosterEntry) {
  selectedEntry.value = entry
}

function openSelfAdd() {
  selfAddError.value = ''
  selfAddOpen.value = true
}

async function enterAs(matchId: string, invitationToken: string, playerName: string) {
  stashInvitation(matchId, { token: invitationToken, playerName })
  await navigateTo(`/invite/${invitationToken}?respond=1`)
}

async function confirmClaim() {
  const entry = selectedEntry.value
  const matchContext = context.value
  if (!entry || !matchContext) return
  const result = await claim(entry.player_id)
  if (result.kind === 'claimed') {
    await enterAs(matchContext.match_id, result.invitationToken, entry.player_name)
    return
  }
  selectedEntry.value = null
  if (result.kind === 'race') {
    toast.warning('Quelqu\'un vient de réclamer ce prénom')
    await load()
  }
  else if (result.kind === 'locked') {
    toast.warning('Les inscriptions sont closes')
    await load()
  }
  else {
    toast.error('Une erreur est survenue', 'Réessayez dans un instant.')
  }
}

async function confirmSelfAdd(firstName: string) {
  const matchContext = context.value
  if (!matchContext) return
  selfAddError.value = ''
  const result = await join(firstName)
  if (result.kind === 'joined') {
    await enterAs(matchContext.match_id, result.invitationToken, firstName)
    return
  }
  if (result.kind === 'name-taken') {
    selfAddError.value = 'Ce prénom est dans la liste — réclamez-le.'
  }
  else if (result.kind === 'locked') {
    selfAddOpen.value = false
    toast.warning('Les inscriptions sont closes')
    await load()
  }
  else {
    selfAddOpen.value = false
    toast.error('Une erreur est survenue', 'Réessayez dans un instant.')
  }
}

function resumeAsStashed() {
  if (stashed.value) navigateTo(`/invite/${stashed.value.token}`)
}

onMounted(load)
</script>

<template>
  <div
    v-if="loading"
    data-testid="join-skeleton"
    class="min-h-dvh max-w-[500px] mx-auto flex flex-col px-5 pt-6 pb-10"
  >
    <header class="flex justify-center pt-2 pb-8">
      <Wordmark :size="32" />
    </header>
    <div class="flex flex-col gap-6 animate-pulse" aria-hidden="true">
      <div class="flex flex-col gap-3">
        <div class="h-8 w-3/5 bg-bg-elevated rounded-[var(--radius-md)]" />
        <div class="h-4 w-4/5 bg-bg-elevated rounded-[var(--radius-md)]" />
      </div>
      <div class="h-[140px] bg-bg-elevated rounded-[var(--radius-lg)]" />
      <div class="flex flex-col gap-2">
        <div v-for="n in 4" :key="n" class="h-[52px] bg-bg-elevated rounded-[var(--radius-md)]" />
      </div>
    </div>
  </div>

  <InviteMinimalState
    v-else-if="view === 'invalid'"
    icon="unlink"
    icon-tone="muted"
    title="Ce lien n'est plus actif"
    subtitle="Demandez le nouveau lien à l'organisateur du match."
    legal="SMO ne stocke ni votre numéro ni votre email. Votre identité est liée au lien du match."
  />

  <InviteMinimalState
    v-else-if="view === 'locked'"
    icon="users"
    icon-tone="muted"
    title="Les inscriptions sont closes"
    subtitle="Les équipes sont en cours de formation. Contactez l'organisateur si vous souhaitez encore participer."
    legal="SMO ne stocke ni votre numéro ni votre email. Vous pouvez fermer cette page."
  />

  <InviteMinimalState
    v-else-if="view === 'error'"
    icon="unlink"
    icon-tone="muted"
    title="Une erreur est survenue"
    subtitle="Impossible de charger la page du match pour le moment."
    legal="SMO ne stocke ni votre numéro ni votre email."
    action-label="Réessayer"
    @action="load"
  />

  <main
    v-else-if="context"
    class="min-h-dvh max-w-[500px] mx-auto flex flex-col px-5 pt-6 pb-10"
  >
    <header class="flex justify-center pt-2 pb-8">
      <Wordmark :size="32" />
    </header>

    <section class="flex flex-col gap-2 mb-6">
      <h1 class="text-[28px] leading-[1.2] font-semibold tracking-[-0.01em] text-fg-default m-0">
        Vous êtes invités
      </h1>
      <p class="text-[15px] leading-[1.5] text-fg-muted m-0">
        <strong class="text-fg-default font-medium">{{ context.organizer_name }}</strong>
        invite le groupe
        <strong class="text-fg-default font-medium">{{ context.group_name }}</strong>
      </p>
    </section>

    <div class="mb-6">
      <InviteDetailsCard
        :scheduled-at="context.scheduled_at"
        :venue="context.venue"
        :capacity="context.capacity"
      />
    </div>

    <div class="mb-6">
      <InviteConfirmed
        :confirmed-count="context.confirmed_count"
        :max-participants="context.max_participants"
        :initials="context.confirmed_initials"
      />
    </div>

    <button
      v-if="stashed"
      type="button"
      data-testid="resume-banner"
      class="w-full mb-6 px-4 py-3 bg-action-primary/10 border border-action-primary/40 rounded-[var(--radius-md)] flex items-center justify-between gap-3 text-sm text-fg-default cursor-pointer transition-colors duration-150 hover:bg-action-primary/15"
      @click="resumeAsStashed"
    >
      <span>
        Vous êtes déjà
        <strong class="font-medium">{{ stashed.playerName }}</strong>
        — reprendre
      </span>
      <ArrowRight :size="16" class="text-action-primary shrink-0" />
    </button>

    <section class="flex flex-col gap-3">
      <h2 class="text-[12px] font-semibold tracking-[0.08em] uppercase text-fg-muted m-0">
        Qui êtes-vous ?
      </h2>
      <div class="flex flex-col gap-2">
        <JoinRosterRow
          v-for="entry in context.roster"
          :key="entry.player_id"
          :entry="entry"
          @claim="openClaim"
        />
        <button
          type="button"
          data-testid="self-add"
          class="w-full h-[52px] px-4 bg-transparent border border-dashed border-border-default rounded-[var(--radius-md)] flex items-center justify-center gap-2 text-[15px] text-fg-muted cursor-pointer transition-colors duration-150 hover:text-fg-default hover:bg-white/5"
          @click="openSelfAdd"
        >
          + Je ne suis pas dans la liste
        </button>
      </div>
    </section>

    <p class="mt-auto pt-8 text-[11px] leading-[1.5] text-fg-muted text-center">
      Votre lien personnel vous sera remis après identification.<br>
      SMO ne stocke ni votre numéro ni votre email.
    </p>

    <LegalFooter class="mt-3 pb-2" />

    <JoinClaimModal
      :open="selectedEntry !== null"
      :busy="submitting"
      :player-name="selectedEntry?.player_name ?? ''"
      @confirm="confirmClaim"
      @cancel="selectedEntry = null"
    />
    <JoinSelfAddModal
      :open="selfAddOpen"
      :busy="submitting"
      :error="selfAddError"
      @confirm="confirmSelfAdd"
      @cancel="selfAddOpen = false"
    />
  </main>
</template>
