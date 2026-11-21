# Proposal

## Why

`RegisterTask` (scheduler.go:208) accepts anything and always returns nil. A nil task or nil schedule is stored and only fails later, with a panic on the main loop or timer goroutine. `WithIntervalSchedule(0)` makes the task fire over and over with no pause. Nonsensical options (negative MaxDrift, `NotAfter` before `NotBefore`, backfill without `NotBefore`, nil callbacks) are silently accepted. Configuration mistakes should fail fast, at the point of registration.

## What Changes

- `RegisterTask` validates its arguments and returns a descriptive error, without storing anything, when:
  - the task is nil;
  - the schedule is nil;
  - a built-in interval schedule has a non-positive interval;
  - `MaxDrift` is negative;
  - `NotBefore` and `NotAfter` are both set and `NotAfter` is not after `NotBefore`;
  - backfill is enabled without `NotBefore`. Today the backfill flag is silently ignored in that case (scheduler.go:291);
  - a success, error or cancel callback option is given a nil function. Today that panics when the callback fires (scheduler.go:253/409/460/468).
- A failed registration leaves any existing task with the same ID untouched. It is not cancelled or replaced.
- Exported sentinel errors (e.g. `ErrInvalidTask`, `ErrInvalidSchedule`, `ErrInvalidOption`) so callers can use `errors.Is`. All problems are reported together with `errors.Join`.

## Capabilities

### New Capabilities

### Modified Capabilities
- `task-registration`: registration now validates the task, schedule and options, and can return an error.

## Impact

- Code: `scheduler.go` (`RegisterTask`, new validation helper, sentinel errors), `schedule.go` (an interval-validity check that `RegisterTask` can call).
- **BREAKING**: `RegisterTask` now returns errors that used to be impossible. Some registrations that were accepted before are now rejected: backfill without `NotBefore`, a zero interval, negative drift and nil callbacks.
- Dependencies on other changes: none required.
- Coordination with other changes:
  - `require-stable-task-ids` owns empty-ID handling, so this change does not validate the ID.
  - `tidy-public-api` renames the interval constructor to `Every`. The interval check must follow whichever name is on main.
  - `add-run-limits` gives `NotAfter` its meaning. This change only checks that the range is consistent.
