// @vitest-environment nuxt
import { describe, it, expect, vi, beforeEach } from 'vitest'
import { mountSuspended, mockNuxtImport } from '@nuxt/test-utils/runtime'
import { flushPromises } from '@vue/test-utils'
import RenameGroupModal from './RenameGroupModal.vue'
import { ApiError } from '~/composables/useApi'
import type { GroupDTO } from '~/types/groups'

const patch = vi.fn()

mockNuxtImport('useApi', () => () => ({
  get: vi.fn(),
  post: vi.fn(),
  patch,
  delete: vi.fn(),
}))

const group: GroupDTO = {
  id: 'g-1',
  name: 'Foot du jeudi',
  organizer_id: 'org-1',
  has_webhook: false,
  created_at: '2026-04-01T00:00:00Z',
}

beforeEach(() => {
  patch.mockReset()
})

describe('RenameGroupModal', () => {
  it('prefills the current name and disables submit until it changes', async () => {
    const wrapper = await mountSuspended(RenameGroupModal, {
      props: { open: true, group },
    })

    const input = wrapper.get('input').element as HTMLInputElement
    expect(input.value).toBe('Foot du jeudi')
    expect(wrapper.get('button[type="submit"]').attributes('disabled')).toBeDefined()
  })

  it('patches the group and emits renamed', async () => {
    patch.mockResolvedValueOnce({ ...group, name: 'Foot du vendredi' })
    const wrapper = await mountSuspended(RenameGroupModal, {
      props: { open: true, group },
    })

    await wrapper.get('input').setValue('Foot du vendredi')
    await wrapper.get('form').trigger('submit')
    await flushPromises()

    expect(patch).toHaveBeenCalledWith('/groups/g-1', { name: 'Foot du vendredi' })
    expect(wrapper.emitted('renamed')?.[0]?.[0]).toMatchObject({ name: 'Foot du vendredi' })
  })

  it('surfaces the api error message on failure', async () => {
    patch.mockRejectedValueOnce(new ApiError(400, 'invalid name'))
    const wrapper = await mountSuspended(RenameGroupModal, {
      props: { open: true, group },
    })

    await wrapper.get('input').setValue('Nouveau nom')
    await wrapper.get('form').trigger('submit')
    await flushPromises()

    expect(wrapper.text()).toContain('invalid name')
    expect(wrapper.emitted('renamed')).toBeUndefined()
  })
})
