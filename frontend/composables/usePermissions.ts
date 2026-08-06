import { rbacCan, type RbacAction, type RbacModule } from '~/utils/rbac'

export function usePermissions() {
  const { user } = useAuth()

  const profiles = computed(() => {
    const multi = user.value?.profiles
    if (Array.isArray(multi) && multi.length > 0) return multi
    const single = user.value?.profile
    return single ? [single] : []
  })

  const can = (module: RbacModule, action: RbacAction) =>
    rbacCan(profiles.value, module, action)

  const canValidateConges = computed(() => can('conges', 'V'))
  const canValidateTma = computed(() => can('tma', 'V'))
  const canValidateCra = computed(() => can('cra', 'V'))
  const canValidateEtt = computed(() => can('ett', 'V'))
  const canReadReporting = computed(() => can('reporting', 'L') || canValidateCra.value)

  return { can, canValidateConges, canValidateTma, canValidateCra, canValidateEtt, canReadReporting }
}
