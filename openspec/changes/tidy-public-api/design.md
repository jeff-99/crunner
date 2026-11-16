# Design

## Context

`schedule.go` exports `ParseCron` (it returns `cron.Schedule`), `WithCronSchedule` (it returns the unexported `cronSchedule` with exported fields) and `WithIntervalSchedule` (it returns the unexported `intervalSchedule`). In `scheduler.go`, `SchedulingOptions` holds the `*SuccessCallback`, `*ErrorCallback` and `*CancelCallback` pointers, and the call sites dereference them. `WithDefaultOptions()` returns a `SchedulingOptions` value and is also used internally by `RegisterTask`.

## Goals / Non-Goals

**Goals:**
- Make the schedule constructors idiomatic and return the `Schedule` interface.
- Remove third-party types from the exported API.
- Use plain func callbacks.
- Rename to a clear defaults function.

**Non-Goals:**
- Any behaviour change, including the cron nil-location handling, which belongs to `fix-cron-time-zone-handling`.
- `ScheduledTasks()` (`add-task-inspection`).
- `RegisterTask` errors (`validate-task-registration`).
- Logger type (`replace-zerolog-with-slog`).
- Renaming `SchedulingOption`/`SchedulerOption`.

## Decisions

- **Return the `Schedule` interface rather than exported concrete types.** This keeps the concrete types free to change (for example, replacing the parser) and matches how `RegisterTask` consumes them. Considered: exporting `IntervalSchedule` and `CronSchedule` so callers can inspect the spec or interval. `add-task-inspection` can add a `String()` method or similar if needed.
- **`Cron` takes functional options (`CronOption`).** It has `CronLocation(loc)` and `CronWithSeconds()`. These names don't collide with `SchedulingOption`s such as `WithID`, and new options (such as a cron `TZ=` prefix policy) can be added without breaking the signature. The default location stays whatever main does today, because this change doesn't fix time-zone handling.
- **`ParseCron` becomes `parseCron`.** No replacement is exported. Users validate an expression by calling `Cron` and checking the error.
- **Plain func callbacks.** A nil func means no callback. The `With*Callback` options assign `f` directly. Considered: keeping pointers so that nil and "set to nil" stay distinct. That distinction is not useful.
- **`DefaultSchedulingOptions()`.** It is a pure rename with no deprecated alias, since the project is pre-1.0 and the proposal documents the break.

## Risks / Trade-offs

- [Wide churn in tests and in the datahub-engine consumer] → It is a mechanical rewrite that follows the migration table in the proposal. A single PR avoids an awkward half-migrated state.
- [Conflicts with in-flight changes that touch the callbacks] → Called out in the proposal. The rebase is mechanical.
