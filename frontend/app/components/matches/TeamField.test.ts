// @vitest-environment nuxt
import { describe, it, expect } from 'vitest'
import { mountSuspended } from '@nuxt/test-utils/runtime'
import TeamField from './TeamField.vue'
import type { TeamMemberDTO } from '~/types/matches'

function team(side: 'A' | 'B', count: number): TeamMemberDTO[] {
  return Array.from({ length: count }, (_, i) => ({
    player_id: `${side}${i}`,
    player_name: `Player ${side}${i}`,
    team: side,
    slot: i,
  }))
}

describe('TeamField', () => {
  it('renders one pion per member of both teams', async () => {
    const wrapper = await mountSuspended(TeamField, {
      props: { teamA: team('A', 3), teamB: team('B', 2), mode: 'view' },
    })

    expect(wrapper.findAll('.md-pion')).toHaveLength(5)
  })

  it('caps each team at five pions', async () => {
    const wrapper = await mountSuspended(TeamField, {
      props: { teamA: team('A', 6), teamB: team('B', 0), mode: 'view' },
    })

    expect(wrapper.findAll('.md-pion')).toHaveLength(5)
  })

  it('forwards pointerdown from a pion in edit mode', async () => {
    const teamA = team('A', 1)
    const wrapper = await mountSuspended(TeamField, {
      props: { teamA, teamB: team('B', 0), mode: 'edit' },
    })

    await wrapper.find('.md-pion').trigger('pointerdown')

    expect(wrapper.emitted('pointerdown')).toHaveLength(1)
    expect(wrapper.emitted('pointerdown')![0]![0]).toEqual(teamA[0])
  })

  it('does not forward pointerdown in view mode', async () => {
    const wrapper = await mountSuspended(TeamField, {
      props: { teamA: team('A', 1), teamB: team('B', 0), mode: 'view' },
    })

    await wrapper.find('.md-pion').trigger('pointerdown')

    expect(wrapper.emitted('pointerdown')).toBeUndefined()
  })
})
