// @vitest-environment nuxt
import { describe, it, expect } from 'vitest'
import { mountSuspended } from '@nuxt/test-utils/runtime'
import LegalFooter from './LegalFooter.vue'

describe('LegalFooter', () => {
  it('links to the three legal documents', async () => {
    const wrapper = await mountSuspended(LegalFooter)
    expect(wrapper.html()).toContain('href="/legal"')
    expect(wrapper.html()).toContain('href="/privacy"')
    expect(wrapper.html()).toContain('href="/terms"')
  })

  it('labels the links in french', async () => {
    const wrapper = await mountSuspended(LegalFooter)
    expect(wrapper.text()).toContain('Mentions légales')
    expect(wrapper.text()).toContain('Confidentialité')
    expect(wrapper.text()).toContain('CGU')
  })
})
