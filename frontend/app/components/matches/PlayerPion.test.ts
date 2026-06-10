// @vitest-environment nuxt
import { describe, it, expect } from 'vitest'
import { mountSuspended } from '@nuxt/test-utils/runtime'
import PlayerPion from './PlayerPion.vue'
import type { TeamMemberDTO } from '~/types/matches'

const alice: TeamMemberDTO = { player_id: 'p1', player_name: 'Alice Martin', team: 'A', slot: 0 }
const bob: TeamMemberDTO = { player_id: 'p2', player_name: 'Bob', team: 'B', slot: 0 }

function mount(props: Record<string, unknown>) {
  return mountSuspended(PlayerPion, {
    props: { player: alice, team: 'red', x: 50, y: 20, mode: 'view', ...props },
  })
}

describe('PlayerPion', () => {
  it('renders the player initials in the disc', async () => {
    const wrapper = await mount({})

    expect(wrapper.find('.md-pion-disc').text()).toBe('AM')
  })

  it('truncates a long name to eight characters with an ellipsis', async () => {
    const wrapper = await mount({})

    expect(wrapper.find('.md-pion-name').text()).toBe('Alice Ma…')
  })

  it('keeps a short name intact', async () => {
    const wrapper = await mount({ player: bob })

    expect(wrapper.find('.md-pion-name').text()).toBe('Bob')
  })

  it('exposes the player id as a data attribute', async () => {
    const wrapper = await mount({})

    expect(wrapper.find('.md-pion').attributes('data-pion-id')).toBe('p1')
  })

  it('positions the pion from the x and y props', async () => {
    const wrapper = await mount({ x: 42, y: 70 })

    const style = wrapper.find('.md-pion').attributes('style') ?? ''
    expect(style).toContain('left: 42%')
    expect(style).toContain('top: 70%')
  })

  it('applies the green class for the green team', async () => {
    const wrapper = await mount({ team: 'green' })

    expect(wrapper.find('.md-pion').classes()).toContain('md-pion-green')
  })

  it('shows the score rounded to one decimal when provided', async () => {
    const wrapper = await mount({ score: 8 })

    expect(wrapper.find('.md-pion-score').text()).toBe('8.0')
  })

  it('hides the score when none is provided', async () => {
    const wrapper = await mount({})

    expect(wrapper.find('.md-pion-score').exists()).toBe(false)
  })

  it('emits pointerdown with the player in edit mode', async () => {
    const wrapper = await mount({ mode: 'edit' })

    await wrapper.find('.md-pion').trigger('pointerdown')

    expect(wrapper.emitted('pointerdown')).toHaveLength(1)
    expect(wrapper.emitted('pointerdown')![0]![0]).toEqual(alice)
  })

  it('does not emit pointerdown in view mode', async () => {
    const wrapper = await mount({ mode: 'view' })

    await wrapper.find('.md-pion').trigger('pointerdown')

    expect(wrapper.emitted('pointerdown')).toBeUndefined()
  })
})
