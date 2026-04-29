// @vitest-environment nuxt
import { describe, it, expect } from 'vitest'
import { mountSuspended } from '@nuxt/test-utils/runtime'
import PrimaryButton from './PrimaryButton.vue'

describe('PrimaryButton', () => {
  it('renders the slot content in the default state', async () => {
    const wrapper = await mountSuspended(PrimaryButton, {
      slots: { default: 'Se connecter' },
    })
    expect(wrapper.text()).toBe('Se connecter')
  })

  it('replaces the slot with the spinner and loading label when loading is true', async () => {
    const wrapper = await mountSuspended(PrimaryButton, {
      props: { loading: true },
      slots: { default: 'Se connecter' },
    })
    expect(wrapper.text()).toContain('Connexion…')
    expect(wrapper.text()).not.toContain('Se connecter')
  })

  it('uses the loadingLabel prop when provided', async () => {
    const wrapper = await mountSuspended(PrimaryButton, {
      props: { loading: true, loadingLabel: 'Inscription…' },
      slots: { default: 'S\'inscrire' },
    })
    expect(wrapper.text()).toContain('Inscription…')
  })

  it('disables the button when loading is true', async () => {
    const wrapper = await mountSuspended(PrimaryButton, {
      props: { loading: true },
    })
    expect(wrapper.find('button').attributes('disabled')).toBeDefined()
  })

  it('disables the button when disabled is true even without loading', async () => {
    const wrapper = await mountSuspended(PrimaryButton, {
      props: { disabled: true },
    })
    expect(wrapper.find('button').attributes('disabled')).toBeDefined()
  })

  it('defaults to type="submit"', async () => {
    const wrapper = await mountSuspended(PrimaryButton)
    expect(wrapper.find('button').attributes('type')).toBe('submit')
  })
})
