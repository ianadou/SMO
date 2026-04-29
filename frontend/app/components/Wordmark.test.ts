// @vitest-environment nuxt
import { describe, it, expect } from 'vitest'
import { mountSuspended } from '@nuxt/test-utils/runtime'
import Wordmark from './Wordmark.vue'

describe('Wordmark', () => {
  it('renders the SMO text and a dot', async () => {
    const wrapper = await mountSuspended(Wordmark)
    expect(wrapper.text()).toContain('SMO')
    expect(wrapper.attributes('aria-label')).toBe('SMO')
  })

  it('scales the text with the size prop', async () => {
    const wrapper = await mountSuspended(Wordmark, { props: { size: 56 } })
    const text = wrapper.find('span > span:first-child')
    expect(text.attributes('style')).toContain('font-size: 56px')
  })

  it('uses the default size of 40 when none is passed', async () => {
    const wrapper = await mountSuspended(Wordmark)
    const text = wrapper.find('span > span:first-child')
    expect(text.attributes('style')).toContain('font-size: 40px')
  })
})
