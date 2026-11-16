# Proposal

## Why

The public API has several rough edges that are left over from its extraction out of datahub-engine:
- Schedules are built with `WithCronSchedule` and `WithIntervalSchedule` (schedule.go:36/69). These look like functional options but aren't, and they return the unexported types `cronSchedule` and `intervalSchedule`, which callers can't name.
- `ParseCron` leaks `github.com/robfig/cron/v3`'s `cron.Schedule` type into the public API.
- Callbacks are stored as pointer-to-func in `SchedulingOptions` (scheduler.go:47-49).
- `WithDefaultOptions()` returns a struct rather than an option.

These rough edges should be fixed before v1.0.0, while breaking changes are still cheap.

## What Changes

- **BREAKING**: replace `WithIntervalSchedule(d)` with `Every(d time.Duration) Schedule`.
- **BREAKING**: replace `WithCronSchedule(spec, loc, includeSeconds)` with `Cron(spec string, opts ...CronOption) (Schedule, error)`. It takes the options `CronLocation(*time.Location)` and `CronWithSeconds()`.
- **BREAKING**: make `ParseCron` unexported, so the `robfig/cron` type is no longer part of the API.
- **BREAKING**: `SchedulingOptions.OnSuccess`, `OnError` and `OnCancel` become plain func fields (`SuccessCallback`, etc.), where nil means no callback.
- **BREAKING**: rename `WithDefaultOptions()` to `DefaultSchedulingOptions()`. It keeps the same return value, and the new name matches what it does.
- No behaviour changes.

## Capabilities

### New Capabilities

### Modified Capabilities
- `schedules`: the interval and cron schedule requirements describe the new constructors, which return a `Schedule`. Cron options replace positional parameters.

## Impact

- Code: `schedule.go`, `scheduler.go` (option types, callback invocation sites at lines 245-255, 408-411, 458-470, 667-671), all tests and examples.
- **BREAKING**: every caller that builds schedules or reads `SchedulingOptions` callback fields must update. The proposal includes a migration table.
- Dependencies on other changes: none.
- Coordination with other changes:
  - `fix-cron-time-zone-handling` changes the nil-location behaviour behind `CronLocation`.
  - `replace-robfig-cron-parser` can then swap the parser without touching the API, because this change makes `ParseCron` unexported.
  - `validate-task-registration` and `split-cancel-callback` touch the same callback fields. Whichever lands second adapts to plain funcs.
  - `ScheduledTasks()` returning bare funcs is left to `add-task-inspection`.
  - `RegisterTask` errors are left to `validate-task-registration`.

### Migration

| Before | After |
|---|---|
| `WithIntervalSchedule(5*time.Second)` | `Every(5*time.Second)` |
| `WithCronSchedule("*/5 * * * *", loc, false)` | `Cron("*/5 * * * *", CronLocation(loc))` |
| `WithCronSchedule("* * * * * *", loc, true)` | `Cron("* * * * * *", CronLocation(loc), CronWithSeconds())` |
| `if o.OnError != nil { (*o.OnError)(ctx, err) }` | `if o.OnError != nil { o.OnError(ctx, err) }` |
| `WithDefaultOptions()` | `DefaultSchedulingOptions()` |
