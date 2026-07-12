export default defineNuxtRouteMiddleware((to) => {
  const publicRoutes = ['/', '/login', '/modules', '/tarifs', '/reserver', '/contact']
  if (publicRoutes.some((route) => to.path === route || to.path.startsWith(route + '/'))) {
    return
  }
})
