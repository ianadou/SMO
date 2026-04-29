const PUBLIC_AUTH_ROUTES = new Set(['/login', '/register'])

export default defineNuxtRouteMiddleware((to) => {
  const auth = useAuthStore()
  if (auth.isAuthenticated && PUBLIC_AUTH_ROUTES.has(to.path)) {
    return navigateTo('/groups', { replace: true })
  }
})
