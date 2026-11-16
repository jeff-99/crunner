## Context

`scheduleTask` (`scheduler.go` ~331) calls `Schedule.Next`; on any error it logs at error level and returns, leaving the unit in `s.tasks` but never rescheduled. Runs are rescheduled by sending the unit to `newTaskCh` after the run (~454) or after a drift skip (~413), so "is there another run?" is naturally answered by the next `Next` call.

## Goals / Non-Goals

**Goals:**
- One-time schedule with the same window semantics as other schedules.
- A general, schedule-level way to end a task, reusable by user schedules.

**Non-Goals:**
- Changing the meaning of other schedule errors.
- Run limits on recurring schedules (`add-run-limits`).
- Keeping completed tasks visible for inspection (`add-task-inspection` may revisit).

## Decisions

- **`At(t)` returns `t` when `lastRun == nil`, else `ErrNoMoreRuns`.** Keeps the existing `Schedule` interface; no new interface method. Alternative: an optional `Finite` interface — rejected, more surface for the same information.
- **After a drift skip the one-time schedule is also exhausted.** Today the skip path sets `LastRunTime` before rescheduling, so the next `Next` call sees `lastRun != nil` and returns `ErrNoMoreRuns`. A missed one-time run is therefore not retried; users wanting "run whenever possible" set a large MaxDrift.
- **Completion removes the task from `s.tasks` in the main loop, without the cancel callback.** The cancel callback means "removed/replaced/skipped" and should not fire on natural completion. Alternative: keep the unit registered but idle — rejected, it leaks units and keeps a stale ID occupied.
- **Removal is checked against the current unit** (pointer equality, as the main loop already does) so a replacement registered under the same ID is never removed.
- **Backfill:** with backfill and NotBefore, the single window is [NotBefore, t); it runs immediately if t is past. No special case needed.

## Risks / Trade-offs

- [Removal races with `RegisterTask`/`RemoveTask` touching `s.tasks` concurrently] → use `CompareAndDelete(id, su)` on the `sync.Map`; becomes trivial once `fix-shared-state-races` makes the main loop sole owner.
- [Users relying on "any schedule error keeps task registered"] → only the new sentinel changes behaviour.
