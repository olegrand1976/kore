export default defineNuxtRouteMiddleware(async (to) => {
  const publicRoutes = ['/', '/login', '/modules', '/tarifs', '/reserver', '/contact']
  if (publicRoutes.some((route) => to.path === route || to.path.startsWith(route + '/'))) {
    return
  }

  const { user, fetchSession } = useAuth()
  if (!user.value) {
    await fetchSession()
  }
  if (!user.value?.ok) {
    return navigateTo('/login')
  }
})
