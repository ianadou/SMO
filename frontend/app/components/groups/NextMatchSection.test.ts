// @vitest-environment nuxt
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'
import { flushPromises } from '@vue/test-utils'
import { mountSuspended } from '@nuxt/test-utils/runtime'
import NextMatchSection from './NextMatchSection.vue'
import type { MatchDTO } from '~/types/matches'

const getMock = vi.fn()

vi.mock('~/composables/useApi', () => ({
  useApi: () => ({ get: getMock }),
}))

function makeMatch(over: Partial<MatchDTO> = {}): MatchDTO {
  return {
    id: 'next-1',
    group_id: 'g-1',
    title: 'matche B',
    venue: 'Salle Pierre Mendès',
    scheduled_at: '2026-06-11T19:30:00Z',
    status: 'open',
    score_a: null,
    score_b: null,
    created_at: '2026-05-22T10:00:00Z',
    ...over,
  }
}

describe('NextMatchSection', () => {
  beforeEach(() => {
    getMock.mockReset()
  })

  afterEach(() => {
    getMock.mockReset()
  })

  it('renders eyebrow, venue and a link to the match detail', async () => {
    getMock.mockResolvedValue([])
    const wrapper = await mountSuspended(NextMatchSection, { props: { match: makeMatch() } })
    expect(wrapper.text()).toContain('Prochain match')
    expect(wrapper.find('a').attributes('href')).toBe('/matches/next-1')
    expect(wrapper.text()).toContain('Salle Pierre Mendès')
  })

  it('shows the confirmed count after the participants endpoint resolves', async () => {
    getMock.mockResolvedValue(Array.from({ length: 8 }, (_, i) => ({ player_id: `p-${i}` })))
    const wrapper = await mountSuspended(NextMatchSection, { props: { match: makeMatch() } })
    await flushPromises()
    expect(getMock).toHaveBeenCalledWith('/matches/next-1/participants')
    expect(wrapper.text()).toContain('8')
    expect(wrapper.text()).toContain('/10 confirmés')
  })

  it('omits the count when the participants endpoint fails', async () => {
    getMock.mockRejectedValue(new Error('boom'))
    const wrapper = await mountSuspended(NextMatchSection, { props: { match: makeMatch() } })
    await flushPromises()
    expect(wrapper.text()).not.toContain('confirmés')
  })

  it('renders the upcoming VS layout via TeamMatchupCard', async () => {
    getMock.mockResolvedValue([])
    const wrapper = await mountSuspended(NextMatchSection, { props: { match: makeMatch() } })
    expect(wrapper.text()).toContain('VS')
    expect(wrapper.findAll('svg.sjersey')).toHaveLength(2)
  })
})
