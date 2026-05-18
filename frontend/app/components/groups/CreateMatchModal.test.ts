// @vitest-environment nuxt
import { describe, it, expect, vi, beforeEach } from 'vitest'
import { mockNuxtImport, mountSuspended } from '@nuxt/test-utils/runtime'
import { nextTick } from 'vue'
import { ApiError } from '~/composables/useApi'
import CreateMatchModal from './CreateMatchModal.vue'

const { postMock } = vi.hoisted(() => ({ postMock: vi.fn() }))
mockNuxtImport('useApi', () => () => ({ post: postMock }))

const open = { props: { open: true, groupId: 'g-1' } }
const FUTURE = '2099-01-01T10:00'

beforeEach(() => {
  postMock.mockReset()
})

async function fillValid(wrapper: Awaited<ReturnType<typeof mountSuspended>>) {
  await wrapper.get('input#match-title').setValue('Match du jeudi')
  await wrapper.get('input#match-venue').setValue('Stade municipal')
  const date = wrapper.get('input#match-scheduled')
  await date.setValue(FUTURE)
  await date.trigger('change')
}

describe('CreateMatchModal', () => {
  it('renders the title, venue and date fields', async () => {
    const wrapper = await mountSuspended(CreateMatchModal, open)
    expect(wrapper.find('input#match-title').exists()).toBe(true)
    expect(wrapper.find('input#match-venue').exists()).toBe(true)
    expect(wrapper.find('input#match-scheduled').exists()).toBe(true)
  })

  it('disables submit until title, venue and date are set', async () => {
    const wrapper = await mountSuspended(CreateMatchModal, open)
    expect(wrapper.get('button[type="submit"]').attributes('disabled')).toBeDefined()
  })

  it('enables submit when all fields are valid', async () => {
    const wrapper = await mountSuspended(CreateMatchModal, open)
    await fillValid(wrapper)
    expect(wrapper.get('button[type="submit"]').attributes('disabled')).toBeUndefined()
  })

  it('rejects a date in the past', async () => {
    const wrapper = await mountSuspended(CreateMatchModal, open)
    await wrapper.get('input#match-title').setValue('Match')
    await wrapper.get('input#match-venue').setValue('Stade')
    const date = wrapper.get('input#match-scheduled')
    await date.setValue('2000-01-01T10:00')
    await date.trigger('change')
    expect(wrapper.text()).toContain('La date doit être dans le futur.')
    expect(wrapper.get('button[type="submit"]').attributes('disabled')).toBeDefined()
  })

  it('emits created and resets on a successful submit', async () => {
    const created = { id: 'm-1', group_id: 'g-1', title: 'Match du jeudi', venue: 'Stade municipal', status: 'draft' }
    postMock.mockResolvedValue(created)
    const wrapper = await mountSuspended(CreateMatchModal, open)
    await fillValid(wrapper)

    await wrapper.get('form').trigger('submit')
    await nextTick()

    expect(postMock).toHaveBeenCalledWith(
      '/matches',
      expect.objectContaining({ group_id: 'g-1', title: 'Match du jeudi', venue: 'Stade municipal' }),
      expect.objectContaining({ signal: expect.anything() }),
    )
    expect(wrapper.emitted('created')?.[0]).toEqual([created])
    expect((wrapper.get('input#match-title').element as HTMLInputElement).value).toBe('')
  })

  it('shows the backend message when the API rejects with an ApiError', async () => {
    postMock.mockRejectedValue(new ApiError(409, 'Match déjà existant'))
    const wrapper = await mountSuspended(CreateMatchModal, open)
    await fillValid(wrapper)

    await wrapper.get('form').trigger('submit')
    await nextTick()

    expect(wrapper.text()).toContain('Match déjà existant')
    expect(wrapper.emitted('created')).toBeUndefined()
  })

  it('shows a generic message on an unexpected error', async () => {
    postMock.mockRejectedValue(new Error('boom'))
    const wrapper = await mountSuspended(CreateMatchModal, open)
    await fillValid(wrapper)

    await wrapper.get('form').trigger('submit')
    await nextTick()

    expect(wrapper.text()).toContain('Connexion au serveur impossible')
  })

  it('resets the form when the modal is closed', async () => {
    const wrapper = await mountSuspended(CreateMatchModal, open)
    await wrapper.get('input#match-title').setValue('Match')

    await wrapper.setProps({ open: false })
    await wrapper.setProps({ open: true })

    expect((wrapper.get('input#match-title').element as HTMLInputElement).value).toBe('')
  })
})
