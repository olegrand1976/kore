export default defineEventHandler(async () => {
  const config = useRuntimeConfig()
  return $fetch(`${config.public.apiBase}/api/v1/public/booking/slots`)
})
