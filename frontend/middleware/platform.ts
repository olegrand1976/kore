export default defineNuxtRouteMiddleware(async () => {
  const { fetchSession, isPlatformAdmin } = useAuth()
  await fetchSession()
  if (!isPlatformAdmin.value) {
    return navigateTo('/dashboard')
  }
})
