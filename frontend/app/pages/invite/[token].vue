<script setup lang="ts">
import { ApiError } from '~/composables/useApi'
import { resolveInviteView } from '~/utils/inviteView'
import { formatMatchDate, formatMatchTime } from '~/utils/inviteFormat'
import InviteDetailsCard from '~/components/invite/InviteDetailsCard.vue'
import InviteConfirmed from '~/components/invite/InviteConfirmed.vue'
import InviteRespondModal from '~/components/invite/InviteRespondModal.vue'
import InviteMinimalState from '~/components/invite/InviteMinimalState.vue'
import InviteIcon from '~/components/invite/InviteIcon.vue'
import LegalFooter from '~/components/app/LegalFooter.vue'
import type { InvitationContext, LoadOutcome, RespondResult } from '~/types/invitation'

definePageMeta({ layout: false })
useHead({ title: 'Invitation — SMO' })

const route = useRoute()
const tokenParam = route.params.token
const token = (Array.isArray(tokenParam) ? tokenParam[0] : tokenParam) ?? ''

const outcome = ref<LoadOutcome | null>(null)
const loading = ref(true)
const modalOpen = ref(false)
const submitting = ref(false)

const context = computed<InvitationContext | null>(() =>
  outcome.value?.kind === 'ok' ? outcome.value.context : null,
)
const view = computed(() => (outcome.value ? resolveInviteView(outcome.value) : 'error'))
const isYes = computed(() => context.value?.response === 'yes')

async function load() {
  const api = useApi()
  loading.value = true
  try {
    const ctx = await api.post<InvitationContext>('/invitations/context', { token })
    if (ctx.match_status === 'completed' || ctx.match_status === 'closed') {
      await navigateTo(`/vote/${token}`, { replace: true })
      return
    }
    outcome.value = { kind: 'ok', context: ctx }
  }
  catch (e) {
    if (e instanceof ApiError && (e.status === 404 || e.status === 400)) {
      outcome.value = { kind: 'invalid' }
    }
    else {
      outcome.value = { kind: 'error' }
    }
  }
  finally {
    loading.value = false
  }
}

async function submit(answer: 'yes' | 'no') {
  const api = useApi()
  submitting.value = true
  try {
    const result = await api.post<RespondResult>('/invitations/respond', { token, answer })
    if (outcome.value?.kind === 'ok') {
      outcome.value = {
        kind: 'ok',
        context: { ...outcome.value.context, response: result.response },
      }
    }
    modalOpen.value = false
  }
  catch (e) {
    modalOpen.value = false
    if (e instanceof ApiError && e.status === 410) {
      if (outcome.value?.kind === 'ok') {
        outcome.value = { kind: 'ok', context: { ...outcome.value.context, state: 'expired' } }
      }
    }
    else if (e instanceof ApiError && e.status === 409) {
      await load()
    }
    else {
      outcome.value = { kind: 'error' }
    }
  }
  finally {
    submitting.value = false
  }
}

