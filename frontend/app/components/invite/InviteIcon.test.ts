// @vitest-environment nuxt
import { describe, it, expect } from 'vitest'
import { mountSuspended } from '@nuxt/test-utils/runtime'
import InviteIcon from './InviteIcon.vue'

describe('InviteIcon', () => {
  it('renders an svg sized by the size prop', async () => {
    const wrapper = await mountSuspended(InviteIcon, { props: { name: 'clock', size: 20 } })
    const svg = wrapper.find('svg')
    expect(svg.attributes('width')).toBe('20')
    expect(svg.attributes('height')).toBe('20')
  })

  it('renders the filled check hero icon with fill and no stroke', async () => {
    const wrapper = await mountSuspended(InviteIcon, { props: { name: 'check-circle-filled', size: 64 } })
    const svg = wrapper.find('svg')
    expect(svg.attributes('fill')).toBe('currentColor')
    expect(svg.attributes('stroke')).toBeUndefined()
  })

  it('renders the unlink icon used by the invalid state', async () => {
    const wrapper = await mountSuspended(InviteIcon, { props: { name: 'unlink', size: 60 } })
    expect(wrapper.find('svg').exists()).toBe(true)
  })
})
