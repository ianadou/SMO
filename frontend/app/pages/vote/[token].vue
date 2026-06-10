<script setup lang="ts">
import { ArrowRight, CircleCheck } from 'lucide-vue-next'
import TeamMatchupCard from '~/components/matches/TeamMatchupCard.vue'
import TeammateRow from '~/components/vote/TeammateRow.vue'
import VoteProgress from '~/components/vote/VoteProgress.vue'
import VoteConfirmModal from '~/components/vote/VoteConfirmModal.vue'
import ResultRow from '~/components/vote/ResultRow.vue'
import SelfScoreCard from '~/components/vote/SelfScoreCard.vue'
import InviteMinimalState from '~/components/invite/InviteMinimalState.vue'
import { resolveVoteView } from '~/utils/voteView'
import { formatMatchDate, formatMatchTime } from '~/utils/inviteFormat'
import { useVotePage } from '~/composables/useVotePage'
import { useToast } from '~/composables/useToast'
import type { VoteContext, VoteContextTeammate } from '~/types/vote'

definePageMeta({ layout: false })
useHead({ title: 'Vote — SMO' })

const route = useRoute()
const tokenParam = route.params.token
const token = (Array.isArray(tokenParam) ? tokenParam[0] : tokenParam) ?? ''

const { outcome, loading, submitting, load, submitVotes } = useVotePage(token)
const toast = useToast()

const drafts = ref<Record<string, number>>({})
const modalOpen = ref(false)

const context = computed<VoteContext | null>(() =>
  outcome.value?.kind === 'ok' ? outcome.value.context : null,
)
const view = computed(() => (outcome.value ? resolveVoteView(outcome.value) : 'error'))

const remaining = computed<VoteContextTeammate[]>(() =>
  context.value?.teammates.filter(teammate => teammate.your_score === null) ?? [],
)
const filledCount = computed(() => {
  if (!context.value) return 0
  const locked = context.value.teammates.length - remaining.value.length
  const drafted = remaining.value.filter(teammate => (drafts.value[teammate.player_id] ?? 0) > 0).length
  return locked + drafted
})
const allRated = computed(
  () => remaining.value.length > 0
    && remaining.value.every(teammate => (drafts.value[teammate.player_id] ?? 0) > 0),
)

const teamLabel = computed(() => (context.value?.voter.team === 'B' ? 'verte' : 'rouge'))
const teamDotClass = computed(() =>
  context.value?.voter.team === 'B' ? 'bg-team-green' : 'bg-team-red',
)
const cardStatus = computed(() => (context.value?.status === 'closed' ? 'closed' : 'finished'))
const cardWinner = computed(() => {
  if (context.value?.winner === 'A') return 'red'
  if (context.value?.winner === 'B') return 'green'
  return undefined
})
const dateLabel = computed(() => {
  if (!context.value) return ''
  const raw = new Date(context.value.scheduled_at).toLocaleDateString('fr-FR', {
    weekday: 'short',
    day: 'numeric',
    month: 'short',
  })
  return raw.charAt(0).toUpperCase() + raw.slice(1)
})

function setDraft(playerId: string, score: number) {
  drafts.value = { ...drafts.value, [playerId]: score }
}

async function confirmVotes() {
  const votes = remaining.value
    .map(teammate => ({
      voted_id: teammate.player_id,
      score: drafts.value[teammate.player_id] ?? 0,
    }))
    .filter(vote => vote.score > 0)
  const failedIds = await submitVotes(votes)
  modalOpen.value = false
  if (failedIds.length > 0) {
    toast.error('Votes incomplets', 'Certains votes n\'ont pas pu être enregistrés. Réessaie.')
  }
  else {
    toast.success('Votes enregistrés', 'Merci pour ton vote.')
  }
}

onMounted(load)
</script>

