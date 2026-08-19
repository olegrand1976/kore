import type { MaybeRef } from 'vue'
import { toValue } from 'vue'

export type MethodologyProfile = 'psa' | 'agile_scrum' | 'agile_kanban'

export const METHODOLOGY_PROFILES: MethodologyProfile[] = ['psa', 'agile_scrum', 'agile_kanban']

export function isAgileProfile(profile: MethodologyProfile | string | undefined | null): boolean {
  return profile === 'agile_scrum' || profile === 'agile_kanban'
}

export function useMethodologyTerms(profile: MaybeRef<MethodologyProfile | string | undefined | null>) {
  const { t } = useI18n()

  return computed(() => {
    const normalized = (toValue(profile) ?? 'psa') as MethodologyProfile

    const suffix = (() => {
      switch (normalized) {
        case 'agile_scrum':
          return 'scrum'
        case 'agile_kanban':
          return 'kanban'
        default:
          return 'psa'
      }
    })()

    return {
      profile: normalized,
      isAgile: isAgileProfile(normalized),
      application: t(`project.terms.application.${suffix}`),
      workItem: t(`project.terms.work_item.${suffix}`),
      workItemNew: t(`project.terms.work_item_new.${suffix}`),
      release: t(`project.terms.release.${suffix}`),
      backlog: t(`project.terms.backlog.${suffix}`),
      estimation: t(`project.terms.estimation.${suffix}`),
      board: t(`project.terms.board.${suffix}`)
    }
  })
}
