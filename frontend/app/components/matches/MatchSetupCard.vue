<script setup lang="ts">
defineProps<{
  kind: 'draft' | 'generate'
  busy?: boolean
}>()

defineEmits<{
  open: []
  generate: []
}>()
</script>

<template>
  <div class="md-setup">
    <div class="md-setup-card">
      <template v-if="kind === 'draft'">
        <div class="md-setup-title">Match en brouillon</div>
        <div class="md-setup-sub">
          Ouvre le match pour générer les équipes et inviter les joueurs.
        </div>
        <button class="md-setup-btn" :disabled="busy" @click="$emit('open')">
          Ouvrir le match
        </button>
      </template>
      <template v-else>
        <div class="md-setup-title">Équipes en attente</div>
        <div class="md-setup-sub">
          La génération automatique n'a pas abouti. Vérifie que tu as assez de joueurs confirmés, puis relance.
        </div>
        <button class="md-setup-btn" :disabled="busy" @click="$emit('generate')">
          Relancer la génération
        </button>
      </template>
    </div>
  </div>
</template>
