## Why

Trigger windows are the core promise of crunner, but they are not contiguous. In `scheduleTask` (scheduler.go ~298-338) a window *starts* at the previous window's end, while for non-backfill tasks the next window's *end* is computed from `LastRunTime` (the actual start of the previous run). Any delay between window end and actual start therefore shifts all later windows (drift accumulates), and windows can overlap or leave gaps (see the TODO at ~300). Separately, a run that takes longer than interval + MaxDrift makes the next window already "late", so the task's own duration causes a drift skip and a spurious cancel callback (finding 7). The drift-skip path papers over this by setting `LastRunTime = now` (~406). `schedulingEpoch` (~147, ~574) is set but unused.

## What Changes

- The next window's end is always computed from the previous window's end (for every task, not only backfill). Windows of a task form an unbroken, non-overlapping sequence on the schedule's grid; actual run start times no longer influence window boundaries.
- The first window starts when the task is first scheduled: at registration on a running scheduler, or at Start for tasks registered while idle (today: registration time, even when Start comes much later).
- The expected trigger time of a run is the later of its window end and the moment the task's previous run finished. Drift is measured against that, so a long-running task never causes its own next run to be skipped; it runs immediately after the previous run instead.
- When a non-backfill run is skipped for drift (e.g. the process was paused), all windows that already ended more than MaxDrift ago are skipped together with a single cancel callback, and scheduling resumes with the first window that has not; windows stay contiguous across the skip.
- Remove the unused `schedulingEpoch` field and the `LastRunTime = now` workaround in the skip path.
- The `lastRun` argument passed to `Schedule.Next` is now the previous window end. **BREAKING** (behavioural) for custom `Schedule` implementations that assumed it was the actual run start; the built-in interval schedule now yields `previous window end + interval`.

## Capabilities

### New Capabilities

### Modified Capabilities
- `task-execution`: Trigger windows (contiguity, first-window start), Maximum drift (expected trigger time accounts for the previous run; grouped skip).
- `schedules`: Pluggable schedules and Interval schedule — the reference time passed to the schedule is the previous window end.

## Impact

- Code: `scheduler.go` (`scheduleTask`, `processTask` skip path, task goroutine records finish time, `Start`, `internalScheduler.schedulingEpoch`), `schedule.go` docs, `scheduler_test.go`.
- API: **BREAKING** behaviour — meaning of `Schedule.Next`'s `lastRun` argument; interval runs no longer drift with run start; first window of tasks registered before Start begins at Start. No signature changes. `TriggerMetadata.LastRunTime` keeps meaning "start of last executed run".
- Dependencies on other changes: depends on `fix-mainloop-self-send-deadlock` (the skip path sends to `newTaskCh` from the main loop; grouped skips must not make that worse). Should land after or together with `fix-shared-state-races` since it reads the previous run's finish time written by the task goroutine. `add-overlap-policy` later builds on the "run immediately after a long run" behaviour defined here; `add-schedule-jitter`, `add-state-store` and `add-run-now` rely on contiguous windows.
