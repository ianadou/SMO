// @vitest-environment nuxt
import { describe, it, expect, vi, beforeEach } from 'vitest'
import { mountSuspended, mockNuxtImport } from '@nuxt/test-utils/runtime'
import { useAuthStore } from '~/stores/auth'
import OrganizerLayout from './organizer.vue'

const { navigate } = vi.hoisted(() => ({ navigate: vi.fn() }))
mockNuxtImport('navigateTo', () => navigate)

const ORGANIZER = {
  id: 'o1',
  email: 'ianadou807@gmail.com',
  display_name: 'Eddin Adou',
  created_at: '2026-06-12T00:00:00Z',
}

describe('organizer layout', () => {
  beforeEach(() => {
    navigate.mockReset()
    useAuthStore().setSession('jwt-abc', ORGANIZER)
  })

  it('renders the top bar with the organizer identity and the page content', async () => {
    const wrapper = await mountSuspended(OrganizerLayout, {
      slots: { default: () => 'contenu de la page' },
    })

    expect(wrapper.text()).toContain('EA')
    expect(wrapper.text()).toContain('contenu de la page')
  })

  it('renders the legal footer links', async () => {
    const wrapper = await mountSuspended(OrganizerLayout)

    expect(wrapper.html()).toContain('href="/legal"')
    expect(wrapper.html()).toContain('href="/privacy"')
    expect(wrapper.html()).toContain('href="/terms"')
  })

  it('clears the session and redirects to login on logout', async () => {
    const wrapper = await mountSuspended(OrganizerLayout)

    await wrapper.get('[aria-label="Menu du compte"]').trigger('click')
    await wrapper.get('[data-testid="logout"]').trigger('click')

    expect(useAuthStore().token).toBeNull()
    expect(navigate).toHaveBeenCalledWith('/login', { replace: true })
  })
})
