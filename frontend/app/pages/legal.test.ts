// @vitest-environment nuxt
import { describe, it, expect } from 'vitest'
import { mountSuspended } from '@nuxt/test-utils/runtime'
import LegalPage from './legal.vue'

describe('legal page', () => {
  it('identifies the publisher and a contact address', async () => {
    const wrapper = await mountSuspended(LegalPage)
    expect(wrapper.text()).toContain('Mentions légales')
    expect(wrapper.text()).toContain('Eddin ADOU')
    expect(wrapper.text()).toContain('ianadou807@gmail.com')
  })

  it('states that the service is not publicly deployed yet', async () => {
    const wrapper = await mountSuspended(LegalPage)
    expect(wrapper.text()).toContain('pas encore déployée')
    expect(wrapper.text()).toContain('OVHcloud')
  })

  it('links back to the other legal documents', async () => {
    const wrapper = await mountSuspended(LegalPage)
    expect(wrapper.html()).toContain('href="/privacy"')
    expect(wrapper.html()).toContain('href="/terms"')
  })
})
