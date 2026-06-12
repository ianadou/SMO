// @vitest-environment nuxt
import { describe, it, expect } from 'vitest'
import { mountSuspended } from '@nuxt/test-utils/runtime'
import TermsPage from './terms.vue'

describe('terms page', () => {
  it('presents the service and the applicable law', async () => {
    const wrapper = await mountSuspended(TermsPage)
    expect(wrapper.text()).toContain("Conditions générales d'utilisation")
    expect(wrapper.text()).toContain('droit français')
  })

  it('warns that invitation links are personal credentials', async () => {
    const wrapper = await mountSuspended(TermsPage)
    expect(wrapper.text()).toContain('lien personnel')
    expect(wrapper.text()).toContain('ne le partagez pas')
  })

  it('states the service is provided as is without warranty', async () => {
    const wrapper = await mountSuspended(TermsPage)
    expect(wrapper.text()).toContain('en l\'état')
  })
})
