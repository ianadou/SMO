// @vitest-environment nuxt
import { describe, it, expect } from 'vitest'
import { mountSuspended } from '@nuxt/test-utils/runtime'
import TeamMatchupCard from './TeamMatchupCard.vue'

describe('TeamMatchupCard', () => {
  it('renders VS, both jerseys, and team labels for an upcoming match', async () => {
    const wrapper = await mountSuspended(TeamMatchupCard, {
      props: { status: 'upcoming', dateLabel: 'Ven. 1 mai', timeLabel: '21:30' },
    })
    expect(wrapper.text()).toContain('VS')
    expect(wrapper.text()).toContain('Ven. 1 mai')
    expect(wrapper.text()).toContain('21:30')
    expect(wrapper.text()).toContain('Équipe rouge')
    expect(wrapper.text()).toContain('Équipe verte')
    expect(wrapper.findAll('svg.sjersey')).toHaveLength(2)
  })

  it('shows the trophy in red when the red team won a finished match', async () => {
    const wrapper = await mountSuspended(TeamMatchupCard, {
      props: { status: 'finished', dateLabel: 'Jeu. 4 juin', winner: 'red' },
    })
    expect(wrapper.text()).toContain('Terminé')
    expect(wrapper.text()).not.toContain('VS')
    const trophy = wrapper.find('svg.mc-center-trophy')
    expect(trophy.exists()).toBe(true)
    expect(trophy.classes()).toContain('is-red')
  })

  it('renders the votes footer only for closed matches with a votes label', async () => {
    const wrapper = await mountSuspended(TeamMatchupCard, {
      props: {
        status: 'closed',
        dateLabel: 'Jeu. 17 avril',
        winner: 'green',
        votesLabel: '12 votes enregistrés',
      },
    })
    expect(wrapper.text()).toContain('Clôturé')
    expect(wrapper.text()).toContain('12 votes enregistrés')
    expect(wrapper.find('svg.mc-center-trophy').classes()).toContain('is-green')
  })

  it('renders the live pill when the status is live', async () => {
    const wrapper = await mountSuspended(TeamMatchupCard, {
      props: { status: 'live', dateLabel: 'Aujourd\'hui' },
    })
    expect(wrapper.text()).toContain('En cours')
    expect(wrapper.find('.mc-pill-live').exists()).toBe(true)
  })

  it('renders the cancelled pill and a cancelled center icon', async () => {
    const wrapper = await mountSuspended(TeamMatchupCard, {
      props: { status: 'cancelled', dateLabel: 'Jeu. 24 avril' },
    })
    expect(wrapper.text()).toContain('Annulé')
    expect(wrapper.find('.mc-pill-cancelled').exists()).toBe(true)
    expect(wrapper.classes()).toContain('is-cancelled')
  })

  it('falls back to a div root by default and a button when interactive=true', async () => {
    const staticWrap = await mountSuspended(TeamMatchupCard, {
      props: { status: 'upcoming', dateLabel: 'Date' },
    })
    expect(staticWrap.element.tagName).toBe('DIV')

    const interactiveWrap = await mountSuspended(TeamMatchupCard, {
      props: { status: 'upcoming', dateLabel: 'Date', interactive: true },
    })
    expect(interactiveWrap.element.tagName).toBe('BUTTON')
  })

  it('emits click only when interactive', async () => {
    const interactiveWrap = await mountSuspended(TeamMatchupCard, {
      props: { status: 'upcoming', dateLabel: 'Date', interactive: true },
    })
    await interactiveWrap.trigger('click')
    expect(interactiveWrap.emitted('click')).toHaveLength(1)
  })
})
