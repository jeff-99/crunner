# Design

## Context

`schedule.go` exports `ParseCron` (it returns `cron.Schedule`), `WithCronSchedule` (it returns the unexported `cronSchedule` with exported fields) and `WithIntervalSchedule` (it returns the unexported `intervalSchedule`). In `scheduler.go`, `SchedulingOptions` holds the `*SuccessCallback`, `*ErrorCallback` and `*CancelCallback` pointers, and the call sites dereference them. `validate.go` uses the pointers to tell "callback option given with a nil func" (rejected, `ErrInvalidOption`) from "no callback option" (fine). `WithDefaultOptions()` returns a `SchedulingOptions` value and is also used internally by `RegisterTask`.

## Goals / Non-Goals

**Goals:**
- Make the schedule constructors idiomatic and return the `Schedule` interface.
- Remove third-party types from the exported API.
- Use plain func callbacks.
- Rename to a clear defaults function.

**Non-Goals:**
- Any behaviour change. Cron time-zone handling (already fixed) and registration validation (already merged) must stay exactly as they are.
- `ScheduledTasks()` (`add-task-inspection`).
- `RegisterTask` errors (`validate-task-registration`).
- Logger type (`replace-zerolog-with-slog`).
- Renaming `SchedulingOption`/`SchedulerOption`.

## Decisions

- **Return the `Schedule` interface rather than exported concrete types.** This keeps the concrete types free to change (for example, replacing the parser) and matches how `RegisterTask` consumes them. Considered: exporting `IntervalSchedule` and `CronSchedule` so callers can inspect the spec or interval. `add-task-inspection` can add a `String()` method or similar if needed.
- **`Cron` takes functional options (`CronOption`).** It has `CronLocation(loc)` and `CronWithSeconds()`. These names don't collide with `SchedulingOption`s such as `WithID`, and new options (such as a cron `TZ=` prefix policy) can be added without breaking the signature. No location option, or a nil location, means `time.Local`, as today.
- **`ParseCron` becomes `parseCron`.** No replacement is exported. Users validate an expression by calling `Cron` and checking the error.
- **Plain func callbacks, nil still rejected.** A nil func field means no callback, and the `With*Callback` options assign `f` directly. To keep `WithSuccessCallback(nil)` an error (the requirement from `validate-task-registration`), each option that receives a nil func also appends a message to an unexported `invalid []string` field on `SchedulingOptions`; `validateRegistration` turns each message into an `ErrInvalidOption`. Alternatives considered: treat nil as "no callback" (loses a deliberate validation, so a forgotten callback variable would go unnoticed); keep the pointer fields (leaves the awkward API); generic option parameters so that a literal `nil` fails to compile (catches only the literal nil, not a nil variable, and puts a type parameter on three public options).
- **`DefaultSchedulingOptions()`.** It is a pure rename with no deprecated alias, since the project is pre-1.0 and the proposal documents the break.

## Risks / Trade-offs

- [Tests that read `schedule.Location` or build a zero-value `cronSchedule` can no longer use the public constructor's concrete type] → Those checks move to in-package tests that use an unexported constructor.

- [Wide churn in tests and in the datahub-engine consumer] → It is a mechanical rewrite that follows the migration table in the proposal. A single PR avoids an awkward half-migrated state.
- [Conflicts with in-flight changes that touch the callbacks] → Called out in the proposal. The rebase is mechanical.
