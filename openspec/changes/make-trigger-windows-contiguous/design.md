## Context

`scheduleTask` derives `triggerWindowStart` from `meta.TriggerWindowEnd` but passes `meta.LastRunTime` (actual run start, recorded by the main loop from the run's `runResult`) as `lastRun` to `Schedule.Next` unless backfilling. `expectedTriggerTime` is the window end; `processTask` cancels the run when `now - expected > MaxDrift`, and on that path sets `LastRunTime = now`, starts the cancel callback in a goroutine tracked by `s.wg`, and calls `scheduleTask` directly (no longer via `newTaskCh`, since `fix-mainloop-self-send-deadlock`). `schedulingEpoch` is assigned in `Start` and never read; the first window uses `su.createdAt`. Backfill already uses window ends and is the behaviour we generalise. Per-task state is owned by the main loop (`fix-shared-state-races`).

## Goals / Non-Goals

**Goals:**
- One rule for window boundaries for all tasks: `end_n = schedule.Next(start, end_{n-1})`, windows contiguous and on the schedule's grid.
- A task's own runtime never causes a drift cancellation.
- A cancelled run never leaves a gap: its window is covered by the next run.

**Non-Goals:**
- Overlap policies (skip/reschedule when a run outlasts its window) — `add-overlap-policy`. Here the behaviour stays "wait, then run immediately".
- Persisting window state across process restarts — `add-state-store`.
- Passing the cancelled window to the callback — `split-cancel-callback`.
- Changing backfill semantics beyond using the shared rule.

## Decisions

- **Track where coverage ends, separately from the pending window.** Each unit keeps `coveredUntil`: the end of its last executed window (initially the window origin, or NotBefore when backfilling). Every scheduled window starts at `coveredUntil`. A normal run advances `coveredUntil` to its window end when its result reaches the main loop; a cancelled run does not. This makes "merge the cancelled window into the next one" and "restart after Stop" fall out of the same rule.
- **Always pass a previous window end to `Schedule.Next`.** Removes the backfill/non-backfill branch for `lastRunPtr`. Alternative: keep `LastRunTime` and round to the grid — rejected, only works for intervals, not for custom schedules.
- **First window starts at "first scheduled" time.** Record a window origin on the unit when it is first scheduled while running (registration on a running scheduler, or Start). Replaces both `createdAt` as base and the unused `schedulingEpoch`, which is deleted. Alternative: use `schedulingEpoch` for all tasks — rejected, tasks registered after Start would get a first window starting before they existed.
- **Expected trigger time = max(window end, previous run finish).** The task goroutine puts the run's finish time in its `runResult`; the main loop stores it on the unit. `fireAfter` uses the same value. Alternative: increase allowed drift by run duration — rejected, harder to reason about and leaks into jitter rules.
- **Cancel and merge on drift (non-backfill).** When `now - expected > MaxDrift`: do not run, start the cancel callback (async, tracked by `s.wg`, as today), leave `coveredUntil` unchanged, and compute the next window as `[coveredUntil, e)` where `e` is found by advancing `e = Next(start, e)` from the cancelled window's end until `e > now`. Schedule the timer for `e`; the run fires there as normal. Exactly one callback per cancellation, no immediate catch-up run. Alternatives rejected: drop the cancelled windows (time range never covered); run immediately with a merged window (MaxDrift would no longer prevent a late run); today's `[end, now + interval)` (off the grid).
- **Restart follows the same rule.** On Start, a unit whose `coveredUntil` is set continues from it: its pending window is recomputed from `coveredUntil`, it is late, so it is cancelled and merged as above. Backfilling tasks are never cancelled (they run immediately with `fireAfter = 0`).
- **Remove `LastRunTime = now` on cancellation.** `LastRunTime` then only reflects executed runs, matching its name.

## Risks / Trade-offs

- [Custom schedules relying on actual run start in `lastRun`] → Called out as BREAKING; documented on the `Schedule` interface and in the PR.
- [Windows wider than one interval after a cancellation] → Documented on `TriggerMetadata`; tasks must use `TriggerWindowStart/End`, not assume the interval.
- [A task that is consistently slower than its interval falls further behind forever, running back-to-back] → Accepted; visible via window timestamps. `add-overlap-policy` adds skip/reschedule policies.
- [Finding the merged window end calls `Schedule.Next` once per elapsed step on the main loop (e.g. 1ms interval after a 1h pause ≈ 3.6M calls)] → Accepted for now; no cap.
- [The callback is still called "cancel" and receives no window info] → `split-cancel-callback` renames it and passes the cancelled window.
- [Existing drift tests assume the old skip semantics] → Updated as part of this change.
