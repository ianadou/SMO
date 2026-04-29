// @vitest-environment nuxt
import { describe, it, expect, beforeEach, vi } from 'vitest'
import { setActivePinia, createPinia } from 'pinia'
import { useAuthStore } from './auth'

const SAMPLE_ORGANIZER = {
  id: 'org-1',
  email: 'alex@example.fr',
  display_name: 'Alex',
  created_at: '2026-01-01T00:00:00Z',
}

describe('useAuthStore', () => {
  beforeEach(() => {
    window.localStorage.clear()
    setActivePinia(createPinia())
    vi.restoreAllMocks()
  })

  it('starts unauthenticated when localStorage is empty', () => {
    const auth = useAuthStore()
    expect(auth.token).toBeNull()
    expect(auth.organizer).toBeNull()
    expect(auth.isAuthenticated).toBe(false)
  })

  it('hydrates from localStorage on init', () => {
    window.localStorage.setItem('smo.auth.token', JSON.stringify('persisted-token'))
    window.localStorage.setItem('smo.auth.organizer', JSON.stringify(SAMPLE_ORGANIZER))

    const auth = useAuthStore()

    expect(auth.token).toBe('persisted-token')
    expect(auth.organizer).toEqual(SAMPLE_ORGANIZER)
    expect(auth.isAuthenticated).toBe(true)
  })

  it('setSession persists token + organizer to localStorage', () => {
    const auth = useAuthStore()
    auth.setSession('new-token', SAMPLE_ORGANIZER)

    expect(auth.isAuthenticated).toBe(true)
    expect(JSON.parse(window.localStorage.getItem('smo.auth.token')!)).toBe('new-token')
    expect(JSON.parse(window.localStorage.getItem('smo.auth.organizer')!)).toEqual(SAMPLE_ORGANIZER)
  })

  it('logout clears state and localStorage', () => {
    const auth = useAuthStore()
    auth.setSession('token', SAMPLE_ORGANIZER)
    auth.logout()

    expect(auth.token).toBeNull()
    expect(auth.organizer).toBeNull()
    expect(auth.isAuthenticated).toBe(false)
    expect(window.localStorage.getItem('smo.auth.token')).toBeNull()
    expect(window.localStorage.getItem('smo.auth.organizer')).toBeNull()
  })

  it('login writes the session returned by the API', async () => {
    const auth = useAuthStore()
    const fetchSpy = vi.spyOn(globalThis, '$fetch').mockResolvedValueOnce({
      token: 'jwt-from-server',
      organizer: SAMPLE_ORGANIZER,
    })

    await auth.login({ email: 'alex@example.fr', password: 'pwd' })

    expect(fetchSpy).toHaveBeenCalledOnce()
    expect(auth.token).toBe('jwt-from-server')
    expect(auth.organizer).toEqual(SAMPLE_ORGANIZER)
  })

  it('register calls register then login and ends authenticated', async () => {
    const auth = useAuthStore()
    const fetchSpy = vi.spyOn(globalThis, '$fetch')
      .mockResolvedValueOnce(SAMPLE_ORGANIZER)
      .mockResolvedValueOnce({ token: 'jwt', organizer: SAMPLE_ORGANIZER })

    await auth.register({
      email: 'alex@example.fr',
      password: 'pwd',
      display_name: 'Alex',
    })

    expect(fetchSpy).toHaveBeenCalledTimes(2)
    expect(auth.isAuthenticated).toBe(true)
  })
})
