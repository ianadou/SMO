// @vitest-environment nuxt
import { describe, it, expect } from 'vitest'
import { mountSuspended } from '@nuxt/test-utils/runtime'
import CguCheckbox from './CguCheckbox.vue'

describe('CguCheckbox', () => {
  it('reflects the modelValue on the underlying checkbox', async () => {
    const wrapper = await mountSuspended(CguCheckbox, {
      props: { modelValue: true },
    })
    expect((wrapper.find('input[type="checkbox"]').element as HTMLInputElement).checked).toBe(true)
  })

  it('emits update:modelValue when toggled', async () => {
    const wrapper = await mountSuspended(CguCheckbox, {
      props: { modelValue: false },
    })
    await wrapper.find('input[type="checkbox"]').setValue(true)
    expect(wrapper.emitted('update:modelValue')?.[0]).toEqual([true])
  })

  it('exposes the CGU and privacy links pointing at hash anchors', async () => {
    const wrapper = await mountSuspended(CguCheckbox, {
      props: { modelValue: false },
    })
    const links = wrapper.findAll('a')
    expect(links).toHaveLength(2)
    expect(links[0]!.text()).toBe('conditions générales')
    expect(links[1]!.text()).toBe('politique de confidentialité')
  })

  it('marks the input as aria-invalid when hasError is true', async () => {
    const wrapper = await mountSuspended(CguCheckbox, {
      props: { modelValue: false, hasError: true },
    })
    expect(wrapper.find('input[type="checkbox"]').attributes('aria-invalid')).toBe('true')
  })
})
