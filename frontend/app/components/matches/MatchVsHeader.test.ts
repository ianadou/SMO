// @vitest-environment nuxt
import { describe, it, expect } from 'vitest'
import { mountSuspended } from '@nuxt/test-utils/runtime'
import MatchVsHeader from './MatchVsHeader.vue'
import type { MatchDTO } from '~/types/matches'

function makeMatch(over: Partial<MatchDTO> = {}): MatchDTO {
  return {
    id: 'm-1',
    group_id: 'g-1',
    title: 'matche 1',
    venue: 'Salle Pierre Mendès',
    scheduled_at: '2026-05-22T19:30:00Z',
    status: 'open',
    score_a: null,
    score_b: null,
    created_at: '2026-05-21T10:00:00Z',
    ...over,
  }
}

describe('MatchVsHeader', () => {
  it('renders the venue with a pin icon below the card', async () => {
    const wrapper = await mountSuspended(MatchVsHeader, { props: { match: makeMatch() } })
    const venue = wrapper.find('.md-header-venue')
    expect(venue.exists()).toBe(true)
    expect(venue.text()).toBe('Salle Pierre Mendès')
  })

  it('shows VS and a formatted date for an open match', async () => {
    const wrapper = await mountSuspended(MatchVsHeader, { props: { match: makeMatch() } })
    expect(wrapper.text()).toContain('VS')
    expect(wrapper.text()).toMatch(/\d{2}:\d{2}/)
  })

  it('shows the trophy with the red team color when the red team won', async () => {
    const wrapper = await mountSuspended(MatchVsHeader, {
      props: { match: makeMatch({ status: 'completed', score_a: 5, score_b: 3 }) },
    })
    expect(wrapper.find('svg.mc-center-trophy').classes()).toContain('is-red')
  })

  it('shows the trophy with the green team color when the green team won', async () => {
    const wrapper = await mountSuspended(MatchVsHeader, {
      props: { match: makeMatch({ status: 'completed', score_a: 1, score_b: 4 }) },
    })
    expect(wrapper.find('svg.mc-center-trophy').classes()).toContain('is-green')
  })

  it('emits back when the back button is clicked', async () => {
    const wrapper = await mountSuspended(MatchVsHeader, { props: { match: makeMatch() } })
    await wrapper.find('button[aria-label="Retour"]').trigger('click')
    expect(wrapper.emitted('back')).toHaveLength(1)
  })

  it('exposes the more-options button with an accessible label', async () => {
    const wrapper = await mountSuspended(MatchVsHeader, { props: { match: makeMatch() } })
    expect(wrapper.find('button[aria-label="Plus d\'options"]').exists()).toBe(true)
  })

  it('maps in_progress status to the live pill', async () => {
    const wrapper = await mountSuspended(MatchVsHeader, {
      props: { match: makeMatch({ status: 'in_progress' }) },
    })
    expect(wrapper.find('.mc-pill-live').exists()).toBe(true)
  })

  it('maps closed status to the "Clôturé" label', async () => {
    const wrapper = await mountSuspended(MatchVsHeader, {
      props: { match: makeMatch({ status: 'closed', score_a: 2, score_b: 1 }) },
    })
    expect(wrapper.text()).toContain('Clôturé')
  })
})
