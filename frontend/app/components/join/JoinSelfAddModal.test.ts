// @vitest-environment nuxt
import { describe, it, expect } from 'vitest'
import { mountSuspended } from '@nuxt/test-utils/runtime'
import JoinSelfAddModal from './JoinSelfAddModal.vue'

function mount(over: Record<string, unknown> = {}) {
  return mountSuspended(JoinSelfAddModal, {
    props: { open: true, busy: false, ...over },
  })
}

describe('JoinSelfAddModal', () => {
  it('disables the confirm while the first name is empty', async () => {
    const wrapper = await mount()

    expect(wrapper.get('[data-testid="self-add-confirm"]').attributes('disabled')).toBeDefined()
  })

  it('emits confirm with the trimmed first name', async () => {
    const wrapper = await mount()

    await wrapper.get('#self-add-name').setValue('  Nadia ')
    await wrapper.get('form').trigger('submit')

    expect(wrapper.emitted('confirm')).toEqual([['Nadia']])
  })

  it('shows the inline error passed by the parent', async () => {
    const wrapper = await mount({ error: 'Ce prénom est dans la liste — réclamez-le.' })

    expect(wrapper.text()).toContain('Ce prénom est dans la liste — réclamez-le.')
  })

  it('emits cancel from the secondary action', async () => {
    const wrapper = await mount()

    await wrapper.get('[data-testid="self-add-cancel"]').trigger('click')

    expect(wrapper.emitted('cancel')).toHaveLength(1)
  })
})
