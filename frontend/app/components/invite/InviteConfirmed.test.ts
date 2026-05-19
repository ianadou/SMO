// @vitest-environment nuxt
import { describe, it, expect } from 'vitest'
import { mountSuspended } from '@nuxt/test-utils/runtime'
import InviteConfirmed from './InviteConfirmed.vue'

describe('InviteConfirmed', () => {
  it('renders the count over the total', async () => {
    const wrapper = await mountSuspended(InviteConfirmed, {
      props: { confirmedCount: 6, maxParticipants: 10, initials: ['IR', 'TB', 'MR'] },
    })
    const text = wrapper.text()
    expect(text).toContain('Déjà confirmés')
    expect(text).toContain('6')
    expect(text).toContain('10')
  })

  it('renders at most six avatars', async () => {
    const wrapper = await mountSuspended(InviteConfirmed, {
      props: {
        confirmedCount: 8,
        maxParticipants: 10,
        initials: ['A', 'B', 'C', 'D', 'E', 'F', 'G', 'H'],
      },
    })
    expect(wrapper.findAll('[data-testid="avatar"]')).toHaveLength(6)
  })

  it('renders nothing when there are no confirmed participants', async () => {
    const wrapper = await mountSuspended(InviteConfirmed, {
      props: { confirmedCount: 0, maxParticipants: 10, initials: [] },
    })
    expect(wrapper.find('[data-testid="confirmed"]').exists()).toBe(false)
  })
})
