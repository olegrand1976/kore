/** Sync select from demand assignee when present; otherwise leave unchanged. */
export function syncAssigneeFromDemand(
  assigneeId: string | undefined | null,
  current: string
): string {
  const fromDemand = assigneeId ? String(assigneeId) : ''
  return fromDemand || current
}

/** Default to first user only when the select is still empty. */
export function syncAssigneeFromUsers(
  current: string,
  users?: { id: string }[] | null
): string {
  if (current) return current
  return users?.[0]?.id ?? ''
}
