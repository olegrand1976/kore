export type AdminTimesheetAction = 'unvalidate' | 'delete'

export function timesheetAdminAction(status: string): AdminTimesheetAction {
  return status === 'Définitif' ? 'unvalidate' : 'delete'
}

export function timesheetAdminConfirmKey(
  action: AdminTimesheetAction,
  hasUser: boolean
): 'cra.unvalidate_confirm' | 'cra.unvalidate_confirm_simple' | 'cra.delete_confirm' | 'cra.delete_confirm_simple' {
  switch (action) {
    case 'unvalidate':
      return hasUser ? 'cra.unvalidate_confirm' : 'cra.unvalidate_confirm_simple'
    case 'delete':
      return hasUser ? 'cra.delete_confirm' : 'cra.delete_confirm_simple'
    default: {
      const _exhaustive: never = action
      return _exhaustive
    }
  }
}
