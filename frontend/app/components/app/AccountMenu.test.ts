// @vitest-environment nuxt
import { describe, it, expect } from 'vitest'
import { mountSuspended } from '@nuxt/test-utils/runtime'
import AccountMenu from './AccountMenu.vue'

const organizer = { displayName: 'Eddin Adou', email: 'ianadou807@gmail.com' }

describe('AccountMenu', () => {
  it('shows the organizer initials on the trigger', async () => {
    const wrapper = await mountSuspended(AccountMenu, { props: organizer })
    expect(wrapper.text()).toContain('EA')
    expect(wrapper.find('button').attributes('aria-expanded')).toBe('false')
  })

  it('opens the menu with name, email and logout action', async () => {
    const wrapper = await mountSuspended(AccountMenu, { props: organizer })

    await wrapper.find('button').trigger('click')

    expect(wrapper.find('button').attributes('aria-expanded')).toBe('true')
    expect(wrapper.text()).toContain('Eddin Adou')
    expect(wrapper.text()).toContain('ianadou807@gmail.com')
    expect(wrapper.text()).toContain('Déconnexion')
  })

  it('emits logout when the action is clicked', async () => {
    const wrapper = await mountSuspended(AccountMenu, { props: organizer })
    await wrapper.find('button').trigger('click')

    await wrapper.get('[data-testid="logout"]').trigger('click')

    expect(wrapper.emitted('logout')).toHaveLength(1)
  })

  it('closes the menu on escape', async () => {
    const wrapper = await mountSuspended(AccountMenu, { props: organizer })
    await wrapper.find('button').trigger('click')

    document.dispatchEvent(new KeyboardEvent('keydown', { key: 'Escape' }))
    await wrapper.vm.$nextTick()

    expect(wrapper.find('button').attributes('aria-expanded')).toBe('false')
    expect(wrapper.text()).not.toContain('Déconnexion')
  })
})
