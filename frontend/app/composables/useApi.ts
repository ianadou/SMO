import type { FetchOptions } from 'ofetch'

export class ApiError extends Error {
  readonly status: number
  readonly publicMessage: string

  constructor(status: number, publicMessage: string) {
    super(`api ${status}: ${publicMessage}`)
    this.status = status
    this.publicMessage = publicMessage
  }
}

export function useApi() {
  const config = useRuntimeConfig()
  const auth = useAuthStore()

  function request<T>(path: string, options: FetchOptions<'json'> = {}): Promise<T> {
    const headers: Record<string, string> = {
      Accept: 'application/json',
      ...(options.headers as Record<string, string> | undefined),
    }
    if (auth.token) {
      headers.Authorization = `Bearer ${auth.token}`
    }

    return $fetch<T>(path, {
      baseURL: config.public.apiBaseUrl,
      ...options,
      headers,
      onResponseError({ response }) {
        const body = response._data as { error?: string } | undefined
        const message = body?.error ?? 'Une erreur est survenue.'
        throw new ApiError(response.status, message)
      },
    })
  }

  return {
    get: <T>(path: string, options: FetchOptions<'json'> = {}) =>
      request<T>(path, { ...options, method: 'GET' }),
    post: <T>(path: string, body: unknown, options: FetchOptions<'json'> = {}) =>
      request<T>(path, { ...options, method: 'POST', body }),
    patch: <T>(path: string, body: unknown, options: FetchOptions<'json'> = {}) =>
      request<T>(path, { ...options, method: 'PATCH', body }),
    delete: <T>(path: string, options: FetchOptions<'json'> = {}) =>
      request<T>(path, { ...options, method: 'DELETE' }),
  }
}