<template>
  <div
    v-if="loading"
    class="min-h-dvh flex items-center justify-center text-fg-muted text-sm"
  >
    Chargement…
  </div>

  <InviteMinimalState
    v-else-if="view === 'invalid'"
    icon="unlink"
    icon-tone="muted"
    title="Lien invalide"
    subtitle="Ce lien de vote n'existe pas ou n'est plus actif. Vérifiez le lien reçu de l'organisateur."
    legal="SMO ne stocke ni votre numéro ni votre email. Votre vote est lié au lien d'invitation reçu."
  />

  <InviteMinimalState
    v-else-if="view === 'not-participant'"
    icon="users"
    icon-tone="muted"
    title="Vous n'avez pas participé"
    subtitle="Seuls les joueurs ayant confirmé leur présence peuvent voter pour ce match."
    legal="SMO ne stocke ni votre numéro ni votre email. Vous pouvez fermer cette page."
  />

  <InviteMinimalState
    v-else-if="view === 'too-early' && context"
    icon="clock"
    icon-tone="muted"
    title="Le vote ouvrira après le match"
    :subtitle="`Rendez-vous ${formatMatchDate(context.scheduled_at)} à ${formatMatchTime(context.scheduled_at)}. Revenez sur ce lien une fois le match terminé.`"
    legal="SMO ne stocke ni votre numéro ni votre email. Votre vote est lié au lien d'invitation reçu."
  />

  <InviteMinimalState
    v-else-if="view === 'error'"
    icon="unlink"
    icon-tone="muted"
    title="Une erreur est survenue"
    subtitle="Impossible de charger la page de vote pour le moment."
    legal="SMO ne stocke ni votre numéro ni votre email."
    action-label="Réessayer"
    @action="load"
  />

  <main
    v-else-if="context"
    class="min-h-dvh max-w-[600px] mx-auto flex flex-col px-5 pt-6"
    :class="view === 'rate' ? 'pb-[120px]' : 'pb-10'"
  >
    <header class="flex justify-center pt-2 pb-5">
      <Wordmark :size="32" />
    </header>

    <div class="mb-2">
      <TeamMatchupCard
        :status="cardStatus"
        :date-label="dateLabel"
        :winner="cardWinner"
      />
    </div>

    <template v-if="view === 'rate'">
      <div
        class="flex items-center gap-2 bg-bg-elevated rounded-[var(--radius-md)] px-4 py-3 my-4 text-sm text-fg-default"
        role="note"
      >
        <span class="w-3 h-3 rounded-full shrink-0" :class="teamDotClass" />
        <span>Vous étiez dans l'<strong class="font-medium">Équipe {{ teamLabel }}</strong></span>
      </div>

      <section class="flex flex-col gap-2 mb-5">
        <h1 class="text-[22px] font-semibold tracking-[-0.01em] text-fg-default m-0">
          Notez vos coéquipiers
        </h1>
        <p class="text-sm text-fg-muted leading-[1.4] m-0 mb-3">
          Donnez une note de 1 à 5 étoiles à
          {{ context.teammates.length > 1 ? `chacun de vos ${context.teammates.length} coéquipiers` : 'votre coéquipier' }}.
        </p>
        <div class="flex flex-col gap-2">
          <TeammateRow
            v-for="teammate in context.teammates"
            :key="teammate.player_id"
            :teammate="teammate"
            :model-value="drafts[teammate.player_id] ?? 0"
            @update:model-value="setDraft(teammate.player_id, $event)"
          />
        </div>
        <div class="py-3 mt-2">
          <VoteProgress
            :done="filledCount"
            :total="context.teammates.length"
            label="notes"
          />
        </div>
      </section>

      <p class="text-[11px] leading-[1.5] text-fg-muted text-center mt-2">
        Le vote est définitif et anonyme.<br>
        Vos coéquipiers ne sauront pas qui les a notés.
      </p>

      <div class="fixed left-0 right-0 bottom-0 z-[5] px-5 pt-3 pb-[calc(theme(spacing.4)+env(safe-area-inset-bottom,0px))] bg-gradient-to-b from-transparent via-bg-base/80 to-bg-base">
        <div class="max-w-[600px] mx-auto">
          <button
            type="button"
            data-testid="submit-votes"
            class="inline-flex items-center justify-center gap-2 w-full h-[52px] rounded-[var(--radius-md)] border-0 text-base font-medium bg-action-primary text-fg-emphasis transition-[background,transform] duration-150 hover:bg-action-primary-hover active:translate-y-px disabled:bg-bg-elevated disabled:text-fg-muted disabled:opacity-60 disabled:cursor-not-allowed cursor-pointer"
            :disabled="!allRated || submitting"
            @click="modalOpen = true"
          >
            {{ allRated
              ? 'Soumettre mes votes (définitif)'
              : `Notez ${remaining.length > 1 ? `les ${remaining.length} coéquipiers` : 'votre coéquipier'}` }}
          </button>
        </div>
      </div>
    </template>

    <template v-else-if="view === 'submitted'">
      <section class="flex flex-col items-center text-center gap-3 pt-6 pb-2 mb-5">
        <span class="text-team-green inline-flex">
          <CircleCheck :size="64" />
        </span>
        <h1 class="text-[28px] leading-[1.2] font-semibold tracking-[-0.01em] m-0">
          Merci pour votre vote
        </h1>
        <p class="text-[15px] text-fg-muted m-0">Votre vote est enregistré</p>
      </section>

      <div class="bg-bg-elevated rounded-[var(--radius-lg)] p-4 flex flex-col gap-3 mb-6">
        <div class="flex items-baseline justify-between text-sm">
          <span class="font-semibold text-fg-default">Votes en cours</span>
          <span class="font-mono tabular-nums text-[13px] text-fg-default">
            {{ context.voters_done }}<span class="text-fg-muted"> / {{ context.voters_total }}</span>
          </span>
        </div>
        <VoteProgress
          :done="context.voters_done"
          :total="context.voters_total"
          label="joueurs ont voté"
        />
        <p class="text-[13px] text-fg-muted leading-[1.5] m-0">
          Les résultats seront disponibles dès que l'organisateur aura clôturé le match.
        </p>
      </div>

      <p class="mt-auto pt-8 text-[11px] leading-[1.5] text-fg-muted text-center">
        Vous pouvez fermer cette page.<br>
        Revenez sur ce lien pour consulter les résultats.
      </p>
    </template>

    <template v-else-if="view === 'results' && context.results">
      <p class="text-sm text-fg-muted leading-[1.4] mt-4 mb-5">
        Vos coéquipiers ont reçu vos notes.
      </p>

      <section class="flex flex-col gap-3 mb-6">
        <h2 class="text-[17px] font-semibold tracking-[-0.01em] text-fg-default m-0">
          Notes finales de l'équipe {{ teamLabel }}
        </h2>
        <div class="flex flex-col gap-2">
          <ResultRow
            v-for="result in context.results.teammates"
            :key="result.player_id"
            :result="result"
          />
        </div>
      </section>

      <section class="flex flex-col gap-3 mb-6">
        <h2 class="text-[17px] font-semibold tracking-[-0.01em] text-fg-default m-0">
          Votre note moyenne ce match
        </h2>
        <SelfScoreCard :initials="context.voter.initials" :self="context.results.self" />
      </section>

      <button
        type="button"
        class="inline-flex items-center justify-center gap-2 w-full h-[52px] rounded-[var(--radius-md)] bg-bg-elevated border border-border-default text-fg-default text-base font-medium opacity-50 cursor-not-allowed"
        disabled
        title="Réservé aux organisateurs connectés"
      >
        Voir tous les classements du groupe
        <ArrowRight :size="16" />
      </button>
    </template>

    <VoteConfirmModal
      :open="modalOpen"
      :busy="submitting"
      :teammates="remaining"
      :drafts="drafts"
      @confirm="confirmVotes"
      @cancel="modalOpen = false"
    />
  </main>
</template>
