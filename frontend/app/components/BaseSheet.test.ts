// @vitest-environment nuxt
import { describe, it, expect } from 'vitest'
import { mountSuspended } from '@nuxt/test-utils/runtime'
import { nextTick } from 'vue'
import BaseSheet from './BaseSheet.vue'

const baseProps = { open: false, title: 'Invitations' }
const slot = { default: '<p data-testid="body">contenu</p>' }

describe('BaseSheet', () => {
  it('renders the title and slot when open', async () => {
    const wrapper = await mountSuspended(BaseSheet, {
      props: { ...baseProps, open: true },
      slots: slot,
    })
    expect(wrapper.find('h2').text()).toBe('Invitations')
    expect(wrapper.find('[data-testid="body"]').exists()).toBe(true)
  })

  it('opens the native dialog when open becomes true', async () => {
    const wrapper = await mountSuspended(BaseSheet, { props: baseProps, slots: slot })
    const el = wrapper.find('dialog').element as HTMLDialogElement
    expect(el.open).toBe(false)

    await wrapper.setProps({ open: true })
    await nextTick()
    expect(el.open).toBe(true)
  })

  it('emits close on cancel', async () => {
    const wrapper = await mountSuspended(BaseSheet, {
      props: { ...baseProps, open: true },
      slots: slot,
    })

    await wrapper.find('dialog').trigger('cancel')

    expect(wrapper.emitted('close')).toHaveLength(1)
  })
})
