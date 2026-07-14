export const ABSENCE_SOURCE_TYPES = ['absence', 'conge', 'leave', 'holiday'] as const

export type AbsenceSourceType = (typeof ABSENCE_SOURCE_TYPES)[number]

export function isAbsenceSourceType(sourceType: string): sourceType is AbsenceSourceType {
  return (ABSENCE_SOURCE_TYPES as readonly string[]).includes(sourceType)
}

export function absenceDayClass(sourceType: string): string {
  switch (sourceType) {
    case 'holiday':
      return 'day-block--absence-holiday'
    case 'leave':
    case 'conge':
      return 'day-block--absence-leave'
    default:
      return 'day-block--absence-other'
  }
}
