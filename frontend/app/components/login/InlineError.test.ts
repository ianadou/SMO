// @vitest-environment nuxt
import { describe, it, expect } from 'vitest'
import { mountSuspended } from '@nuxt/test-utils/runtime'
import InlineError from './InlineError.vue'

describe('InlineError', () => {
  it('renders the slot content', async () => {
    const wrapper = await mountSuspended(InlineError, {
      slots: { default: 'Identifiants incorrects.' },
    })
    expect(wrapper.text()).toContain('Identifiants incorrects.')
  })

  it('exposes role="alert" so screen readers announce the message', async () => {
    const wrapper = await mountSuspended(InlineError, {
      slots: { default: 'oops' },
    })
    expect(wrapper.attributes('role')).toBe('alert')
  })
})
