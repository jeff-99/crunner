## Context

The task function runs in a goroutine started by `processTask` (scheduler.go ~427). The call is already wrapped in an inner closure (`func() { ... err = taskUnit.task(taskCtx, metaCopy) }()`, ~432-436), which is the natural place for a `recover`. Everything after that closure (metadata update, callbacks, reschedule via `newTaskCh`) already handles a non-nil `err`.

## Goals / Non-Goals

**Goals:**
- Contain panics from the task function and surface them through the existing error path.
- Keep enough information (value + stack) to debug.

**Non-Goals:**
- Recovering panics in user callbacks (`OnSuccess`/`OnError`/`OnCancel`) or custom `Schedule.Next`; those are left for a later change (or the hooks change).
- A separate panic callback/hook — that belongs to `replace-summary-logging-with-hooks`.

## Decisions

- **Recover inside the existing inner closure** with `defer func(){ if r := recover(); r != nil { err = &PanicError{Value: r, Stack: debug.Stack()} } }()`. This guarantees `taskCancel` still runs and the rest of the goroutine (counts, callbacks, `wg.Done`, reschedule) is unchanged. Alternative: recover at the top of the goroutine — rejected, it would skip the metadata update and the reschedule, silently dropping the task.
- **Exported `PanicError` type** so callers can `errors.As` it. `Unwrap()` returns `Value` when it is an `error`, so `errors.Is` works with panics like `panic(io.EOF)`. Alternative: `fmt.Errorf("panic: %v")` — rejected, loses stack and type information.
- **Stack via `runtime/debug.Stack()`** captured inside the deferred function so it includes the panicking frame. Standard library only.
- **Logged at error level** like other failed runs; the message includes the stack.

## Risks / Trade-offs

- [A task that panics while holding locks or leaving shared state inconsistent may misbehave on the next run] → Documented: recovery keeps the scheduler alive but does not make the task's own state consistent.
- [`panic(nil)` / `runtime.Goexit`] → Since Go 1.21 `panic(nil)` yields `*runtime.PanicNilError`, which is recovered normally. `runtime.Goexit` is not a panic; the deferred `wg.Done` still runs but the task is not rescheduled — out of scope, noted in a code comment.
