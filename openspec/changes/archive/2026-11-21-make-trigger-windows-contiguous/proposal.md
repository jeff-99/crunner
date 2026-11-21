## Why

Trigger windows are the core promise of crunner, but they are not contiguous. In `scheduleTask` a window *starts* at the previous window's end, while for non-backfill tasks the next window's *end* is computed from `LastRunTime` (the actual start of the previous run). Any delay between window end and actual start therefore shifts all later windows (drift accumulates), and windows widen or move off the schedule's grid (see the TODO in `scheduleTask`). Separately, a run that takes longer than interval + MaxDrift makes the next window already "late", so the task's own duration causes a drift skip and a spurious cancel callback. After a real drift skip the drift-skip path sets `LastRunTime = now`, so the next window becomes `[skipped end, now + interval)`: off the grid, and the skipped window's own time range is never covered. `schedulingEpoch` is set but unused.

## What Changes

- The next window's end is always computed from the previous window's end (for every task, not only backfill). Windows of a task form an unbroken, non-overlapping sequence on the schedule's grid; actual run start times no longer influence window boundaries.
- The first window starts when the task is first scheduled: at registration on a running scheduler, or at Start for tasks registered while idle (today: registration time, even when Start comes much later).
- The expected trigger time of a run is the later of its window end and the moment the task's previous run finished. Drift is measured against that, so a long-running task never causes its own next run to be cancelled; it runs immediately after the previous run instead.
- When a non-backfill run is more than MaxDrift late (e.g. the process was paused or the scheduler stopped), the run is cancelled: it is not executed and its cancel callback is called once. Its window is not lost: the next run's window starts at the cancelled window's start and ends at the first point of the schedule after the current time, and that run fires at that end as normal. Example (interval 10s, MaxDrift 5s): the timer for [10s, 20s) fires at 83s → callback once, next run covers [10s, 90s) and fires at 90s.
- The same rule applies when the scheduler is restarted after a Stop.
- Remove the unused `schedulingEpoch` field and the `LastRunTime = now` workaround in the drift path.
- The `lastRun` argument passed to `Schedule.Next` is now a previous point of the window sequence (normally the previous window end).

## Capabilities

### New Capabilities

### Modified Capabilities
- `task-execution`: Trigger windows (contiguity, first-window start, multi-step windows after a cancellation), Maximum drift (expected trigger time accounts for the previous run; a cancelled window is merged into the next run's window).
- `schedules`: Pluggable schedules and Interval schedule — the reference time passed to the schedule is a previous window end, never a run start.

## Impact

- Code: `scheduler.go` (`scheduleTask`, `processTask` drift path, `runResult` carries the run's finish time, `Start`, removal of `internalScheduler.schedulingEpoch`), `schedule.go` and `TriggerMetadata` docs, `scheduler_test.go`.
- API: no signature changes. **BREAKING** behaviour:
  - `Schedule.Next`'s `lastRun` argument is a previous window end, not the actual run start. Custom schedules that assumed otherwise are affected.
  - Interval runs no longer drift with run start.
  - The first window of tasks registered before Start begins at Start.
  - A window can span more than one interval (after a drift cancellation); tasks must not assume `TriggerWindowEnd - TriggerWindowStart == interval`.
  - After a drift cancellation, the next run fires at its normal grid time, not at `now + interval`.
  - `TriggerMetadata.LastRunTime` keeps meaning "start of last executed run" and is no longer set when a run is cancelled.
- Dependencies on other changes: `fix-mainloop-self-send-deadlock` and `fix-shared-state-races` (both merged). `add-overlap-policy` later builds on the "run immediately after a long run" behaviour defined here; `add-schedule-jitter`, `add-state-store` and `add-run-now` rely on contiguous windows. `split-cancel-callback` should pass the cancelled window to its callback.
