// @vitest-environment nuxt
import { describe, it, expect } from 'vitest'
import { mountSuspended } from '@nuxt/test-utils/runtime'
import InviteMinimalState from './InviteMinimalState.vue'

const base = {
  icon: 'clock' as const,
  iconTone: 'warn' as const,
  title: 'Invitation expirée',
  subtitle: 'La date limite de réponse est passée.',
  legal: 'Vous pouvez fermer cette page.',
}

describe('InviteMinimalState', () => {
  it('renders the title, subtitle and legal', async () => {
    const wrapper = await mountSuspended(InviteMinimalState, { props: base })
    expect(wrapper.find('h1').text()).toBe('Invitation expirée')
    expect(wrapper.text()).toContain('La date limite de réponse est passée.')
    expect(wrapper.text()).toContain('Vous pouvez fermer cette page.')
  })

  it('applies the warn tone class to the icon wrapper', async () => {
    const wrapper = await mountSuspended(InviteMinimalState, { props: base })
    expect(wrapper.find('[data-testid="invite-icon"]').classes()).toContain('text-warn')
  })

  it('emits retry when the optional action button is present and clicked', async () => {
    const wrapper = await mountSuspended(InviteMinimalState, {
      props: { ...base, actionLabel: 'Réessayer' },
    })
    await wrapper.get('button').trigger('click')
    expect(wrapper.emitted('action')).toHaveLength(1)
  })

  it('renders no action button when actionLabel is absent', async () => {
    const wrapper = await mountSuspended(InviteMinimalState, { props: base })
    expect(wrapper.find('button').exists()).toBe(false)
  })
})
