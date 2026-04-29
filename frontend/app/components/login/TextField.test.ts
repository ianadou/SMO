// @vitest-environment nuxt
import { describe, it, expect } from 'vitest'
import { mountSuspended } from '@nuxt/test-utils/runtime'
import TextField from './TextField.vue'

describe('TextField', () => {
  it('renders the label and links it to the input via for/id', async () => {
    const wrapper = await mountSuspended(TextField, {
      props: { id: 'email', label: 'Email', modelValue: '' },
    })
    expect(wrapper.find('label').attributes('for')).toBe('email')
    expect(wrapper.find('input').attributes('id')).toBe('email')
    expect(wrapper.find('label').text()).toBe('Email')
  })

  it('emits update:modelValue when the user types', async () => {
    const wrapper = await mountSuspended(TextField, {
      props: { id: 'email', label: 'Email', modelValue: '' },
    })
    await wrapper.find('input').setValue('alex@gmail.com')
    expect(wrapper.emitted('update:modelValue')?.[0]).toEqual(['alex@gmail.com'])
  })

  it('sets aria-invalid when hasError is true', async () => {
    const wrapper = await mountSuspended(TextField, {
      props: { id: 'email', label: 'Email', modelValue: '', hasError: true },
    })
    expect(wrapper.find('input').attributes('aria-invalid')).toBe('true')
  })

  it('does NOT set aria-invalid when hasError is false', async () => {
    const wrapper = await mountSuspended(TextField, {
      props: { id: 'email', label: 'Email', modelValue: '', hasError: false },
    })
    expect(wrapper.find('input').attributes('aria-invalid')).toBe('false')
  })

  it('forwards type, autocomplete, inputmode, placeholder', async () => {
    const wrapper = await mountSuspended(TextField, {
      props: {
        id: 'email',
        label: 'Email',
        modelValue: '',
        type: 'email',
        autocomplete: 'email',
        inputmode: 'email',
        placeholder: 'toi@exemple.fr',
      },
    })
    const input = wrapper.find('input')
    expect(input.attributes('type')).toBe('email')
    expect(input.attributes('autocomplete')).toBe('email')
    expect(input.attributes('inputmode')).toBe('email')
    expect(input.attributes('placeholder')).toBe('toi@exemple.fr')
  })

  it('renders the right slot content', async () => {
    const wrapper = await mountSuspended(TextField, {
      props: { id: 'email', label: 'Email', modelValue: '' },
      slots: { right: '<button data-testid="slotted">x</button>' },
    })
    expect(wrapper.find('[data-testid="slotted"]').exists()).toBe(true)
  })
})
