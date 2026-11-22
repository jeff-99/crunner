## Context

Task state lives in `schedulingUnit.meta` plus the pending `timer`, and is owned by the main loop: it is mutated in `scheduleWindow`, `processTask` and `handleRunResult`, all on the main loop. Before `Start` (and for a task being registered) the caller goroutine is the only one touching a unit. Inspection must read this state from arbitrary caller goroutines without reintroducing races, including during the `Start` and `Stop` transitions.

## Goals / Non-Goals

**Goals:**
- Race-free, copy-based snapshots of every task at any time (idle, starting, running, stopping, stopped).
- Expose what the scheduler already tracks; new bookkeeping only for "next run time" and "running".

**Non-Goals:**
- Tags (`add-task-tags`), metrics/event hooks (`replace-summary-logging-with-hooks`), history of past runs, counters for skipped runs.

## Decisions

- **Publish immutable snapshots instead of asking the main loop.** Each `schedulingUnit` holds an `atomic.Pointer[TaskInfo]`. The writer builds a fresh `TaskInfo` after every state change and stores it; `Task`/`ScheduledTasks` only load. The writer is always a single goroutine: `RegisterTask` before the unit is stored in the registry, then the main loop. Alternatives rejected: a request/reply channel to the main loop with a direct read when idle (racy around `Start`, blocks while the loop is busy or gone); a per-unit mutex (contradicts the single-owner model).
- **`TaskInfo` is a plain struct of values** (times as `time.Time`, zero meaning "never"; `LastError error`) named like `TriggerMetadata`, avoiding the `*time.Time` aliasing there. Alternative: return `TriggerMetadata` — rejected, it lacks next run / running and contains pointers.
- **Publish points** (all on the main loop except the first): `RegisterTask` (initial zero-count snapshot), `scheduleWindow` when a timer is armed (sets `NextRunTime` to the moment it fires and the upcoming window) or when scheduling fails (clears `NextRunTime`), `processTask` when a run is launched (`Running = true`, `NextRunTime` cleared), `handleRunResult` (counts, times, last error, `Running = false`), and main loop exit (clears `NextRunTime` and `Running`, since timers are stopped).
- **Next run time** is when the armed timer fires: the expected trigger time, or now when that is in the past or for backfill.
- **Ordering**: `ScheduledTasks` sorts by ID for deterministic output (the registry range order is random). A listing is a set of per-task snapshots, not one atomic view of all tasks.
- **Generated IDs become discoverable** through `ScheduledTasks`, closing the gap left by `require-stable-task-ids`.

## Risks / Trade-offs

- [A snapshot can lag the true state by one state change] → Each snapshot is internally consistent and replaced atomically; the lag is bounded by the main loop's processing of a single message.
- [Allocation of one `TaskInfo` per state change] → A handful per run; negligible next to a run.
- [Breaking return type and interface method] → Stated in the proposal; pre-v1.

## Migration Plan

Callers that used `len(ScheduledTasks())` keep working; callers needing the funcs must keep their own references. Callers that need generated IDs read them from `ScheduledTasks`.
