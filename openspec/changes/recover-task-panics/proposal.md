## Why

A panic inside a task function runs on an unguarded goroutine (`processTask`, scheduler.go ~427-469) and crashes the whole host process. A scheduler library should contain failures of a single task, not take down the application.

## What Changes

- Recover panics raised by the task function in the task goroutine.
- Convert a recovered panic into an error of a new exported type `*PanicError` (carrying the panic value and stack trace) and treat it exactly like a returned error: error count, `LastError`, error log, error callback.
- After a recovered panic the task is rescheduled as usual and the scheduler keeps running.
- New exported type `PanicError` (additive, not breaking). No other API change.

## Capabilities

### New Capabilities

### Modified Capabilities
- `task-execution`: the Callbacks requirement now covers panicking runs; a new requirement states that panics are contained.

## Impact

- Code: `scheduler.go` (task goroutine in `processTask`), new `PanicError` type (e.g. in `task.go`), tests in `scheduler_test.go`.
- API: additive only — new exported `PanicError` type with `Value any`, `Stack []byte`, `Error()` and `Unwrap()` (returns the value when it is an `error`). No breaking change.
- Dependencies on other changes: none; implementable on main. `replace-summary-logging-with-hooks` will later add a dedicated panic hook on top of this; `fix-shared-state-races` touches the same goroutine, so whichever lands second rebases.
