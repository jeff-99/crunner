# Proposal

## Why

The public API has several rough edges that are left over from its extraction out of datahub-engine:
- Schedules are built with `WithCronSchedule` and `WithIntervalSchedule`. These look like functional options but aren't, and they return the unexported types `cronSchedule` and `intervalSchedule`, which callers can't name.
- `ParseCron` leaks `github.com/robfig/cron/v3`'s `cron.Schedule` type into the public API.
- Callbacks are stored as pointer-to-func in `SchedulingOptions`.
- `WithDefaultOptions()` returns a struct rather than an option.

These rough edges should be fixed before v1.0.0, while breaking changes are still cheap.

## What Changes

- **BREAKING**: replace `WithIntervalSchedule(d)` with `Every(d time.Duration) Schedule`.
- **BREAKING**: replace `WithCronSchedule(spec, loc, includeSeconds)` with `Cron(spec string, opts ...CronOption) (Schedule, error)`. It takes the options `CronLocation(*time.Location)` and `CronWithSeconds()`. The default location stays `time.Local`, and a nil location means `time.Local`.
- **BREAKING**: make `ParseCron` unexported, so the `robfig/cron` type is no longer part of the API.
- **BREAKING**: `SchedulingOptions.OnSuccess`, `OnError` and `OnCancel` become plain func fields (`SuccessCallback`, etc.), where nil means no callback.
- `WithSuccessCallback(nil)`, `WithErrorCallback(nil)` and `WithCancelCallback(nil)` are still rejected by `RegisterTask` with `ErrInvalidOption` (the behaviour from `validate-task-registration`): the option records the problem on `SchedulingOptions` in a private field that registration validation reads, so a nil func passed explicitly is distinguishable from "no callback option given".
- **BREAKING**: rename `WithDefaultOptions()` to `DefaultSchedulingOptions()`. It keeps the same return value, and the new name matches what it does.
- No other behaviour changes.

## Capabilities

### New Capabilities

### Modified Capabilities
- `schedules`: the interval and cron schedule requirements describe the new constructors, which return a `Schedule`. Cron options replace positional parameters. All other text of the requirements (reference time is the previous window end, default and per-run time zone handling, `CRON_TZ` precedence) is unchanged.

## Impact

- Code: `schedule.go`, `scheduler.go` (option types, callback invocation sites, `DefaultSchedulingOptions`), `validate.go` (reads the recorded callback problems instead of dereferencing pointers), all tests and examples.
- **BREAKING**: every caller that builds schedules or reads `SchedulingOptions` callback fields must update. The proposal includes a migration table.
- Dependencies on other changes: none outstanding. The changes it touches (`fix-cron-time-zone-handling`, `validate-task-registration`, `make-trigger-windows-contiguous`, `require-stable-task-ids`) are merged and are preserved.
- Coordination with other changes:
  - `replace-robfig-cron-parser` can swap the parser without touching the API, because this change makes `ParseCron` unexported.
  - `split-cancel-callback` touches the same callback fields and adapts to plain funcs.
  - `ScheduledTasks()` returning bare funcs is left to `add-task-inspection`.

### Migration

| Before | After |
|---|---|
| `WithIntervalSchedule(5*time.Second)` | `Every(5*time.Second)` |
| `WithCronSchedule("*/5 * * * *", loc, false)` | `Cron("*/5 * * * *", CronLocation(loc))` |
| `WithCronSchedule("* * * * * *", loc, true)` | `Cron("* * * * * *", CronLocation(loc), CronWithSeconds())` |
| `if o.OnError != nil { (*o.OnError)(ctx, err) }` | `if o.OnError != nil { o.OnError(ctx, err) }` |
| `WithDefaultOptions()` | `DefaultSchedulingOptions()` |
