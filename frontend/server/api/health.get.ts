export default defineEventHandler(async () => {
  const config = useRuntimeConfig()
  const response = await $fetch<{ data: { status: string } }>(`${config.public.apiBase}/health`)
  return response
})
