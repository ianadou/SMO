<script setup lang="ts">
import { Copy, Link, Loader2, Share2, UserPlus } from 'lucide-vue-next'
import BaseSheet from '~/components/BaseSheet.vue'
import BaseModal from '~/components/BaseModal.vue'
import AvatarRing from '~/components/ui/AvatarRing.vue'
import { playerInitials } from '~/utils/playerInitials'
import type { InviteRow } from '~/types/invitation'
import type { MatchShareLinkDTO } from '~/types/shareLink'

const props = withDefaults(
  defineProps<{
    open: boolean
    rows: InviteRow[]
    confirmedCount: number
    loading?: boolean
    invitingId?: string | null
    locked?: boolean
    failed?: boolean
    shareLink?: MatchShareLinkDTO | null
    linkBusy?: boolean
  }>(),
  { loading: false, invitingId: null, locked: false, failed: false, shareLink: null, linkBusy: false },
)

const emit = defineEmits<{
  'invite': [playerId: string]
  'share': [row: InviteRow]
  'close': []
  'retry': []
  'generate-link': []
  'copy-link': []
  'revoke-link': []
}>()

const chipByStatus = {
  'pending': { label: 'En attente', class: 'bg-warn/10 text-warn' },
  'yes': { label: '✓ Vient', class: 'bg-team-green/10 text-team-green' },
  'no': { label: '✗ Absent', class: 'bg-team-red/10 text-team-red' },
  'not-invited': { label: 'Non invité', class: 'bg-bg-subtle text-fg-muted' },
} as const

const regenerateConfirmOpen = ref(false)
const revokeConfirmOpen = ref(false)

const linkExpiresLabel = computed(() =>
  props.shareLink
    ? new Date(props.shareLink.expires_at).toLocaleDateString('fr-FR', {
        day: 'numeric',
        month: 'long',
        year: 'numeric',
      })
    : '',
)

function confirmRegenerate() {
  regenerateConfirmOpen.value = false
  emit('generate-link')
}

function confirmRevoke() {
  revokeConfirmOpen.value = false
  emit('revoke-link')
}
</script>

