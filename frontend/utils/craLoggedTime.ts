export type LoggedTimeWeek = {
  lines: Array<{ duration?: number | string | null }>
}

/** True when any CRA line on the month carries a positive duration (minutes). */
export function timesheetHasLoggedTime(weeks: LoggedTimeWeek[] | null | undefined): boolean {
  if (!weeks?.length) return false
  return weeks.some((week) =>
    (week.lines ?? []).some((line) => Number(line.duration) > 0)
  )
}
