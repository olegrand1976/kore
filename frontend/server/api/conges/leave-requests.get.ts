import { rbacCan } from '~/utils/rbac'

export default defineEventHandler(async (event) => {
  const headers = apiAuthHeaders(event)
  const query = { ...getQuery(event) }
  const session = parseSessionFromEvent(event)

  if (session?.userId && !query.userId && !rbacCan(session.profile, 'conges', 'V')) {
    query.userId = session.userId
  }

  return $fetch(`${apiBase()}/api/v1/leave-requests`, { headers, query })
})
