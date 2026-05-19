// @vitest-environment nuxt
import { describe, it, expect } from 'vitest'
import { mountSuspended } from '@nuxt/test-utils/runtime'
import InviteRespondModal from './InviteRespondModal.vue'

describe('InviteRespondModal', () => {
  it('renders the question title when open', async () => {
    const wrapper = await mountSuspended(InviteRespondModal, {
      props: { open: true, busy: false },
    })
    expect(wrapper.find('h2').text()).toBe('Vous venez à ce match ?')
  })

  it('emits answer "yes" when the Oui button is clicked', async () => {
    const wrapper = await mountSuspended(InviteRespondModal, {
      props: { open: true, busy: false },
    })
    await wrapper.get('[data-testid="answer-yes"]').trigger('click')
    expect(wrapper.emitted('answer')).toEqual([['yes']])
  })

  it('emits answer "no" when the Non button is clicked', async () => {
    const wrapper = await mountSuspended(InviteRespondModal, {
      props: { open: true, busy: false },
    })
    await wrapper.get('[data-testid="answer-no"]').trigger('click')
    expect(wrapper.emitted('answer')).toEqual([['no']])
  })

  it('emits cancel from the Annuler link', async () => {
    const wrapper = await mountSuspended(InviteRespondModal, {
      props: { open: true, busy: false },
    })
    await wrapper.get('[data-testid="answer-cancel"]').trigger('click')
    expect(wrapper.emitted('cancel')).toHaveLength(1)
  })

  it('disables the answer buttons while busy', async () => {
    const wrapper = await mountSuspended(InviteRespondModal, {
      props: { open: true, busy: true },
    })
    expect(wrapper.get('[data-testid="answer-yes"]').attributes('disabled')).toBeDefined()
    expect(wrapper.get('[data-testid="answer-no"]').attributes('disabled')).toBeDefined()
  })
})