onMounted(async () => {
  await load()
  if (route.query.respond === '1' && view.value === 'initial') {
    modalOpen.value = true
  }
})
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
    subtitle="Ce lien d'invitation n'existe pas ou n'est plus actif. Vérifiez le lien reçu de l'organisateur."
    legal="SMO ne stocke ni votre numéro ni votre email. Votre réponse est liée au lien d'invitation reçu."
  />

  <InviteMinimalState
    v-else-if="view === 'expired'"
    icon="clock"
    icon-tone="warn"
    title="Invitation expirée"
    subtitle="La date limite de réponse est passée. Contactez l'organisateur si vous souhaitez encore participer."
    legal="SMO ne stocke ni votre numéro ni votre email. Vous pouvez fermer cette page."
  />

  <InviteMinimalState
    v-else-if="view === 'locked-pending'"
    icon="users"
    icon-tone="muted"
    title="Réponses closes"
    subtitle="Les équipes ont été formées, les réponses ne sont plus acceptées pour ce match."
    legal="SMO ne stocke ni votre numéro ni votre email. Vous pouvez fermer cette page."
  />

  <InviteMinimalState
    v-else-if="view === 'error'"
    icon="unlink"
    icon-tone="muted"
    title="Une erreur est survenue"
    subtitle="Impossible de charger l'invitation pour le moment."
    legal="SMO ne stocke ni votre numéro ni votre email."
    action-label="Réessayer"
    @action="load"
  />

  <main
    v-else
    class="min-h-dvh max-w-[500px] mx-auto flex flex-col px-5 pt-6 pb-10"
  >
    <header class="flex justify-center pt-2 pb-8">
      <Wordmark :size="32" />
    </header>

    <template v-if="view === 'initial' && context">
      <section class="flex flex-col gap-2 mb-6">
        <h1 class="text-[28px] leading-[1.2] font-semibold tracking-[-0.01em] text-fg-default m-0">
          Vous êtes invité
        </h1>
        <p class="text-[15px] leading-[1.5] text-fg-muted m-0">
          <strong class="text-fg-default font-medium">{{ context.organizer_name }}</strong>
          vous invite au match du groupe
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

      <div class="mb-8">
        <InviteConfirmed
          :confirmed-count="context.confirmed_count"
          :max-participants="context.max_participants"
          :initials="context.confirmed_initials"
        />
      </div>

      <button
        type="button"
        data-testid="respond-cta"
        class="inline-flex items-center justify-center gap-2 w-full h-[52px] rounded-[var(--radius-md)] border-0 text-base font-medium cursor-pointer bg-action-primary text-fg-emphasis transition-[background,transform] duration-150 hover:bg-action-primary-hover active:translate-y-px"
        @click="modalOpen = true"
      >
        <InviteIcon name="check-circle" :size="18" />
        <span>Répondre</span>
      </button>

      <p class="mt-auto pt-8 text-[11px] leading-[1.5] text-fg-muted text-center">
        SMO ne stocke ni votre numéro ni votre email.<br>
        Votre réponse est liée au lien d'invitation reçu.
      </p>

      <LegalFooter class="mt-3 pb-2" />
    </template>

    <template v-else-if="view === 'result' && context">
      <section class="flex flex-col items-center text-center gap-3 pt-4 pb-2 mb-6">
        <span :class="isYes ? 'text-team-green' : 'text-team-red'" class="inline-flex">
          <InviteIcon :name="isYes ? 'check-circle-filled' : 'x-circle-filled'" :size="64" />
        </span>
        <h1 class="text-[28px] leading-[1.2] font-semibold tracking-[-0.01em] m-0">
          {{ isYes ? 'Vous êtes inscrit' : 'Réponse enregistrée' }}
        </h1>
        <p class="text-[15px] text-fg-muted m-0">
          <template v-if="isYes">
            Rendez-vous
            <strong class="text-fg-default font-medium">{{ formatMatchDate(context.scheduled_at) }}</strong>
            à
            <span class="text-fg-default font-mono tabular-nums">{{ formatMatchTime(context.scheduled_at) }}</span>
          </template>
          <template v-else>
            Vous ne participerez pas à ce match
          </template>
        </p>
      </section>

      <div class="mb-6">
        <InviteDetailsCard
          :scheduled-at="context.scheduled_at"
          :venue="context.venue"
          :capacity="context.capacity"
          compact
        />
      </div>

      <div
        v-if="context.state === 'locked'"
        class="bg-bg-elevated border border-border-default rounded-[var(--radius-md)] px-4 py-3 flex items-center gap-3 text-[13px] text-fg-muted leading-[1.4]"
      >
        <span class="shrink-0 inline-flex"><InviteIcon name="users" :size="18" /></span>
        <span>Les équipes ont été formées, votre réponse ne peut plus être modifiée.</span>
      </div>
      <button
        v-else
        type="button"
        data-testid="modify-cta"
        class="inline-flex items-center justify-center w-full h-[52px] rounded-[var(--radius-md)] bg-bg-elevated border border-border-default text-fg-default text-base font-medium cursor-pointer transition-colors duration-150 hover:bg-bg-subtle"
        @click="modalOpen = true"
      >
        Modifier ma réponse
      </button>

      <p class="mt-auto pt-8 text-[11px] leading-[1.5] text-fg-muted text-center">
        Vous pouvez fermer cette page.<br>
        Votre réponse est sauvegardée.
      </p>
    </template>

    <InviteRespondModal
      :open="modalOpen"
      :busy="submitting"
      @answer="submit"
      @cancel="modalOpen = false"
    />
  </main>
</template>
