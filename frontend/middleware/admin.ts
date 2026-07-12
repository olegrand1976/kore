export default defineNuxtRouteMiddleware(async () => {
  const { fetchSession, isAdmin } = useAuth()
  await fetchSession()
  if (!isAdmin.value) {
    return navigateTo('/dashboard')
  }
})
