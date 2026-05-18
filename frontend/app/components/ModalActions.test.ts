// @vitest-environment nuxt
import { describe, it, expect } from 'vitest'
import { mountSuspended } from '@nuxt/test-utils/runtime'
import ModalActions from './ModalActions.vue'

describe('ModalActions', () => {
  it('renders the submit label', async () => {
    const wrapper = await mountSuspended(ModalActions, {
      props: { submitting: false, canSubmit: true, submitLabel: 'Créer le groupe' },
    })
    expect(wrapper.text()).toContain('Créer le groupe')
  })

  it('does not render an error when error is empty', async () => {
    const wrapper = await mountSuspended(ModalActions, {
      props: { submitting: false, canSubmit: true, submitLabel: 'Go' },
    })
    expect(wrapper.find('[role="alert"]').exists()).toBe(false)
  })

  it('renders the error message when error is set', async () => {
    const wrapper = await mountSuspended(ModalActions, {
      props: { submitting: false, canSubmit: true, submitLabel: 'Go', error: 'Boom' },
    })
    expect(wrapper.text()).toContain('Boom')
  })

  it('emits cancel when the Annuler button is clicked', async () => {
    const wrapper = await mountSuspended(ModalActions, {
      props: { submitting: false, canSubmit: true, submitLabel: 'Go' },
    })
    await wrapper.get('button[type="button"]').trigger('click')
    expect(wrapper.emitted('cancel')).toHaveLength(1)
  })

  it('disables the Annuler button while submitting', async () => {
    const wrapper = await mountSuspended(ModalActions, {
      props: { submitting: true, canSubmit: false, submitLabel: 'Go' },
    })
    expect(wrapper.get('button[type="button"]').attributes('disabled')).toBeDefined()
  })

  it('disables the submit button when canSubmit is false', async () => {
    const wrapper = await mountSuspended(ModalActions, {
      props: { submitting: false, canSubmit: false, submitLabel: 'Go' },
    })
    expect(wrapper.get('button[type="submit"]').attributes('disabled')).toBeDefined()
  })
})
