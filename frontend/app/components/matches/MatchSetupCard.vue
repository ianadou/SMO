<script setup lang="ts">
import { ref } from 'vue'

defineProps<{
  kind: 'draft' | 'generate'
  busy?: boolean
}>()

const emit = defineEmits<{
  open: []
  generate: [strategy: 'random' | 'ranking']
}>()

const strategy = ref<'random' | 'ranking'>('ranking')
</script>

<template>
  <div class="md-setup">
    <div class="md-setup-card">
      <template v-if="kind === 'draft'">
        <div class="md-setup-title">Match en brouillon</div>
        <div class="md-setup-sub">Ouvre le match pour inviter et composer les équipes.</div>
        <button class="md-setup-btn" :disabled="busy" @click="emit('open')">
          Ouvrir le match
        </button>
      </template>
      <template v-else>
        <span class="md-setup-label">Stratégie</span>
        <div class="md-seg">
          <button
            type="button"
            :class="{ 'is-on': strategy === 'random' }"
            @click="strategy = 'random'"
          >
            Aléatoire
          </button>
          <button
            type="button"
            :class="{ 'is-on': strategy === 'ranking' }"
            @click="strategy = 'ranking'"
          >
            Classement
          </button>
        </div>
        <button
          class="md-setup-btn"
          :disabled="busy"
          @click="emit('generate', strategy)"
        >
          Générer les équipes
        </button>
      </template>
    </div>
  </div>
</template>
