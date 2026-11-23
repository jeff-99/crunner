## Why

Every due run is started in its own goroutine immediately (`scheduler.go:426-469`), so the number of simultaneously executing tasks is unbounded. Applications with many tasks (or a burst of backfill runs) need to cap concurrent work to protect databases and other shared resources, as gocron v2 does with its scheduler-wide limit mode.

## What Changes

- Add scheduler option `WithConcurrencyLimit(n int, mode LimitMode)` capping concurrently executing runs across all tasks.
- `LimitWait`: due runs queue in trigger order and start when a slot frees; MaxDrift is checked when the run actually starts, so long waits lead to drift skips.
- `LimitSkip`: a run that becomes due while all slots are busy is skipped (skip callback) and the task's next window is scheduled.
- Without the option, behaviour is unchanged (unlimited).
- Queued runs are discarded when the scheduler stops or the task is removed/replaced.

## Capabilities

### New Capabilities
- `run-concurrency`: scheduler-wide bound on the number of concurrently executing runs and the handling of runs that exceed it.

### Modified Capabilities
<!-- none: per-task execution rules are unchanged -->

## Impact

- Code: `scheduler.go` (new option, `LimitMode` type, slot accounting and wait queue in the main loop); tests.
- API: additive; no breaking change.
- Dependencies: `fix-shared-state-races` (main loop owns run state and completion messages), `fix-mainloop-self-send-deadlock` (queueing in the main loop must not self-send on a bounded channel), `split-cancel-callback` (limit skips use the "skipped" callback). Note: `New` has no error return, so an invalid limit (n < 1) panics (see design).
