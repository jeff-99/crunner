## Context

`scheduleTask` (scheduler.go ~279-389) derives `triggerWindowStart` from `meta.TriggerWindowEnd` but passes `meta.LastRunTime` (actual start, set in the task goroutine ~440) as `lastRun` to `Schedule.Next` unless backfilling (~316-325). `expectedTriggerTime` is the window end; `processTask` (~398) skips when `now - expected > MaxDrift`, and on skip sets `LastRunTime = now` (~406) so the next computation moves forward. `schedulingEpoch` is assigned in `Start` (~574) and never read; the first window uses `su.createdAt`. Backfill already uses window ends (~319-320) and is the behaviour we generalise.

## Goals / Non-Goals

**Goals:**
- One rule for window boundaries for all tasks: `end_n = schedule.Next(start, end_{n-1})`.
- A task's own runtime never triggers a drift skip.
- Bounded work after a long pause (no skip storm).

**Non-Goals:**
- Overlap policies (skip/reschedule when a run outlasts its window) — `add-overlap-policy`. Here the behaviour stays "wait, then run immediately".
- Persisting window state across process restarts — `add-state-store`.
- Changing backfill semantics beyond using the shared rule.

## Decisions

- **Always pass previous window end to `Schedule.Next`.** Removes the backfill/non-backfill branch for `lastRunPtr`. Alternative: keep `LastRunTime` and round to the grid — rejected, only works for intervals, not for custom schedules.
- **First window starts at "first scheduled" time.** Record a `windowOrigin` on the unit when it is first scheduled while running (registration on a running scheduler, or Start). Replaces both `createdAt` as base and the unused `schedulingEpoch`, which is deleted. Alternative: use `schedulingEpoch` for all tasks — rejected, tasks registered after Start would get a first window starting before they existed. On a Stop/Start restart, the unit keeps its previous window end, so windows continue contiguously; non-backfill tasks then pass through the grouped-skip rule.
- **Expected trigger time = max(window end, previous run finish).** The task goroutine records the finish time on the unit (handed to the main loop with the reschedule message, consistent with `fix-shared-state-races` ownership). `fireAfter` uses the same value. Alternative: increase allowed drift by run duration — rejected, harder to reason about and leaks into jitter rules.
- **Grouped skip for non-backfill tasks.** In the skip path, advance windows in the main loop (`end = Next(end)`) while `end < now - MaxDrift`, mark the covered range as skipped, call the cancel callback once, then schedule the first remaining window. Loop is bounded by elapsed time / smallest step; `Schedule.Next` errors abort scheduling as today. Alternative: one skip per window — rejected, a 1s interval task after a 1h pause would emit 3600 callbacks. Backfilling tasks don't skip (they always run immediately with `fireAfter = 0`).
- **Remove `LastRunTime = now` on skip.** `LastRunTime` then only reflects executed runs, matching its name.

## Risks / Trade-offs

- [Custom schedules relying on actual run start in `lastRun`] → Called out as BREAKING; documented on the `Schedule` interface.
- [A task that is consistently slower than its interval falls further behind forever, running back-to-back] → Visible via window timestamps; `add-overlap-policy` adds skip/reschedule policies. Documented.
- [Grouped skip hides how many windows were skipped from the cancel callback] → Logged with the skipped range; `split-cancel-callback` can pass window info.
- [Existing tests in `TestTaskTriggerWindows*` assume zero-duration runs] → They remain valid; new tests cover delayed starts.
