## Context

Runs are triggered by a `time.AfterFunc` timer that sends the unit on `tasksIn`; the main loop calls `processTask`, which checks drift and runs the task in a goroutine, which on return updates `meta` and re-sends to `newTaskCh` for the next window (`scheduler.go` ~384-456). Nothing tracks whether a unit is currently running: non-overlap is a consequence of only scheduling the next window after the run returns.

## Goals / Non-Goals

**Goals:**
- Immediate, out-of-band runs with a well-defined window.
- Preserve scheduled window continuity and non-overlap.

**Non-Goals:**
- Queueing multiple manual requests (a second request while running is rejected).
- Overlap policies. What a scheduled trigger does while a run is executing is governed by MaxDrift (generous: runs back to back; small: skipped and merged).
- Run-now for tasks registered while the scheduler is idle (returns `ErrSchedulerNotRunning`).

## Decisions

- **Window = [last completed scheduled window end, request time).** Alternatives considered: (a) run the pending window early, truncated at now — rejected, it shortens or shifts the scheduled windows that downstream consumers rely on; (b) zero-length/no window — rejected, tasks that query by window would do nothing. Overlap between the manual window and the following scheduled window is intentional and flagged by `Manual: true`, so idempotent tasks can reprocess safely.
- **Main loop owns the request.** `RunNow` sends `{id, reply chan error}` on a new `runNowCh`; the main loop looks up the unit, checks `running`, builds a metadata copy with the manual window and starts the goroutine; the reply carries `ErrTaskNotFound`/`ErrTaskRunning`. This keeps all decisions on the single owner of task state (`fix-shared-state-races`).
- **Per-unit `running` flag** set when any run starts and cleared on return (via the main loop). When a timer delivers a unit on `tasksIn` while `running`, the main loop parks it as `pendingTrigger` and processes it when the run returns; drift is measured at that moment, as specified.
- **After a manual run, do not reschedule.** The scheduled timer is untouched; only `Count`, success/error counts, `LastRunTime` etc. are updated. To avoid `LastRunTime` shifting the next interval window today, the interval computation must use window end (`make-trigger-windows-contiguous`); until that lands, manual runs update statistics but not the field used by `Schedule.Next` (keep a separate scheduled-last-run value).

## Risks / Trade-offs

- [Interface method addition breaks mocks] → documented as BREAKING; acceptable pre-v1.
- [Manual window overlaps next scheduled window] → explicit `Manual` flag and documentation.
- [RunNow blocks if the main loop is busy] → the request respects the scheduler context and returns `ErrSchedulerNotRunning` when it is cancelled.
