// @vitest-environment nuxt
import { describe, it, expect, vi, beforeEach } from 'vitest'
import { mountSuspended, mockNuxtImport } from '@nuxt/test-utils/runtime'
import { flushPromises } from '@vue/test-utils'
import JoinPage from './[token].vue'
import { ApiError } from '~/composables/useApi'
import type { ShareLinkContext } from '~/types/shareLink'

const get = vi.fn()
const post = vi.fn()

mockNuxtImport('useApi', () => () => ({
  get,
  post,
  patch: vi.fn(),
  delete: vi.fn(),
}))

mockNuxtImport('useRoute', () => () => ({ params: { token: 'tok-share' }, query: {} }))

const { navigate } = vi.hoisted(() => ({ navigate: vi.fn() }))
mockNuxtImport('navigateTo', () => navigate)

function shareLinkContext(over: Partial<ShareLinkContext> = {}): ShareLinkContext {
  return {
    match_id: 'm-1',
    organizer_name: 'Alex L.',
    group_name: 'Foot du jeudi',
    match_title: 'Match',
    venue: 'Salle Pierre Mendès, Lyon',
    scheduled_at: '2026-06-18T19:30:00+02:00',
    match_status: 'open',
    capacity: '10 (5v5)',
    confirmed_count: 4,
    max_participants: 10,
    confirmed_initials: ['IR', 'TB'],
    roster: [
      { player_id: 'p-1', player_name: 'Inès', state: 'responded' },
      { player_id: 'p-2', player_name: 'Marc', state: 'claimable' },
      { player_id: 'p-3', player_name: 'Théo', state: 'claimed' },
    ],
    ...over,
  }
}

beforeEach(() => {
  get.mockReset()
  post.mockReset()
  navigate.mockReset()
  window.localStorage.clear()
})

