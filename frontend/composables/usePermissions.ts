import { rbacCan, type RbacAction, type RbacModule } from '~/utils/rbac'

export function usePermissions() {
  const { effectiveProfiles } = useAuth()

  const can = (module: RbacModule, action: RbacAction) =>
    rbacCan(effectiveProfiles.value, module, action)

  const canValidateConges = computed(() => can('conges', 'V'))
  const canValidateTma = computed(() => can('tma', 'V'))
  const canValidateCra = computed(() => can('cra', 'V'))
  const canValidateEtt = computed(() => can('ett', 'V'))
  const canReadReporting = computed(() => can('reporting', 'L') || canValidateCra.value)

  return { can, canValidateConges, canValidateTma, canValidateCra, canValidateEtt, canReadReporting }
}
