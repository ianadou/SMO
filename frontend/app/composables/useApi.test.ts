// @vitest-environment nuxt
import { describe, it, expect, vi, beforeEach } from 'vitest'
import { setActivePinia, createPinia } from 'pinia'
import { useApi, ApiError } from './useApi'
import { useAuthStore } from '~/stores/auth'

const fetchMock = vi.fn()

const ORGANIZER = {
  id: 'o1',
  email: 'alex@example.fr',
  display_name: 'Alex',
  created_at: '2026-01-01T00:00:00Z',
}

beforeEach(() => {
  window.localStorage.clear()
  setActivePinia(createPinia())
  fetchMock.mockReset()
  fetchMock.mockResolvedValue({ ok: true })
  vi.stubGlobal('$fetch', fetchMock)
})

function lastCall() {
  return fetchMock.mock.calls.at(-1)!
}

describe('useApi', () => {
  it('issues a GET with the json accept header', async () => {
    await useApi().get('/groups')

    const [path, opts] = lastCall()
    expect(path).toBe('/groups')
    expect(opts.method).toBe('GET')
    expect(opts.headers.Accept).toBe('application/json')
  })

  it('sends the body for POST, PUT and PATCH', async () => {
    const body = { name: 'Foot' }

    await useApi().post('/groups', body)
    expect(lastCall()[1]).toMatchObject({ method: 'POST', body })

    await useApi().put('/matches/1/teams', body)
    expect(lastCall()[1]).toMatchObject({ method: 'PUT', body })

    await useApi().patch('/players/1/ranking', body)
    expect(lastCall()[1]).toMatchObject({ method: 'PATCH', body })
  })

  it('issues a DELETE without a body', async () => {
    await useApi().delete('/groups/1')

    expect(lastCall()[1].method).toBe('DELETE')
    expect(lastCall()[1].body).toBeUndefined()
  })

  it('omits the Authorization header when unauthenticated', async () => {
    await useApi().get('/groups')

    expect(lastCall()[1].headers.Authorization).toBeUndefined()
  })

  it('adds a Bearer token when authenticated', async () => {
    useAuthStore().setSession('jwt-abc', ORGANIZER)

    await useApi().get('/groups')

    expect(lastCall()[1].headers.Authorization).toBe('Bearer jwt-abc')
  })

  it('maps a response error to an ApiError carrying the server message', async () => {
    await useApi().get('/groups')
    const { onResponseError } = lastCall()[1]

    try {
      onResponseError({ response: { status: 409, _data: { error: 'already exists' } } })
      expect.unreachable('onResponseError should throw')
    } catch (e) {
      expect(e).toBeInstanceOf(ApiError)
      expect((e as ApiError).status).toBe(409)
      expect((e as ApiError).publicMessage).toBe('already exists')
    }
  })

  it('falls back to a generic message when the error body has none', async () => {
    await useApi().get('/groups')
    const { onResponseError } = lastCall()[1]

    try {
      onResponseError({ response: { status: 500, _data: undefined } })
      expect.unreachable('onResponseError should throw')
    } catch (e) {
      expect((e as ApiError).publicMessage).toBe('Une erreur est survenue.')
    }
  })
})
