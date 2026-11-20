## Why

`processTask` runs on the main loop. On the drift-skip path it sends the unit back on `newTaskCh` (`scheduler.go:413-417`), a channel only the main loop drains (buffer 100). If the buffer is full — e.g. many tasks finishing or being registered while a burst of tasks is skipped after a suspend/resume — the main loop blocks on itself and the whole scheduler stalls until Stop. The same path also calls the user's cancel callback synchronously on the main loop (`scheduler.go:408-411`), so a slow or blocking callback stalls all scheduling.

## What Changes

- On the drift-skip path the main loop reschedules the task directly instead of sending it to itself through `newTaskCh`.
- The cancel callback for a drift skip is invoked asynchronously (off the main loop), consistent with the replace/remove paths, and is tracked so Stop waits for it.
- No public API change.

## Capabilities

### New Capabilities

### Modified Capabilities
- `task-execution`: the "Maximum drift" requirement states the cancel callback is invoked without blocking scheduling of other tasks, and that any number of simultaneous skips cannot stall the scheduler.

## Impact

- Code: `scheduler.go` (`processTask` skip path), tests.
- API: none. Behavioural note: the drift-skip cancel callback now runs on its own goroutine, so it may run concurrently with the task's next run; callers that relied on it completing before rescheduling must synchronise themselves.
- Dependencies: none.
- Depends on: none. Edits `processTask`, also touched by `fix-shared-state-races`; either order works with a rebase. `split-cancel-callback` later renames this callback to a "skipped" callback and must keep the asynchronous invocation.
