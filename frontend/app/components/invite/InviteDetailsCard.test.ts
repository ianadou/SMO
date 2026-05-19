// @vitest-environment nuxt
import { describe, it, expect } from 'vitest'
import { mountSuspended } from '@nuxt/test-utils/runtime'
import InviteDetailsCard from './InviteDetailsCard.vue'

const props = {
  scheduledAt: '2026-05-07T19:30:00+02:00',
  venue: 'Salle Pierre Mendès, Lyon',
  capacity: '10 (5v5)',
}

describe('InviteDetailsCard', () => {
  it('renders the formatted date, time, venue and capacity', async () => {
    const wrapper = await mountSuspended(InviteDetailsCard, { props })
    const text = wrapper.text()
    expect(text).toContain('Jeudi 7 mai 2026')
    expect(text).toContain('19h30')
    expect(text).toContain('Salle Pierre Mendès, Lyon')
    expect(text).toContain('10')
    expect(text).toContain('5v5')
  })

  it('adds the compact class when compact', async () => {
    const wrapper = await mountSuspended(InviteDetailsCard, {
      props: { ...props, compact: true },
    })
    expect(wrapper.get('[data-testid="details-card"]').classes()).toContain('py-3')
  })
})
