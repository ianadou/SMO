// @vitest-environment nuxt
import { describe, it, expect } from 'vitest'
import { mountSuspended } from '@nuxt/test-utils/runtime'
import JoinClaimModal from './JoinClaimModal.vue'

function mount(over: Record<string, unknown> = {}) {
  return mountSuspended(JoinClaimModal, {
    props: { open: true, busy: false, playerName: 'Marc', ...over },
  })
}

describe('JoinClaimModal', () => {
  it('asks for confirmation with the player name and the permanence note', async () => {
    const wrapper = await mount()

    expect(wrapper.text()).toContain('Vous êtes Marc ?')
    expect(wrapper.text()).toContain('Ce prénom sera définitivement le vôtre pour ce match')
  })

  it('emits confirm from the green CTA', async () => {
    const wrapper = await mount()

    await wrapper.get('[data-testid="claim-confirm"]').trigger('click')

    expect(wrapper.emitted('confirm')).toHaveLength(1)
  })

  it('emits cancel from the secondary action', async () => {
    const wrapper = await mount()

    await wrapper.get('[data-testid="claim-cancel"]').trigger('click')

    expect(wrapper.emitted('cancel')).toHaveLength(1)
  })

  it('disables both actions while busy', async () => {
    const wrapper = await mount({ busy: true })

    expect(wrapper.get('[data-testid="claim-confirm"]').attributes('disabled')).toBeDefined()
    expect(wrapper.get('[data-testid="claim-cancel"]').attributes('disabled')).toBeDefined()
  })
})
