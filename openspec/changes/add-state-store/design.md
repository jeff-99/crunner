# Design

## Context

Window state is held in `schedulingUnit.meta` (`scheduler.go` ~141). `scheduleTask` (~281) derives the first window start from `su.createdAt` or `NotBefore`, and later windows from `meta.TriggerWindowEnd`. Runs finish in a goroutine started by `processTask` (~434), which updates `meta` and sends the unit back on `newTaskCh`. The drift-skip path (~405) runs on the main loop today; `fix-mainloop-self-send-deadlock` moves it off.

## Goals / Non-Goals

**Goals:**
- Minimal, backend-agnostic interface that a user can implement on any database.
- Shape the interface so that a backend can also implement the `Locker` from `add-distributed-locking` and save state in the same transaction as holding the window lock.
- No store I/O on the main loop.

**Non-Goals:**
- Shipping persistent backends (file, SQL) in this repo.
- Persisting run counts, last error or other statistics.
- Retrying failed windows; failed runs stay "handled" like today.

## Decisions

**Interface shape.**
```go
type State struct {
    WindowStart time.Time
    WindowEnd   time.Time
}
type StateStore interface {
    Load(ctx context.Context, taskID string) (State, bool, error)
    Save(ctx context.Context, taskID string, state State) error
}
func NewMemoryStateStore() StateStore
func WithStateStore(StateStore) SchedulerOption
```
A struct (not individual fields in the interface) lets fields be added later without breaking implementations. Alternative: store the full `TriggerMetadata` — rejected, it contains `error` values and counts that do not serialise cleanly and are not needed to resume.

**What counts as handled.** Success, error and drift-skip all mark a window handled, matching today's in-memory behaviour where none of them is retried. Recording only successes would turn a failing task into an endless backfill loop after restart.

**When to load.** Once per scheduling unit, before its first window is computed, in a goroutine started when the unit is queued for first scheduling (from `Start` or `RegisterTask`); the loaded state is handed to the main loop with the unit. Re-registering a task with the same ID loads again. Alternative: load on every window — rejected, unnecessary I/O; the in-memory meta is authoritative while the process runs (the distributed-locking change re-reads under the lock where it matters).

**When to save.** In the task goroutine after the run returns and callbacks are invoked, before sending the unit for rescheduling; on the drift-skip path in the goroutine that handles the skip. A per-call context derived from the scheduler context is used, so `Stop` cancels pending store calls.

**Load errors fail closed.** Starting from the registration time after a transient store error would silently lose coverage, and starting from `NotBefore` would replay windows. Treating it like a schedule error is the existing, documented failure mode. Save errors only log: the in-memory state is still correct, and the next successful save overwrites.

**Resume point with backfill.** `max(NotBefore, state.WindowEnd)`, so moving `NotBefore` forward still works.

## Risks / Trade-offs

- [A task that fails to load is never scheduled until re-registration or restart] → logged at error level; documented. Users wanting resilience can make `Load` retry internally.
- [Crash between run return and `Save` re-runs that window after restart] → at-least-once semantics, documented on `StateStore`. Exactly-once needs the transactional backend described in `add-distributed-locking`.
- [Without backfill, a long downtime yields one large first window] → intended: the window still tells the task exactly what period it covers; users wanting per-interval windows enable backfill.

## Migration Plan

Purely additive; existing users get the in-memory store. No rollback concerns.
