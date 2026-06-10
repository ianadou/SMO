// @vitest-environment nuxt
import { describe, it, expect } from 'vitest'
import { mountSuspended } from '@nuxt/test-utils/runtime'
import InviteSheet from './InviteSheet.vue'
import type { InviteRow } from '~/types/invitation'

const rows: InviteRow[] = [
  { playerId: 'p-1', playerName: 'Inès R.', status: 'yes', shareUrl: null },
  { playerId: 'p-2', playerName: 'Théo B.', status: 'pending', shareUrl: null },
  { playerId: 'p-3', playerName: 'Marc R.', status: 'not-invited', shareUrl: null },
  { playerId: 'p-4', playerName: 'Paul S.', status: 'fresh', shareUrl: 'http://x/invite/tok' },
]

function mount(over: Record<string, unknown> = {}) {
  return mountSuspended(InviteSheet, {
    props: { open: true, rows, confirmedCount: 1, ...over },
  })
}

describe('InviteSheet', () => {
  it('renders one row per player with its status', async () => {
    const wrapper = await mount()

    expect(wrapper.get('[data-testid="invite-row-p-1"]').text()).toContain('✓ Vient')
    expect(wrapper.get('[data-testid="invite-row-p-2"]').text()).toContain('En attente')
    expect(wrapper.get('[data-testid="invite-p-3"]').text()).toContain('Inviter')
    expect(wrapper.get('[data-testid="share-p-4"]').text()).toContain('Partager le lien')
  })

  it('emits invite with the player id', async () => {
    const wrapper = await mount()

    await wrapper.get('[data-testid="invite-p-3"]').trigger('click')

    expect(wrapper.emitted('invite')).toEqual([['p-3']])
  })

  it('emits share with the fresh row', async () => {
    const wrapper = await mount()

    await wrapper.get('[data-testid="share-p-4"]').trigger('click')

    expect(wrapper.emitted('share')?.[0]?.[0]).toMatchObject({ playerId: 'p-4' })
  })

  it('hides the invite action and explains the lock when locked', async () => {
    const wrapper = await mount({ locked: true })

    expect(wrapper.find('[data-testid="invite-p-3"]').exists()).toBe(false)
    expect(wrapper.get('[data-testid="invite-row-p-3"]').text()).toContain('Non invité')
    expect(wrapper.text()).toContain('réponses sont verrouillées')
  })

  it('shows the retry action when loading failed', async () => {
    const wrapper = await mount({ failed: true, rows: [] })

    expect(wrapper.text()).toContain('Impossible de charger')

    await wrapper.get('button').trigger('click')

    expect(wrapper.emitted('retry')).toHaveLength(1)
  })
})
