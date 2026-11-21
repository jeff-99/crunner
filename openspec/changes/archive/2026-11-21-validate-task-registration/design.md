# Design

## Context

`RegisterTask` (scheduler.go:208-277) applies options to `WithDefaultOptions()`, builds a `schedulingUnit`, runs the overwrite handling (stop timer, fire `onCancel`), stores the unit and returns nil. It never checks its input. Callback options store `&f` (scheduler.go:92-108), so a nil func produces a non-nil pointer that panics when dereferenced and called. Backfill only takes effect when `NotBefore != nil` (scheduler.go:291); otherwise it is silently ignored.

## Goals / Non-Goals

**Goals:**
- Check the arguments in one place, before anything has side effects.
- Return errors that callers can match.

**Non-Goals:**
- ID validation. That belongs to `require-stable-task-ids`.
- Validating custom `Schedule` implementations beyond the nil check. Their errors still show up through `Next`.
- Checking `NotBefore`/`NotAfter` against the current time.

## Decisions

- **Validate first, then mutate.** A `validate(task, schedule, opts) error` helper runs immediately after the options are applied and before the overwrite block. If it returns an error, `RegisterTask` returns straight away. That is what keeps an existing task intact when the new registration is rejected.
- **Sentinel errors + `errors.Join`.** Add `ErrInvalidTask`, `ErrInvalidSchedule` and `ErrInvalidOption`. Each problem is wrapped as `fmt.Errorf("%w: maxDrift must be >= 0", ErrInvalidOption)` and the problems are joined. Considered: a structured `ValidationError` type. It is more API surface than needed today.
- **Nil schedule check covers typed nils.** Check `schedule == nil`, and also use a `reflect.ValueOf(schedule)` nil check for pointer, func, map and similar kinds, so a `(*mySchedule)(nil)` is caught as well. `reflect` is in the standard library.
- **Interval check through an unexported method.** Add `validate() error` on the interval schedule type and check for it with an unexported interface (`interface{ validate() error }`). This way the cron schedule can add its own checks later without changes in `RegisterTask`. A zero-value cron schedule (nil inner `cron.Schedule`) is rejected the same way.
- **Backfill without NotBefore is an error, not a no-op.** It is almost certainly a misconfiguration. Considered: defaulting `NotBefore` to the registration time. That would make backfill a silent no-op again.

## Risks / Trade-offs

- [Existing callers relying on silently ignored backfill start getting errors] → Called out as BREAKING in the proposal. The fix is to add `WithNotBefore` or drop `WithBackfill`.
- [Typed-nil check through reflection is slightly magical] → It is limited to nil-able kinds and covered by a test.
