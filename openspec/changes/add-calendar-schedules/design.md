## Context

`Schedule.Next(ctx, schedulingStart, lastRun)` (`schedule.go` line 11) returns the end of the next window. The scheduler passes `lastRun` = previous window end (backfill) or last actual run start (normal; to become window end with `make-trigger-windows-contiguous`). The calendar schedules only need a pure "next occurrence strictly after base" function.

## Goals / Non-Goals

**Goals:**
- Stdlib-only, allocation-light `Next` computation.
- Explicit, tested DST behaviour.

**Non-Goals:**
- Last-day-of-month / "nth weekday" rules, intervals like "every 2 weeks" (cron or custom schedules cover these).
- Changing cron or interval schedules.

## Decisions

- **One internal type `calendarSchedule`** holding sorted, deduplicated times, a weekday set (7 bits) and a day-of-month set (31 bits); `Daily` = all weekdays, all days. `Next` walks day-by-day from base's date in `loc` (bounded loop: any valid day set recurs within 62 days), picking the first candidate `time.Date(...)` strictly after base. Alternative: translate to cron expressions — rejected, couples to the cron parser that `replace-robfig-cron-parser` is replacing and hides validation errors.
- **Base time:** `lastRun` if non-nil, else `schedulingStart`; candidates must be strictly after base, so a run at exactly a scheduled time advances to the next occurrence.
- **DST:** `time.Date` normalises non-existent wall times forward (02:30 in a gap → 03:30), which matches the spec. For ambiguous times the Go docs leave the chosen occurrence unspecified, so we compute the instant explicitly with the pre-transition offset (the first occurrence). Because the walk is by calendar date and candidates must be strictly after base, the second occurrence of the same wall time is never produced. Tests pin the behaviour with `Europe/Amsterdam`.
- **`TimeOfDay`** is a small struct with `NewTimeOfDay(h, m, s int) (TimeOfDay, error)`; constructors re-validate so zero/struct-literal values are checked too.
- **nil location → `time.Local`**, consistent with `fix-cron-time-zone-handling`.

## Risks / Trade-offs

- [Ambiguous-time handling is subtle] → dedicated table-driven tests around both 2026 transitions in Europe/Amsterdam and America/New_York.
- [Day-by-day walk is O(days)] → bounded to a few hundred iterations worst case; negligible versus scheduling frequency.
