<script setup lang="ts">
import { computed, onMounted, ref, watch } from 'vue'
import MatchVsHeader from '~/components/matches/MatchVsHeader.vue'
import TeamField from '~/components/matches/TeamField.vue'
import PresentList from '~/components/matches/PresentList.vue'
import MatchSetupCard from '~/components/matches/MatchSetupCard.vue'
import MatchValidateBar from '~/components/matches/MatchValidateBar.vue'
import { useMatchDetail } from '~/composables/useMatchDetail'
import { useTeamDrag } from '~/composables/useTeamDrag'
import { splitTeams, toTeamArrays } from '~/utils/teamComposition'
import type { TeamMemberDTO } from '~/types/matches'

definePageMeta({ layout: false, middleware: 'auth' })

const route = useRoute()
const matchId = route.params.id as string

const detail = useMatchDetail(matchId)
const { match, members, screen, loading, error } = detail

const red = ref<TeamMemberDTO[]>([])
const green = ref<TeamMemberDTO[]>([])
const drag = useTeamDrag(red, green)

watch(
  members,
  (list) => {
    const split = splitTeams(list)
    red.value = split.teamA
    green.value = split.teamB
  },
  { immediate: true },
)

const isCompositionSaved = computed(() => match.value?.status === 'teams_ready')

const validateLabel = computed(() =>
  isCompositionSaved.value ? 'Enregistrer la composition' : 'Valider les équipes',
)

useHead(() => ({
  title: match.value ? `${match.value.title} — SMO` : 'Match — SMO',
}))

onMounted(() => detail.load())

function back() {
  navigateTo(match.value ? `/groups/${match.value.group_id}` : '/groups')
}

function validate() {
  const { team_a, team_b } = toTeamArrays(red.value, green.value)
  return isCompositionSaved.value
    ? detail.saveTeams(team_a, team_b)
    : detail.validateTeams(team_a, team_b)
}
</script>

<template>
  <div :class="['md-screen', { 'has-bottombar': screen === 'composition' }]">
    <MatchVsHeader v-if="match" :match="match" @back="back" />

    <div v-if="loading && !match" class="md-state">Chargement…</div>

    <div v-else-if="error && !match" class="md-state">
      <span>{{ error }}</span>
      <button @click="detail.load()">Réessayer</button>
    </div>

    <template v-else-if="match">
      <MatchSetupCard
        v-if="screen === 'setup-draft'"
        kind="draft"
        :busy="loading"
        @open="detail.openMatch()"
      />

      <MatchSetupCard
        v-else-if="screen === 'setup-generate'"
        kind="generate"
        :busy="loading"
        @generate="detail.generate()"
      />

      <template v-else>
        <div class="md-field-wrap">
          <div class="md-field-meta">
            <span>Composition</span>
            <span class="md-field-meta-state">
              <template v-if="screen === 'composition'">Glisse pour échanger</template>
              <template v-else-if="screen === 'locked'">Verrouillée · coup d'envoi imminent</template>
              <template v-else-if="screen === 'finished'">Match en cours · lecture seule</template>
              <template v-else>Clôturé</template>
            </span>
          </div>
          <TeamField
            :team-a="red"
            :team-b="green"
            :mode="screen === 'composition' ? 'edit' : 'view'"
            :drag="drag.drag.value"
            @pointerdown="drag.onPointerDown"
          />
        </div>

        <PresentList :red="red" :green="green" />

        <MatchValidateBar
          v-if="screen === 'composition'"
          :busy="loading"
          :label="validateLabel"
          @validate="validate"
        />
      </template>
    </template>
  </div>
</template>
