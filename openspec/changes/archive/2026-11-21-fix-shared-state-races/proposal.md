## Why

The scheduler has data races on per-task state (review finding 1). `su.timer` is written by the main loop in `scheduleTask` (`scheduler.go:388`) without holding `stateMu`, while `RegisterTask` (~236), `RemoveTask` (~656) and `Stop` (~615) read and stop it under `stateMu` from other goroutines. Task metadata (`LastRunTime`, counts, `LastError`) is written by the task goroutine (`scheduler.go:440-456`) while the main loop reads it in `emitTaskSummaries` (~472) and `scheduleTask`. `go test -race` can flag these, and the results are undefined.

## What Changes

- The main loop becomes the sole owner of mutable per-task state (timer, trigger metadata, backfill flag, summary counters). No other goroutine reads or writes it.
- The task goroutine no longer mutates metadata; it reports the run's outcome (start time, error) back to the main loop, which updates metadata, invokes nothing else, and reschedules.
- `RegisterTask`, `RemoveTask` and `Stop` no longer touch timers directly. Stale timers of replaced/removed tasks are neutralised by the main loop's existing "is this still the registered unit" check, and the main loop stops all pending timers when it exits; `Stop` waits for the main loop to exit.
- `stateMu` is removed.
- No public API change.

## Capabilities

### New Capabilities

### Modified Capabilities
- `scheduler-lifecycle`: adds a requirement that the scheduler is safe for concurrent use (race-free under the Go race detector).

## Impact

- Code: `scheduler.go` (`scheduleTask`, `processTask`, `mainLoop`, `RegisterTask`, `RemoveTask`, `Stop`, `emitTaskSummaries`), new internal completion channel/message type, tests.
- API: none (no breaking change).
- Dependencies: none.
- Depends on: none. Touches the same functions as `fix-mainloop-self-send-deadlock` and `make-task-registration-atomic`; whichever lands second rebases. Later changes (`add-task-inspection`, `replace-summary-logging-with-hooks`) build on the single-owner model introduced here.
