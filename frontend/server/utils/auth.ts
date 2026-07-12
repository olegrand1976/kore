export function apiAuthHeaders(event: Parameters<typeof defineEventHandler>[0] extends never ? never : import('h3').H3Event): Record<string, string> {
  const token = getCookie(event, 'kore_access_token')
  if (!token) {
    return {}
  }
  return { Authorization: `Bearer ${token}` }
}

export function apiBase(): string {
  return useRuntimeConfig().public.apiBase
}
