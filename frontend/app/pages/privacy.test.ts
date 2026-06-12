// @vitest-environment nuxt
import { describe, it, expect } from 'vitest'
import { mountSuspended } from '@nuxt/test-utils/runtime'
import PrivacyPage from './privacy.vue'

describe('privacy page', () => {
  it('names the data controller and the contact for GDPR requests', async () => {
    const wrapper = await mountSuspended(PrivacyPage)
    expect(wrapper.text()).toContain('Politique de confidentialité')
    expect(wrapper.text()).toContain('Eddin ADOU')
    expect(wrapper.text()).toContain('ianadou807@gmail.com')
    expect(wrapper.text()).toContain('CNIL')
  })

  it('discloses localStorage usage instead of cookies', async () => {
    const wrapper = await mountSuspended(PrivacyPage)
    expect(wrapper.text()).toContain('localStorage')
    expect(wrapper.text()).toContain('aucun cookie')
  })

  it('is honest about vote visibility for the organizer', async () => {
    const wrapper = await mountSuspended(PrivacyPage)
    expect(wrapper.text()).toContain('coéquipiers ne voient jamais qui les a notés')
    expect(wrapper.text()).toContain("L'organisateur du groupe peut consulter le détail des votes")
  })

  it('explains that players are identified by first name only', async () => {
    const wrapper = await mountSuspended(PrivacyPage)
    expect(wrapper.text()).toContain('prénom')
    expect(wrapper.text()).toContain('aucun compte')
  })
})
