// @vitest-environment nuxt
import { describe, it, expect } from 'vitest'
import { mountSuspended } from '@nuxt/test-utils/runtime'
import StarRow from './StarRow.vue'

describe('StarRow', () => {
  it('renders five radio buttons with the rated ones checked', async () => {
    const wrapper = await mountSuspended(StarRow, { props: { modelValue: 3 } })

    const stars = wrapper.findAll('[role="radio"]')
    expect(stars).toHaveLength(5)
    expect(stars[2]?.attributes('aria-checked')).toBe('true')
    expect(stars[3]?.attributes('aria-checked')).toBe('false')
  })

  it('emits the tapped value', async () => {
    const wrapper = await mountSuspended(StarRow, { props: { modelValue: 0 } })

    await wrapper.get('[data-testid="star-4"]').trigger('click')

    expect(wrapper.emitted('update:modelValue')).toEqual([[4]])
  })

  it('toggles off when tapping the current value', async () => {
    const wrapper = await mountSuspended(StarRow, { props: { modelValue: 4 } })

    await wrapper.get('[data-testid="star-4"]').trigger('click')

    expect(wrapper.emitted('update:modelValue')).toEqual([[0]])
  })

  it('ignores taps when locked', async () => {
    const wrapper = await mountSuspended(StarRow, {
      props: { modelValue: 3, locked: true },
    })

    await wrapper.get('[data-testid="star-5"]').trigger('click')

    expect(wrapper.emitted('update:modelValue')).toBeUndefined()
  })
})
