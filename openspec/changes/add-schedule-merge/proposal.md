## Why

Some real schedules cannot be written as one cron expression. Cron fields are a cross product, so "06:00 and 18:30 every day" needs two expressions, and so does "every weekday at 09:00 plus the 1st of the month at 00:00". Today a user has to register several tasks for this, which splits one logical job into several IDs, windows and statistics. gocron v2 has no combinator either, but `add-calendar-schedules` proposed three calendar constructors mainly to cover this gap.

## What Changes

- Add `Merge(schedules ...Schedule) (Schedule, error)`, a schedule whose next run is the earliest next run of any of its children.
- Every child is asked for its next run with the same scheduling start and reference time; the merged schedule returns the minimum. Windows are unchanged: each window ends at the next occurrence of any child.
- Instants produced by several children at once collapse into one run.
- An error from any child is returned as is; the merged schedule never silently ignores a broken child.
- `Merge` returns an error for an empty list or a nil child. A single child is returned unchanged.
- Additive, not breaking.

## Capabilities

### New Capabilities

### Modified Capabilities
- `schedules`: adds the merged schedule.

## Impact

- `schedule.go` (or a new `merge.go`): one small unexported type and the constructor. Standard library only.
- Not breaking.
- Dependencies: none. It only uses the existing `Schedule` interface, so it works with every schedule, including user-defined ones, and with `Every` and `Cron` as renamed by `tidy-public-api`.
- Interacts with `add-one-time-schedule`: once `ErrNoMoreRuns` exists, an exhausted child must be dropped and the merged schedule must only report `ErrNoMoreRuns` when all children are exhausted. That belongs in whichever of the two changes lands second.
- Replaces the motivation for `add-calendar-schedules`, which is closed in favour of this change plus a DST-aware cron schedule (`replace-robfig-cron-parser`).
