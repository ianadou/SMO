// @vitest-environment nuxt
import { describe, it, expect } from 'vitest'
import { mountSuspended } from '@nuxt/test-utils/runtime'
import { nextTick } from 'vue'
import BaseModal from './BaseModal.vue'

const baseProps = { open: false, title: 'Nouveau groupe' }
const slot = { default: '<p data-testid="body">contenu</p>' }

describe('BaseModal', () => {
  it('renders the title', async () => {
    const wrapper = await mountSuspended(BaseModal, {
      props: { ...baseProps, open: true },
      slots: slot,
    })
    expect(wrapper.find('h2').text()).toBe('Nouveau groupe')
  })

  it('renders the slot content only when open', async () => {
    const wrapper = await mountSuspended(BaseModal, { props: baseProps, slots: slot })
    expect(wrapper.find('[data-testid="body"]').exists()).toBe(false)

    await wrapper.setProps({ open: true })
    expect(wrapper.find('[data-testid="body"]').exists()).toBe(true)
  })

  it('links the dialog to the title via aria-labelledby', async () => {
    const wrapper = await mountSuspended(BaseModal, {
      props: { ...baseProps, open: true },
      slots: slot,
    })
    const labelledby = wrapper.find('dialog').attributes('aria-labelledby')
    expect(labelledby).toBe(wrapper.find('h2').attributes('id'))
  })

  it('opens the native dialog when open becomes true', async () => {
    const wrapper = await mountSuspended(BaseModal, { props: baseProps, slots: slot })
    const el = wrapper.find('dialog').element as HTMLDialogElement
    expect(el.open).toBe(false)

    await wrapper.setProps({ open: true })
    await nextTick()
    expect(el.open).toBe(true)

    await wrapper.setProps({ open: false })
    expect(el.open).toBe(false)
  })

  it('emits close when the close button is clicked', async () => {
    const wrapper = await mountSuspended(BaseModal, {
      props: { ...baseProps, open: true },
      slots: slot,
    })
    await wrapper.get('button[aria-label="Fermer"]').trigger('click')
    expect(wrapper.emitted('close')).toHaveLength(1)
  })

  it('does not emit close from the close button when closeDisabled', async () => {
    const wrapper = await mountSuspended(BaseModal, {
      props: { ...baseProps, open: true, closeDisabled: true },
      slots: slot,
    })
    await wrapper.get('button[aria-label="Fermer"]').trigger('click')
    expect(wrapper.emitted('close')).toBeUndefined()
  })

  it('emits close on the dialog cancel (Escape) event', async () => {
    const wrapper = await mountSuspended(BaseModal, {
      props: { ...baseProps, open: true },
      slots: slot,
    })
    await wrapper.get('dialog').trigger('cancel')
    expect(wrapper.emitted('close')).toHaveLength(1)
  })

  it('emits close on a backdrop click', async () => {
    const wrapper = await mountSuspended(BaseModal, {
      props: { ...baseProps, open: true },
      slots: slot,
    })
    await wrapper.get('dialog').trigger('click')
    expect(wrapper.emitted('close')).toHaveLength(1)
  })

  it('does not emit close on backdrop click when closeDisabled', async () => {
    const wrapper = await mountSuspended(BaseModal, {
      props: { ...baseProps, open: true, closeDisabled: true },
      slots: slot,
    })
    await wrapper.get('dialog').trigger('click')
    expect(wrapper.emitted('close')).toBeUndefined()
  })
})
