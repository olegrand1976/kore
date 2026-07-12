export default defineEventHandler(async (event) => {
  const config = useRuntimeConfig()
  const body = await readBody(event)
  return $fetch(`${config.public.apiBase}/api/v1/public/booking/appointments`, {
    method: 'POST',
    body
  })
})
