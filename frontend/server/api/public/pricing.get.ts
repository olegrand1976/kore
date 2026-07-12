function apiBase() {
  const config = useRuntimeConfig()
  return config.public.apiBase as string
}

export default defineEventHandler(async () => {
  return $fetch(`${apiBase()}/api/v1/public/pricing`)
})
