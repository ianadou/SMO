// @vitest-environment nuxt
import { describe, it, expect } from 'vitest'
import { mountSuspended } from '@nuxt/test-utils/runtime'
import PresentList from './PresentList.vue'
import type { TeamMemberDTO } from '~/types/matches'

const red: TeamMemberDTO[] = [
  { player_id: 'r1', player_name: 'Alice Martin', team: 'A', slot: 0 },
  { player_id: 'r2', player_name: 'Bob', team: 'A', slot: 1 },
]
const green: TeamMemberDTO[] = [
  { player_id: 'g1', player_name: 'Léa Thomas', team: 'B', slot: 0 },
]

describe('PresentList', () => {
  it('renders a chip for every player on both teams', async () => {
    const wrapper = await mountSuspended(PresentList, { props: { red, green } })

    const chips = wrapper.findAll('.md-chip')
    expect(chips).toHaveLength(3)
    const labels = chips.map((c) => c.text())
    expect(labels.some((t) => t.includes('Alice Martin'))).toBe(true)
    expect(labels.some((t) => t.includes('Bob'))).toBe(true)
    expect(labels.some((t) => t.includes('Léa Thomas'))).toBe(true)
  })

  it('shows the placed count as the sum of both teams', async () => {
    const wrapper = await mountSuspended(PresentList, { props: { red, green } })

    expect(wrapper.findAll('.num')[0]!.text()).toBe('3')
  })

  it('renders each chip avatar with the player initials', async () => {
    const wrapper = await mountSuspended(PresentList, { props: { red, green } })

    const avatars = wrapper.findAll('.md-chip-avatar').map((a) => a.text())
    expect(avatars).toEqual(['AM', 'BO', 'LT'])
  })

  it('separates red chips from green chips by class', async () => {
    const wrapper = await mountSuspended(PresentList, { props: { red, green } })

    expect(wrapper.findAll('.md-chip.is-red')).toHaveLength(2)
    expect(wrapper.findAll('.md-chip.is-green')).toHaveLength(1)
  })
})
