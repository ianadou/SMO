// @vitest-environment nuxt
import { describe, it, expect, vi, beforeEach } from 'vitest'
import { mockNuxtImport, mountSuspended } from '@nuxt/test-utils/runtime'
import { nextTick } from 'vue'
import { ApiError } from '~/composables/useApi'
import CreateGroupModal from './CreateGroupModal.vue'

const { postMock } = vi.hoisted(() => ({ postMock: vi.fn() }))
mockNuxtImport('useApi', () => () => ({ post: postMock }))

const open = { props: { open: true } }

beforeEach(() => {
  postMock.mockReset()
})

describe('CreateGroupModal', () => {
  it('renders the name and webhook fields', async () => {
    const wrapper = await mountSuspended(CreateGroupModal, open)
    expect(wrapper.find('input#group-name').exists()).toBe(true)
    expect(wrapper.find('input#group-webhook').exists()).toBe(true)
  })

  it('disables submit when the name is empty', async () => {
    const wrapper = await mountSuspended(CreateGroupModal, open)
    expect(wrapper.get('button[type="submit"]').attributes('disabled')).toBeDefined()
  })

  it('enables submit once a non-empty name is typed', async () => {
    const wrapper = await mountSuspended(CreateGroupModal, open)
    await wrapper.get('input#group-name').setValue('Foot du jeudi')
    expect(wrapper.get('button[type="submit"]').attributes('disabled')).toBeUndefined()
  })

  it('rejects a name longer than 100 characters', async () => {
    const wrapper = await mountSuspended(CreateGroupModal, open)
    await wrapper.get('input#group-name').setValue('A'.repeat(101))
    expect(wrapper.text()).toContain('101/100')
    expect(wrapper.get('button[type="submit"]').attributes('disabled')).toBeDefined()
  })

  it('rejects an invalid Discord webhook URL on blur', async () => {
    const wrapper = await mountSuspended(CreateGroupModal, open)
    await wrapper.get('input#group-name').setValue('Foot')
    const webhook = wrapper.get('input#group-webhook')
    await webhook.setValue('https://malicious.com/x')
    await webhook.trigger('blur')
    expect(wrapper.text()).toContain('Format attendu')
    expect(wrapper.get('button[type="submit"]').attributes('disabled')).toBeDefined()
  })

  it('emits created and resets the form on a successful submit', async () => {
    const created = { id: 'g-1', name: 'Foot', organizer_id: 'o-1', has_webhook: false, created_at: '2026-01-01T00:00:00Z' }
    postMock.mockResolvedValue(created)
    const wrapper = await mountSuspended(CreateGroupModal, open)
    await wrapper.get('input#group-name').setValue('Foot')

    await wrapper.get('form').trigger('submit')
    await nextTick()

    expect(postMock).toHaveBeenCalledWith('/groups', { name: 'Foot' }, expect.objectContaining({ signal: expect.anything() }))
    expect(wrapper.emitted('created')?.[0]).toEqual([created])
    expect((wrapper.get('input#group-name').element as HTMLInputElement).value).toBe('')
  })

  it('shows the backend message when the API rejects with an ApiError', async () => {
    postMock.mockRejectedValue(new ApiError(400, 'Nom déjà pris'))
    const wrapper = await mountSuspended(CreateGroupModal, open)
    await wrapper.get('input#group-name').setValue('Foot')

    await wrapper.get('form').trigger('submit')
    await nextTick()

    expect(wrapper.text()).toContain('Nom déjà pris')
    expect(wrapper.emitted('created')).toBeUndefined()
  })

  it('shows a generic message on an unexpected error', async () => {
    postMock.mockRejectedValue(new Error('boom'))
    const wrapper = await mountSuspended(CreateGroupModal, open)
    await wrapper.get('input#group-name').setValue('Foot')

    await wrapper.get('form').trigger('submit')
    await nextTick()

    expect(wrapper.text()).toContain('Connexion au serveur impossible')
  })

  it('resets the form when the modal is closed', async () => {
    const wrapper = await mountSuspended(CreateGroupModal, open)
    await wrapper.get('input#group-name').setValue('Foot')

    await wrapper.setProps({ open: false })
    await wrapper.setProps({ open: true })

    expect((wrapper.get('input#group-name').element as HTMLInputElement).value).toBe('')
  })
})
