// @vitest-environment nuxt
import { afterEach, beforeEach, describe, expect, it } from 'vitest'
import { mountSuspended } from '@nuxt/test-utils/runtime'
import ToastStack from './ToastStack.vue'
import { useToast } from '~/composables/useToast'

describe('ToastStack', () => {
  beforeEach(() => {
    const { toasts } = useToast()
    toasts.value = []
  })

  afterEach(() => {
    const { toasts } = useToast()
    toasts.value = []
  })

  it('renders nothing visible when there are no toasts', async () => {
    const wrapper = await mountSuspended(ToastStack)
    expect(wrapper.findAll('.toast')).toHaveLength(0)
  })

  it('exposes a region with the Notifications label', async () => {
    const wrapper = await mountSuspended(ToastStack)
    const section = wrapper.find('section.toast-stack')
    expect(section.exists()).toBe(true)
    expect(section.attributes('aria-label')).toBe('Notifications')
    expect(section.attributes('aria-live')).toBe('polite')
  })

  it('renders a toast for each pushed item with the matching kind class', async () => {
    const t = useToast()
    t.success('Saved', 'ok', 0)
    t.error('Boom', 'kaboom', 0)
    const wrapper = await mountSuspended(ToastStack)
    const toasts = wrapper.findAll('.toast')
    expect(toasts).toHaveLength(2)
    expect(toasts[0]!.classes()).toContain('is-success')
    expect(toasts[1]!.classes()).toContain('is-error')
    expect(wrapper.text()).toContain('Saved')
    expect(wrapper.text()).toContain('kaboom')
  })

  it('dismisses a toast when the close button is clicked', async () => {
    const t = useToast()
    t.info('Hello', undefined, 0)
    const wrapper = await mountSuspended(ToastStack)
    expect(wrapper.findAll('.toast')).toHaveLength(1)
    await wrapper.find('.toast-close').trigger('click')
    expect(t.toasts.value).toHaveLength(0)
  })
})
