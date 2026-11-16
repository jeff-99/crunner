## Why

`WithNotAfter` (`scheduler.go` ~67) is accepted and stored in `SchedulingOptions.NotAfter` but never read: tasks keep running past it. There is also no way to say "run this at most N times" (gocron v2 `WithLimitedRuns`).

## What Changes

- Implement `NotAfter`: no run whose trigger window ends after NotAfter is executed; the task completes instead. Backfill also stops at NotAfter.
- Add `WithMaxRuns(n int)`: the task completes after n executed runs (successful or failed). Runs skipped for drift do not count.
- A task that reaches either limit is removed from the scheduler without its cancel callback.
- `RegisterTask` returns an error when NotAfter is not after NotBefore or when MaxRuns is not positive. **BREAKING (behavioural)**: these registrations previously succeeded silently.

## Capabilities

### New Capabilities

### Modified Capabilities
- `task-registration`: validation of the NotAfter and MaxRuns options.
- `task-execution`: run limits (NotAfter, MaxRuns) and completion of limited tasks.

## Impact

- `scheduler.go`: `SchedulingOptions` gains `MaxRuns`; `RegisterTask` validates; `scheduleTask` checks NotAfter against the computed window end; the post-run path checks the run count.
- API: additive `WithMaxRuns`; `WithNotAfter` starts having an effect (behaviour change for anyone passing it today).
- Dependencies: none required. Shares the "remove completed task without cancel callback" path with `add-one-time-schedule`; whichever lands second reuses the helper. If `validate-task-registration` lands first, put the new checks in its validation function.
