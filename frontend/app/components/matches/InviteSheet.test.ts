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

    await wrapper.get('[data-testid="retry-invitations"]').trigger('click')

    expect(wrapper.emitted('retry')).toHaveLength(1)
  })

  it('offers to generate the match link when none is held', async () => {
    const wrapper = await mount()

    await wrapper.get('[data-testid="generate-link"]').trigger('click')

    expect(wrapper.emitted('generate-link')).toHaveLength(1)
  })

  it('shows the link with its expiry and emits copy-link', async () => {
    const wrapper = await mount({
      shareLink: { token: 'tok-share', url: 'http://x/join/tok-share', expires_at: '2026-06-17T19:00:00Z' },
    })

    expect(wrapper.get('[data-testid="share-link-url"]').text()).toBe('http://x/join/tok-share')
    expect(wrapper.text()).toContain('Expire le 17 juin 2026')

    await wrapper.get('[data-testid="copy-link"]').trigger('click')

    expect(wrapper.emitted('copy-link')).toHaveLength(1)
  })

  it('emits generate-link only after the regenerate confirm', async () => {
    const wrapper = await mount({
      shareLink: { token: 'tok-share', url: 'http://x/join/tok-share', expires_at: '2026-06-17T19:00:00Z' },
    })

    await wrapper.get('[data-testid="regenerate-link"]').trigger('click')
    expect(wrapper.emitted('generate-link')).toBeUndefined()
    expect(wrapper.text()).toContain('L\'ancien lien du match cessera de fonctionner')

    await wrapper.get('[data-testid="confirm-regenerate"]').trigger('click')

    expect(wrapper.emitted('generate-link')).toHaveLength(1)
  })

  it('emits revoke-link only after the revoke confirm', async () => {
    const wrapper = await mount({
      shareLink: { token: 'tok-share', url: 'http://x/join/tok-share', expires_at: '2026-06-17T19:00:00Z' },
    })

    await wrapper.get('[data-testid="revoke-link"]').trigger('click')
    expect(wrapper.emitted('revoke-link')).toBeUndefined()

    await wrapper.get('[data-testid="confirm-revoke"]').trigger('click')

    expect(wrapper.emitted('revoke-link')).toHaveLength(1)
  })

  it('shows the claimed badge only for rows reported as claimed', async () => {
    const claimedRows: InviteRow[] = [
      { playerId: 'p-1', playerName: 'Inès R.', status: 'yes', shareUrl: null, claimed: true },
      { playerId: 'p-2', playerName: 'Théo B.', status: 'pending', shareUrl: null },
    ]
    const wrapper = await mount({ rows: claimedRows })

    expect(wrapper.get('[data-testid="claimed-p-1"]').text()).toContain('réclamé ✓')
    expect(wrapper.find('[data-testid="claimed-p-2"]').exists()).toBe(false)
  })
})
