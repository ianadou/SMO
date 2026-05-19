// @vitest-environment nuxt
import { describe, it, expect, vi, beforeEach } from 'vitest'
import { mountSuspended, mockNuxtImport } from '@nuxt/test-utils/runtime'
import { flushPromises } from '@vue/test-utils'
import InvitePage from './[token].vue'
import { ApiError } from '~/composables/useApi'

const post = vi.fn()

mockNuxtImport('useApi', () => () => ({
  get: vi.fn(),
  post,
  patch: vi.fn(),
  delete: vi.fn(),
}))

mockNuxtImport('useRoute', () => () => ({ params: { token: 'tok-123' } }))

const fullContext = {
  organizer_name: 'Alex L.',
  group_name: 'Foot du jeudi',
  match_title: 'Match',
  venue: 'Salle Pierre Mendès, Lyon',
  scheduled_at: '2026-05-07T19:30:00+02:00',
  capacity: '10 (5v5)',
  confirmed_count: 6,
  max_participants: 10,
  confirmed_initials: ['IR', 'TB'],
  response: 'pending',
  expires_at: '2026-05-07T18:00:00+02:00',
  state: 'respondable',
}

beforeEach(() => {
  post.mockReset()
})

describe('Invitation page', () => {
  it('shows the initial view with organizer, group and the Répondre CTA', async () => {
    post.mockResolvedValueOnce(fullContext)
    const wrapper = await mountSuspended(InvitePage)
    await flushPromises()
    expect(wrapper.text()).toContain('Vous êtes invité')
    expect(wrapper.text()).toContain('Alex L.')
    expect(wrapper.text()).toContain('Foot du jeudi')
    expect(wrapper.get('[data-testid="respond-cta"]').text()).toContain('Répondre')
  })

  it('posts the token from the route to the context endpoint', async () => {
    post.mockResolvedValueOnce(fullContext)
    await mountSuspended(InvitePage)
    await flushPromises()
    expect(post).toHaveBeenCalledWith('/invitations/context', { token: 'tok-123' })
  })

  it('shows the invalid screen on a 404', async () => {
    post.mockRejectedValueOnce(new ApiError(404, 'invitation not found'))
    const wrapper = await mountSuspended(InvitePage)
    await flushPromises()
    expect(wrapper.text()).toContain('Lien invalide')
  })

  it('shows the expired screen when state is expired', async () => {
    post.mockResolvedValueOnce({ ...fullContext, state: 'expired' })
    const wrapper = await mountSuspended(InvitePage)
    await flushPromises()
    expect(wrapper.text()).toContain('Invitation expirée')
  })

  it('shows the error screen with a retry on a 500', async () => {
    post.mockRejectedValueOnce(new ApiError(500, 'boom'))
    const wrapper = await mountSuspended(InvitePage)
    await flushPromises()
    expect(wrapper.text()).toContain('Une erreur est survenue')
    expect(wrapper.get('button').text()).toContain('Réessayer')
  })

  it('answers yes through the modal and transitions to the result view', async () => {
    post.mockResolvedValueOnce(fullContext)
    post.mockResolvedValueOnce({ response: 'yes', responded_at: '2026-05-01T10:00:00Z' })
    const wrapper = await mountSuspended(InvitePage)
    await flushPromises()

    await wrapper.get('[data-testid="respond-cta"]').trigger('click')
    await wrapper.get('[data-testid="answer-yes"]').trigger('click')
    await flushPromises()

    expect(post).toHaveBeenLastCalledWith('/invitations/respond', { token: 'tok-123', answer: 'yes' })
    expect(wrapper.text()).toContain('Vous êtes inscrit')
  })

  it('falls back to the expired screen when respond returns 410', async () => {
    post.mockResolvedValueOnce(fullContext)
    post.mockRejectedValueOnce(new ApiError(410, 'invitation expired'))
    const wrapper = await mountSuspended(InvitePage)
    await flushPromises()

    await wrapper.get('[data-testid="respond-cta"]').trigger('click')
    await wrapper.get('[data-testid="answer-no"]').trigger('click')
    await flushPromises()

    expect(wrapper.text()).toContain('Invitation expirée')
  })

  it('reopens the modal from the result view and changes the answer', async () => {
    post.mockResolvedValueOnce(fullContext)
    post.mockResolvedValueOnce({ response: 'yes', responded_at: '2026-05-01T10:00:00Z' })
    post.mockResolvedValueOnce({ response: 'no', responded_at: '2026-05-01T10:05:00Z' })
    const wrapper = await mountSuspended(InvitePage)
    await flushPromises()

    await wrapper.get('[data-testid="respond-cta"]').trigger('click')
    await wrapper.get('[data-testid="answer-yes"]').trigger('click')
    await flushPromises()
    expect(wrapper.text()).toContain('Vous êtes inscrit')

    await wrapper.get('[data-testid="modify-cta"]').trigger('click')
    await wrapper.get('[data-testid="answer-no"]').trigger('click')
    await flushPromises()

    expect(post).toHaveBeenLastCalledWith('/invitations/respond', { token: 'tok-123', answer: 'no' })
    expect(wrapper.text()).toContain('Réponse enregistrée')
  })

  it('refetches context and shows the locked-pending screen when respond returns 409 and the player never answered', async () => {
    post.mockResolvedValueOnce(fullContext)
    post.mockRejectedValueOnce(new ApiError(409, 'match locked'))
    post.mockResolvedValueOnce({ ...fullContext, state: 'locked', response: 'pending' })
    const wrapper = await mountSuspended(InvitePage)
    await flushPromises()

    await wrapper.get('[data-testid="respond-cta"]').trigger('click')
    await wrapper.get('[data-testid="answer-yes"]').trigger('click')
    await flushPromises()

    expect(wrapper.text()).toContain('Les équipes ont été formées')
  })
})
