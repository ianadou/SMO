// @vitest-environment nuxt
import { describe, it, expect } from 'vitest'
import { mountSuspended } from '@nuxt/test-utils/runtime'
import MatchSetupCard from './MatchSetupCard.vue'

describe('MatchSetupCard', () => {
  it('emits open when the draft action is clicked', async () => {
    const wrapper = await mountSuspended(MatchSetupCard, {
      props: { kind: 'draft' },
    })

    await wrapper.find('.md-setup-btn').trigger('click')

    expect(wrapper.emitted('open')).toHaveLength(1)
  })

  it('emits generate without arguments', async () => {
    const wrapper = await mountSuspended(MatchSetupCard, {
      props: { kind: 'generate' },
    })

    await wrapper.find('.md-setup-btn').trigger('click')
    expect(wrapper.emitted('generate')).toHaveLength(1)
    expect(wrapper.emitted('generate')![0]).toEqual([])
  })

  it('disables the action while busy', async () => {
    const wrapper = await mountSuspended(MatchSetupCard, {
      props: { kind: 'draft', busy: true },
    })

    expect(wrapper.find('.md-setup-btn').attributes('disabled')).toBeDefined()
  })
})
