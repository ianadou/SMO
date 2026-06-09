// @vitest-environment nuxt
import { describe, it, expect } from 'vitest'
import { mountSuspended } from '@nuxt/test-utils/runtime'
import MatchValidateBar from './MatchValidateBar.vue'

describe('MatchValidateBar', () => {
  it('renders the validate label', async () => {
    const wrapper = await mountSuspended(MatchValidateBar)

    expect(wrapper.find('.md-bottombar-btn').text()).toBe('Valider les équipes')
  })

  it('emits validate when the button is clicked', async () => {
    const wrapper = await mountSuspended(MatchValidateBar)

    await wrapper.find('.md-bottombar-btn').trigger('click')

    expect(wrapper.emitted('validate')).toHaveLength(1)
  })

  it('disables the button while busy', async () => {
    const wrapper = await mountSuspended(MatchValidateBar, {
      props: { busy: true },
    })

    expect(wrapper.find('.md-bottombar-btn').attributes('disabled')).toBeDefined()
  })
})
