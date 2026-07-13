import { rbacCan, type RbacAction, type RbacModule } from '~/utils/rbac'

export function usePermissions() {
  const { user } = useAuth()

  const profile = computed(() => user.value?.profile)

  const can = (module: RbacModule, action: RbacAction) =>
    rbacCan(profile.value, module, action)

  const canValidateConges = computed(() => can('conges', 'V'))
  const canValidateTma = computed(() => can('tma', 'V'))
  const canValidateCra = computed(() => can('cra', 'V'))

  return { can, canValidateConges, canValidateTma, canValidateCra }
}
