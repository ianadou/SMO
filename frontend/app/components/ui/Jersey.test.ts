// @vitest-environment nuxt
import { describe, it, expect } from 'vitest'
import { mountSuspended } from '@nuxt/test-utils/runtime'
import Jersey from './Jersey.vue'

describe('Jersey', () => {
  it('renders an SVG with the red class by default', async () => {
    const wrapper = await mountSuspended(Jersey)
    const svg = wrapper.find('svg')
    expect(svg.exists()).toBe(true)
    expect(svg.classes()).toContain('is-red')
  })

  it('applies the green color class when color="green"', async () => {
    const wrapper = await mountSuspended(Jersey, { props: { color: 'green' } })
    expect(wrapper.find('svg').classes()).toContain('is-green')
  })

  it('honors the size prop on width and height', async () => {
    const wrapper = await mountSuspended(Jersey, { props: { size: 36 } })
    const svg = wrapper.find('svg')
    expect(svg.attributes('width')).toBe('36')
    expect(svg.attributes('height')).toBe('36')
  })

  it('exposes an accessible role and label when ariaLabel is set', async () => {
    const wrapper = await mountSuspended(Jersey, { props: { ariaLabel: 'Équipe rouge' } })
    const svg = wrapper.find('svg')
    expect(svg.attributes('role')).toBe('img')
    expect(svg.attributes('aria-label')).toBe('Équipe rouge')
  })

  it('hides the SVG from assistive tech when no ariaLabel is provided', async () => {
    const wrapper = await mountSuspended(Jersey)
    const svg = wrapper.find('svg')
    expect(svg.attributes('role')).toBe('presentation')
    expect(svg.attributes('aria-hidden')).toBe('true')
  })
})
