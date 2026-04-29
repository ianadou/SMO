import { defineStore } from 'pinia'
import type {
  LoginPayload,
  LoginResponseDTO,
  OrganizerDTO,
  RegisterPayload,
} from '~/types/auth'

const TOKEN_STORAGE_KEY = 'smo.auth.token'
const ORGANIZER_STORAGE_KEY = 'smo.auth.organizer'

interface AuthState {
  token: string | null
  organizer: OrganizerDTO | null
}

function readStorage<T>(key: string): T | null {
  if (typeof window === 'undefined') return null
  const raw = window.localStorage.getItem(key)
  if (!raw) return null
  try {
    return JSON.parse(raw) as T
  } catch {
    return null
  }
}

function writeStorage(key: string, value: unknown) {
  if (typeof window === 'undefined') return
  window.localStorage.setItem(key, JSON.stringify(value))
}

function clearStorage() {
  if (typeof window === 'undefined') return
  window.localStorage.removeItem(TOKEN_STORAGE_KEY)
  window.localStorage.removeItem(ORGANIZER_STORAGE_KEY)
}

export const useAuthStore = defineStore('auth', {
  state: (): AuthState => ({
    token: readStorage<string>(TOKEN_STORAGE_KEY),
    organizer: readStorage<OrganizerDTO>(ORGANIZER_STORAGE_KEY),
  }),

  getters: {
    isAuthenticated: (state): boolean => state.token !== null,
  },

  actions: {
    async login(payload: LoginPayload) {
      const api = useApi()
      const response = await api.post<LoginResponseDTO>('/auth/login', payload)
      this.setSession(response.token, response.organizer)
    },

    async register(payload: RegisterPayload) {
      const api = useApi()
      await api.post<OrganizerDTO>('/auth/register', payload)
      await this.login({ email: payload.email, password: payload.password })
    },

    setSession(token: string, organizer: OrganizerDTO) {
      this.token = token
      this.organizer = organizer
      writeStorage(TOKEN_STORAGE_KEY, token)
      writeStorage(ORGANIZER_STORAGE_KEY, organizer)
    },

    logout() {
      this.token = null
      this.organizer = null
      clearStorage()
    },
  },
})
