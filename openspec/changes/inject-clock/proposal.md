# Proposal

## Why

The scheduler reads wall-clock time and arms timers directly through the `time` package (`time.Now()` in `RegisterTask`, `scheduleTask`, `processTask` and the task goroutine; `time.AfterFunc` in `scheduleTask`, scheduler.go ~219/292/372/396/438/371). Applications embedding crunner cannot control time, so they cannot deterministically test code that depends on windows, backfill or drift skips, nor simulate clock jumps.

## What Changes

- Add an exported `Clock` interface (current time + one-shot timer) and an exported `Timer` interface it returns.
- Add a scheduler option `WithClock(Clock)`; the default is a clock backed by the `time` package, so behaviour without the option is unchanged.
- All time reads and timer creation in the scheduler go through the configured clock: registration time, window computation, trigger timers, drift check and run start times passed to tasks.
- No breaking API changes; purely additive.

## Capabilities

### New Capabilities

### Modified Capabilities
- `scheduler-lifecycle`: adds a requirement that the scheduler's notion of time comes from an injectable clock, defaulting to the real clock.

## Impact

- Code: `scheduler.go` (new option, `internalScheduler` field, replace direct `time.Now`/`time.AfterFunc` calls; `schedulingUnit.timer` becomes the `Timer` interface), new `clock.go`.
- API: additive — `Clock`, `Timer`, `WithClock`. No breaking change.
- Dependencies: none added.
- Related changes: no hard dependency. Overlaps textually with `fix-shared-state-races` (timer ownership), `make-trigger-windows-contiguous` and `replace-summary-logging-with-hooks` (summary ticker); whichever lands later rebases. `make-timeouts-configurable` and `add-stop-timeout` keep using context deadlines, which stay on real time (see design).