describe('Join page', () => {
  it('shows the roster with the organizer, the group and the section label', async () => {
    get.mockResolvedValueOnce(shareLinkContext())
    const wrapper = await mountSuspended(JoinPage)
    await flushPromises()

    expect(wrapper.text()).toContain('Vous êtes invités')
    expect(wrapper.text()).toContain('Alex L.')
    expect(wrapper.text()).toContain('Foot du jeudi')
    expect(wrapper.text()).toContain('Qui êtes-vous ?')
    expect(wrapper.text()).toContain('Votre lien personnel vous sera remis après identification.')
  })

  it('marks taken names and keeps claimable ones tappable', async () => {
    get.mockResolvedValueOnce(shareLinkContext())
    const wrapper = await mountSuspended(JoinPage)
    await flushPromises()

    expect(wrapper.get('[data-testid="roster-p-1"]').text()).toContain('✓ a répondu')
    expect(wrapper.get('[data-testid="roster-p-3"]').text()).toContain('déjà réclamé')
    expect(wrapper.get('[data-testid="roster-p-2"]').element.tagName).toBe('BUTTON')
    expect(wrapper.get('[data-testid="self-add"]').text()).toContain('Je ne suis pas dans la liste')
  })

  it('shows the inactive screen on a 404', async () => {
    get.mockRejectedValueOnce(new ApiError(404, 'share link not found'))
    const wrapper = await mountSuspended(JoinPage)
    await flushPromises()

    expect(wrapper.text()).toContain('Ce lien n\'est plus actif')
    expect(wrapper.text()).toContain('Demandez le nouveau lien à l\'organisateur')
  })

  it('shows the closed screen once attendance is locked', async () => {
    get.mockResolvedValueOnce(shareLinkContext({ match_status: 'teams_ready' }))
    const wrapper = await mountSuspended(JoinPage)
    await flushPromises()

    expect(wrapper.text()).toContain('Les inscriptions sont closes')
  })

  it('shows the error screen with a retry on a 500', async () => {
    get.mockRejectedValueOnce(new ApiError(500, 'boom'))
    const wrapper = await mountSuspended(JoinPage)
    await flushPromises()

    expect(wrapper.text()).toContain('Une erreur est survenue')
    expect(wrapper.get('button').text()).toContain('Réessayer')
  })

  it('claims a name through the modal, stashes the token and routes to the personal invite', async () => {
    get.mockResolvedValueOnce(shareLinkContext())
    post.mockResolvedValueOnce({ invitation_token: 'tok-perso' })
    const wrapper = await mountSuspended(JoinPage)
    await flushPromises()

    await wrapper.get('[data-testid="roster-p-2"]').trigger('click')
    expect(wrapper.text()).toContain('Vous êtes Marc ?')

    await wrapper.get('[data-testid="claim-confirm"]').trigger('click')
    await flushPromises()

    expect(post).toHaveBeenCalledWith('/share/tok-share/claim', { player_id: 'p-2' })
    expect(window.localStorage.getItem('smo.invitation.m-1')).toBe(
      JSON.stringify({ token: 'tok-perso', playerName: 'Marc' }),
    )
    expect(navigate).toHaveBeenCalledWith('/invite/tok-perso?respond=1')
  })

  it('refetches the roster when the claim loses the race', async () => {
    get.mockResolvedValue(shareLinkContext())
    post.mockRejectedValueOnce(new ApiError(409, 'invitation already claimed'))
    const wrapper = await mountSuspended(JoinPage)
    await flushPromises()

    await wrapper.get('[data-testid="roster-p-2"]').trigger('click')
    await wrapper.get('[data-testid="claim-confirm"]').trigger('click')
    await flushPromises()

    expect(get).toHaveBeenCalledTimes(2)
    expect(navigate).not.toHaveBeenCalled()
  })

  it('joins with a typed first name and routes to the personal invite', async () => {
    get.mockResolvedValueOnce(shareLinkContext())
    post.mockResolvedValueOnce({ invitation_token: 'tok-perso' })
    const wrapper = await mountSuspended(JoinPage)
    await flushPromises()

    await wrapper.get('[data-testid="self-add"]').trigger('click')
    await wrapper.get('#self-add-name').setValue('Nadia')
    await wrapper.get('form').trigger('submit')
    await flushPromises()

    expect(post).toHaveBeenCalledWith('/share/tok-share/join', { player_name: 'Nadia' })
    expect(window.localStorage.getItem('smo.invitation.m-1')).toBe(
      JSON.stringify({ token: 'tok-perso', playerName: 'Nadia' }),
    )
    expect(navigate).toHaveBeenCalledWith('/invite/tok-perso?respond=1')
  })

  it('shows the inline claim hint when the typed name is already invited', async () => {
    get.mockResolvedValueOnce(shareLinkContext())
    post.mockRejectedValueOnce(new ApiError(409, 'player already invited to this match'))
    const wrapper = await mountSuspended(JoinPage)
    await flushPromises()

    await wrapper.get('[data-testid="self-add"]').trigger('click')
    await wrapper.get('#self-add-name').setValue('Marc')
    await wrapper.get('form').trigger('submit')
    await flushPromises()

    expect(wrapper.text()).toContain('Ce prénom est dans la liste — réclamez-le.')
    expect(navigate).not.toHaveBeenCalled()
  })

  it('offers to resume when a personal token is stashed for this match', async () => {
    window.localStorage.setItem(
      'smo.invitation.m-1',
      JSON.stringify({ token: 'tok-perso', playerName: 'Marc' }),
    )
    get.mockResolvedValueOnce(shareLinkContext())
    const wrapper = await mountSuspended(JoinPage)
    await flushPromises()

    const banner = wrapper.get('[data-testid="resume-banner"]')
    expect(banner.text()).toContain('Vous êtes déjà')
    expect(banner.text()).toContain('Marc')

    await banner.trigger('click')

    expect(navigate).toHaveBeenCalledWith('/invite/tok-perso')
  })
})
