## Context

Task state lives in `schedulingUnit.meta` (`scheduler.go:155-175`) plus the pending `timer`. Today the task goroutine writes `meta` (`scheduler.go:440-456`) while the main loop reads it; `fix-shared-state-races` makes the main loop the sole owner of this state. Inspection must read the same state from caller goroutines without reintroducing races.

## Goals / Non-Goals

**Goals:**
- Race-free, copy-based snapshots of every task.
- Expose what the scheduler already tracks; no new bookkeeping beyond "next run time" and "running".

**Non-Goals:**
- Tags (added by `add-task-tags`), metrics/event hooks (`replace-summary-logging-with-hooks`), history of past runs.

## Decisions

- **`TaskInfo` is a plain struct of values** (times as `time.Time`, zero meaning "never"; `LastError error`). Avoids `*time.Time` aliasing seen in `TriggerMetadata`. Alternative: return `TriggerMetadata` — rejected, it lacks next run / running and contains pointers.
- **Snapshots via the main loop while running**: `Task`/`ScheduledTasks` send a request on an inspection channel and the main loop replies with copies built from state it owns. When the scheduler is idle no goroutine mutates task state, so snapshots are built directly from the registry under its lock. Alternative: a per-unit `sync.RWMutex` — rejected as it contradicts the single-owner model from `fix-shared-state-races`.
- **Next run time** is recorded by the main loop when it arms a timer (the computed fire time); cleared when no run is pending (idle scheduler, schedule error).
- **Running flag** set by the main loop when it launches a run and cleared when it receives the completion message.
- **Ordering**: sort by ID for deterministic output (the `sync.Map` range order is random).

## Risks / Trade-offs

- [Inspection blocks while the main loop is busy] → requests are cheap and the main loop never blocks on user code after `fix-mainloop-self-send-deadlock`; requests also select on the scheduler context to avoid hanging during Stop.
- [Breaking return type] → stated in proposal; pre-v1.

## Migration Plan

Callers that used `len(ScheduledTasks())` keep working; callers needing the funcs must keep their own references.
