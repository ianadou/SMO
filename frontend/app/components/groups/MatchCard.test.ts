// @vitest-environment nuxt
import { describe, it, expect } from 'vitest'
import { mountSuspended } from '@nuxt/test-utils/runtime'
import MatchCard from './MatchCard.vue'

describe('groups/MatchCard', () => {
  it('renders a clickable card linking to the match detail page', async () => {
    const wrapper = await mountSuspended(MatchCard, {
      props: { id: 'abc', scheduledAt: '2026-05-22T19:30:00Z', status: 'open' },
    })
    const link = wrapper.find('a.match-card-link')
    expect(link.exists()).toBe(true)
    expect(link.attributes('href')).toBe('/matches/abc')
  })

  it('shows VS for an upcoming open match', async () => {
    const wrapper = await mountSuspended(MatchCard, {
      props: { id: 'abc', scheduledAt: '2026-05-22T19:30:00Z', status: 'open' },
    })
    expect(wrapper.text()).toContain('VS')
  })

  it('shows the trophy in red when the red team won', async () => {
    const wrapper = await mountSuspended(MatchCard, {
      props: {
        id: 'abc',
        scheduledAt: '2026-05-04T19:30:00Z',
        status: 'completed',
        scoreA: 5,
        scoreB: 3,
      },
    })
    expect(wrapper.find('svg.mc-center-trophy').classes()).toContain('is-red')
    expect(wrapper.text()).toContain('Terminé')
  })

  it('shows the trophy in green when the green team won', async () => {
    const wrapper = await mountSuspended(MatchCard, {
      props: {
        id: 'abc',
        scheduledAt: '2026-05-04T19:30:00Z',
        status: 'closed',
        scoreA: 1,
        scoreB: 4,
      },
    })
    expect(wrapper.find('svg.mc-center-trophy').classes()).toContain('is-green')
    expect(wrapper.text()).toContain('Clôturé')
  })

  it('emits the live pill for an in_progress match', async () => {
    const wrapper = await mountSuspended(MatchCard, {
      props: { id: 'abc', scheduledAt: '2026-05-22T19:30:00Z', status: 'in_progress' },
    })
    expect(wrapper.find('.mc-pill-live').exists()).toBe(true)
  })

  it('omits the winner trophy when scores are equal', async () => {
    const wrapper = await mountSuspended(MatchCard, {
      props: {
        id: 'abc',
        scheduledAt: '2026-05-04T19:30:00Z',
        status: 'completed',
        scoreA: 2,
        scoreB: 2,
      },
    })
    const trophy = wrapper.find('svg.mc-center-trophy')
    expect(trophy.exists()).toBe(true)
    expect(trophy.classes()).toContain('is-red')
  })
})
