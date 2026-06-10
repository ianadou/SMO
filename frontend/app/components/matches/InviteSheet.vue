<script setup lang="ts">
import { Loader2, Share2, UserPlus } from 'lucide-vue-next'
import BaseSheet from '~/components/BaseSheet.vue'
import AvatarRing from '~/components/ui/AvatarRing.vue'
import { playerInitials } from '~/utils/playerInitials'
import type { InviteRow } from '~/types/invitation'

withDefaults(
  defineProps<{
    open: boolean
    rows: InviteRow[]
    confirmedCount: number
    loading?: boolean
    invitingId?: string | null
    locked?: boolean
    failed?: boolean
  }>(),
  { loading: false, invitingId: null, locked: false, failed: false },
)

const emit = defineEmits<{
  invite: [playerId: string]
  share: [row: InviteRow]
  close: []
  retry: []
}>()

const chipByStatus = {
  'pending': { label: 'En attente', class: 'bg-warn/10 text-warn' },
  'yes': { label: '✓ Vient', class: 'bg-team-green/10 text-team-green' },
  'no': { label: '✗ Absent', class: 'bg-team-red/10 text-team-red' },
  'not-invited': { label: 'Non invité', class: 'bg-bg-subtle text-fg-muted' },
} as const
</script>

<template>
  <BaseSheet :open="open" title="Invitations" @close="emit('close')">
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
  </BaseSheet>
</template>
