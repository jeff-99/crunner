## Context

`SchedulingOptions.NotAfter` exists but is unused. `scheduleTask` computes `triggerWindowEnd` from `Schedule.Next` (~331); runs are executed in a goroutine that increments `meta.Count` and re-sends the unit on `newTaskCh` (~440-456). `RegisterTask` currently always returns nil.

## Goals / Non-Goals

**Goals:**
- Make NotAfter effective; add MaxRuns; remove finished tasks.

**Non-Goals:**
- Cancelling an in-flight run when NotAfter passes (a run that started for a window ending ≤ NotAfter finishes normally).
- Per-window "last run" signal in metadata.

## Decisions

- **NotAfter compares against the window end** (the scheduled trigger time), inclusive. Window end is what identifies a run in crunner's model, and it is deterministic; actual start time is not. Alternative: compare actual start time — rejected, jitter or drift would make the last run nondeterministic.
- **Check in `scheduleTask`, before arming the timer.** If `triggerWindowEnd.After(*NotAfter)`, complete the task instead of scheduling. This covers backfill too, since backfill goes through the same path.
- **MaxRuns counts executions, checked when the run returns** (in the post-run path that today re-sends to `newTaskCh`): if `Count >= MaxRuns`, complete instead of rescheduling. Counting executions rather than triggers means a drift skip does not consume budget, matching user intent of "N actual runs".
- **Completion = `CompareAndDelete` from `s.tasks` without cancel callback**, executed on the main loop so it cannot remove a replacement registered under the same ID. Same helper as `add-one-time-schedule`.
- **Validation in `RegisterTask`** returns an error before touching `s.tasks`, so an invalid re-registration does not cancel the existing task.

## Risks / Trade-offs

- [Users passing `WithNotAfter` today without expecting an effect] → called out as behaviour change in the proposal and changelog.
- [Post-run completion touches state from the task goroutine] → route through the main loop (send a completion request on the existing channel) rather than deleting from the goroutine; `fix-shared-state-races` will make this the norm.
