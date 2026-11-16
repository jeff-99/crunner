## Why

crunner only offers recurring schedules (`intervalSchedule`, `cronSchedule` in `schedule.go`). Running a task once at a given time (a gap with gocron v2's `OneTimeJob`) currently requires a custom `Schedule` that has no clean way to say "no more runs": returning an error is logged as a failure and leaves the task registered forever (`scheduler.go` ~331-335).

## What Changes

- Add a one-time schedule constructor `At(t time.Time) Schedule` that triggers exactly once at `t`.
- Add an exported sentinel error `ErrNoMoreRuns` that any `Schedule` may return from `Next` to signal it is exhausted.
- The scheduler treats `ErrNoMoreRuns` as normal completion: no error is logged and the task is removed from the scheduler after its last run (or skip). Other schedule errors keep today's behaviour.
- The single run's window is [registration time (or NotBefore when backfilling), `t`); a `t` in the past fires immediately and is subject to MaxDrift like any late run.

## Capabilities

### New Capabilities

### Modified Capabilities
- `schedules`: adds the one-time schedule and the "no more runs" signal for pluggable schedules.
- `task-execution`: adds completion of a task whose schedule is exhausted.

## Impact

- `schedule.go`: new `At` constructor, `oneTimeSchedule` type, `ErrNoMoreRuns`.
- `scheduler.go`: `scheduleTask` handles `ErrNoMoreRuns` by removing the task from `s.tasks` without invoking the cancel callback.
- Not breaking: purely additive API. Custom schedules that already return errors are unaffected unless they return `ErrNoMoreRuns`.
- Dependencies: none required. If `tidy-public-api` lands first, follow its constructor naming (it already proposes `Every`/`Cron`; `At` matches). `add-run-limits` reuses the same task-completion path; whichever lands second should share the helper. `add-task-inspection` should expose completed tasks as gone.
