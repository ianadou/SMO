// @vitest-environment nuxt
import { describe, it, expect } from 'vitest'
import { mountSuspended } from '@nuxt/test-utils/runtime'
import AppTopBar from './AppTopBar.vue'

const props = { displayName: 'Eddin Adou', email: 'ianadou807@gmail.com' }

describe('AppTopBar', () => {
  it('links the wordmark to the groups list', async () => {
    const wrapper = await mountSuspended(AppTopBar, { props })
    expect(wrapper.html()).toContain('href="/groups"')
    expect(wrapper.text()).toContain('SMO')
  })

  it('renders the account menu with the organizer initials', async () => {
    const wrapper = await mountSuspended(AppTopBar, { props })
    expect(wrapper.text()).toContain('EA')
  })

  it('forwards the logout event from the account menu', async () => {
    const wrapper = await mountSuspended(AppTopBar, { props })
    await wrapper.find('button').trigger('click')

    await wrapper.get('[data-testid="logout"]').trigger('click')

    expect(wrapper.emitted('logout')).toHaveLength(1)
  })
})
