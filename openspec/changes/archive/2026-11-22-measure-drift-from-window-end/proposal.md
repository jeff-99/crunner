## Why

`make-trigger-windows-contiguous` made the expected trigger time of a run `max(window end, previous run's finish)`. The effect is that waiting for a previous run never counts as drift: a task whose runs overrun its interval is never skipped, runs back to back, and its windows fall further behind the schedule forever. Nothing signals this; no skip callback fires and drift stays at zero.

The intended behaviour is the one the scheduler had before that change and that the overlap-policy proposal assumes: drift is measured from the scheduled time of the window. An overrun that leaves the next run more than MaxDrift late is skipped, with the skip callback, and its window is merged into the next run. Lag is then bounded by MaxDrift and every missed run is visible.

Example (interval 10s, MaxDrift 5s, runs taking 7.5s, 12.5s, 1s, 19s): the run for window [20,30) starts at 32.5s (2.5s late, within MaxDrift) and the run for window [40,50) would start at 59s (9s late), so it is skipped and the next run covers [40,60) at 60s.

## What Changes

- The expected trigger time of a run is its window end again; the previous run's finish no longer moves it.
- A run that would start more than MaxDrift after its window end is skipped (skip callback once, window merged into the next run), including when the delay is caused by the task's own previous run.
- A run that is late by at most MaxDrift starts immediately, as before.
- Removes the `previousFinish` state and the finish time carried in `runResult`, which only existed for the old rule.
- Documentation of `WithMaxDrift` says that overruns count as drift.
- **BREAKING** (behavioural): tasks whose runs overrun their interval by more than MaxDrift now get skipped runs (and skip callbacks) instead of silently falling behind. No signature changes.

## Capabilities

### New Capabilities

### Modified Capabilities
- `task-execution`: "Drift skipping" measures drift from the window end and no longer exempts overruns; the long-running-task scenario is split into within-MaxDrift and beyond-MaxDrift cases.

## Impact

- Code: `scheduler.go` (`scheduleWindow` expected trigger time; remove `schedulingUnit.previousFinish`, `runResult.finish` and their updates in the task goroutine and `handleRunResult`; `WithMaxDrift` doc), `scheduler_test.go`.
- API: no signature changes. **BREAKING** behaviour as described above.
- Dependencies on other changes: none outstanding. It reverses one decision of `make-trigger-windows-contiguous` (merged) and should land before `add-overlap-policy`, whose Wait policy assumes it.