<template>
  <BaseSheet :open="open" title="Invitations" @close="emit('close')">
    <section class="bg-bg-base border border-border-default rounded-[var(--radius-md)] p-3 mb-4 flex flex-col gap-2.5">
      <span class="text-[11px] font-semibold tracking-[0.08em] uppercase text-fg-muted">
        Lien du match
      </span>

      <button
        v-if="!shareLink"
        type="button"
        data-testid="generate-link"
        class="inline-flex items-center justify-center gap-2 w-full h-11 rounded-[var(--radius-md)] border-0 bg-action-primary text-fg-emphasis text-sm font-semibold cursor-pointer hover:bg-action-primary-hover active:translate-y-px disabled:opacity-60 disabled:cursor-not-allowed"
        :disabled="linkBusy"
        @click="emit('generate-link')"
      >
        <Loader2 v-if="linkBusy" :size="16" class="animate-spin" />
        <Link v-else :size="16" />
        Générer le lien du match
      </button>

      <template v-else>
        <div class="flex items-center gap-2">
          <span
            data-testid="share-link-url"
            class="flex-1 min-w-0 truncate font-mono text-[13px] text-fg-default"
          >{{ shareLink.url }}</span>
          <button
            type="button"
            data-testid="copy-link"
            class="inline-flex items-center gap-1.5 h-9 px-3 rounded-[var(--radius-md)] border-0 bg-action-primary text-fg-emphasis text-[13px] font-semibold cursor-pointer hover:bg-action-primary-hover active:translate-y-px shrink-0"
            @click="emit('copy-link')"
          >
            <Copy :size="14" />
            Copier
          </button>
        </div>
        <span class="text-[11px] text-fg-muted">Expire le {{ linkExpiresLabel }}</span>
        <div class="flex gap-2">
          <button
            type="button"
            data-testid="regenerate-link"
            class="flex-1 h-9 rounded-[var(--radius-md)] bg-transparent border border-border-default text-fg-default text-[13px] cursor-pointer hover:bg-white/5 disabled:opacity-60 disabled:cursor-not-allowed"
            :disabled="linkBusy"
            @click="regenerateConfirmOpen = true"
          >
            Régénérer
          </button>
          <button
            type="button"
            data-testid="revoke-link"
            class="flex-1 h-9 rounded-[var(--radius-md)] bg-transparent border border-team-red/40 text-team-red text-[13px] cursor-pointer hover:bg-team-red/10 disabled:opacity-60 disabled:cursor-not-allowed"
            :disabled="linkBusy"
            @click="revokeConfirmOpen = true"
          >
            Révoquer
          </button>
        </div>
      </template>
    </section>

    <p class="text-[13px] text-fg-muted m-0 mb-3">
      <span class="font-mono tabular-nums text-fg-default">{{ confirmedCount }}</span>
      confirmé{{ confirmedCount > 1 ? 's' : '' }} · chaque joueur reçoit son lien personnel,
      visible une seule fois.
    </p>

    <div v-if="loading" class="py-8 text-center text-sm text-fg-muted">
      Chargement…
    </div>

    <div v-else-if="failed" class="py-6 flex flex-col items-center gap-3">
      <p class="text-sm text-fg-muted m-0">Impossible de charger les invitations.</p>
      <button
        type="button"
        data-testid="retry-invitations"
        class="px-4 h-10 rounded-[var(--radius-md)] bg-bg-subtle border border-border-default text-fg-default text-sm cursor-pointer hover:bg-white/5"
        @click="emit('retry')"
      >
        Réessayer
      </button>
    </div>

    <div v-else-if="rows.length === 0" class="py-8 text-center text-sm text-fg-muted">
      Aucun joueur dans ce groupe. Ajoute des joueurs depuis la page du groupe.
    </div>

    <div v-else class="flex flex-col gap-2">
      <div
        v-for="row in rows"
        :key="row.playerId"
        :data-testid="`invite-row-${row.playerId}`"
        class="bg-bg-base border border-border-default rounded-[var(--radius-md)] px-3 py-2.5 flex items-center gap-3"
        :class="row.status === 'fresh' ? 'border-action-primary bg-action-primary/5' : ''"
      >
        <AvatarRing :initials="playerInitials(row.playerName)" :size="36" />
        <span class="flex-1 min-w-0 text-sm font-medium text-fg-default truncate">
          {{ row.playerName }}
        </span>

        <span
          v-if="row.claimed"
          :data-testid="`claimed-${row.playerId}`"
          class="text-[11px] font-medium rounded-full px-2 py-0.5 bg-action-primary/10 text-action-primary shrink-0"
        >
          réclamé ✓
        </span>

        <button
          v-if="row.status === 'fresh'"
          type="button"
          :data-testid="`share-${row.playerId}`"
          class="inline-flex items-center gap-1.5 h-9 px-3 rounded-[var(--radius-md)] border-0 bg-action-primary text-fg-emphasis text-[13px] font-semibold cursor-pointer hover:bg-action-primary-hover active:translate-y-px"
          @click="emit('share', row)"
        >
          <Share2 :size="14" />
          Partager le lien
        </button>

        <button
          v-else-if="row.status === 'not-invited' && !locked"
          type="button"
          :data-testid="`invite-${row.playerId}`"
          class="inline-flex items-center gap-1.5 h-9 px-3 rounded-[var(--radius-md)] bg-transparent border border-border-default text-fg-default text-[13px] cursor-pointer hover:bg-white/5 disabled:opacity-60 disabled:cursor-not-allowed"
          :disabled="invitingId !== null"
          @click="emit('invite', row.playerId)"
        >
          <Loader2 v-if="invitingId === row.playerId" :size="14" class="animate-spin" />
          <UserPlus v-else :size="14" />
          Inviter
        </button>

        <span
          v-else
          class="text-[11px] font-medium rounded-full px-2.5 py-1"
          :class="chipByStatus[row.status === 'fresh' ? 'pending' : row.status].class"
        >
          {{ chipByStatus[row.status === 'fresh' ? 'pending' : row.status].label }}
        </span>
      </div>
    </div>

    <p v-if="locked" class="text-[11px] text-fg-muted text-center mt-3 mb-0">
      Les équipes sont formées — les réponses sont verrouillées.
    </p>

    <BaseModal
      :open="regenerateConfirmOpen"
      variant="confirm"
      title="Régénérer le lien ?"
      @close="regenerateConfirmOpen = false"
    >
      <p class="text-[13px] leading-[1.5] text-fg-muted text-center m-0">
        L'ancien lien du match cessera de fonctionner immédiatement.
      </p>
      <button
        type="button"
        data-testid="confirm-regenerate"
        class="inline-flex items-center justify-center w-full h-11 rounded-[var(--radius-md)] border-0 bg-action-primary text-fg-emphasis text-sm font-semibold cursor-pointer hover:bg-action-primary-hover active:translate-y-px"
        @click="confirmRegenerate"
      >
        Régénérer le lien
      </button>
      <button
        type="button"
        data-testid="cancel-regenerate"
        class="self-center bg-transparent border-0 cursor-pointer text-fg-muted text-sm underline underline-offset-[3px] px-4 py-2 hover:text-fg-default"
        @click="regenerateConfirmOpen = false"
      >
        Annuler
      </button>
    </BaseModal>

    <BaseModal
      :open="revokeConfirmOpen"
      variant="confirm"
      title="Révoquer le lien ?"
      @close="revokeConfirmOpen = false"
    >
      <p class="text-[13px] leading-[1.5] text-fg-muted text-center m-0">
        Le lien du match cessera de fonctionner immédiatement.
        Les liens personnels déjà remis restent valides.
      </p>
      <button
        type="button"
        data-testid="confirm-revoke"
        class="inline-flex items-center justify-center w-full h-11 rounded-[var(--radius-md)] border-0 bg-team-red text-fg-emphasis text-sm font-semibold cursor-pointer hover:brightness-110 active:translate-y-px"
        @click="confirmRevoke"
      >
        Révoquer le lien
      </button>
      <button
        type="button"
        data-testid="cancel-revoke"
        class="self-center bg-transparent border-0 cursor-pointer text-fg-muted text-sm underline underline-offset-[3px] px-4 py-2 hover:text-fg-default"
        @click="revokeConfirmOpen = false"
      >
        Annuler
      </button>
    </BaseModal>
  </BaseSheet>
</template>
