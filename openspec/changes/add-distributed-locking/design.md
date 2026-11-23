# Design

## Context

A run currently goes: timer fires → main loop `processTask` checks drift (`scheduler.go` ~401) → goroutine runs the task, updates meta, invokes callbacks, sends the unit back on `newTaskCh` (~434-466). `add-state-store` adds a `Save` after the run and a `Load` before the first window. Window ends are derived from the previous window end once `make-trigger-windows-contiguous` lands.

## Goals / Non-Goals

**Goals:**
- At most one concurrent execution per (task, window) across instances sharing a locker.
- Let one backend (e.g. a user-written Postgres implementation using a transaction or advisory lock) implement locker and state store together so that "lock window X, run, record X done" is atomic.
- Keep the no-locker path identical to today.

**Non-Goals:**
- Leader election / a single active scheduler.
- Fencing tokens or detecting a lock lost mid-run (a backend can cancel work itself; see Open Questions).
- Shipping any backend.

## Decisions

**Interfaces.**
```go
var ErrLockHeld = errors.New("crunner: lock held by another holder")
type Locker interface { Lock(ctx context.Context, key string) (Lock, error) }
type Lock   interface { Unlock(ctx context.Context) error }
func WithLocker(Locker) SchedulerOption
```
`Lock` is non-blocking from the scheduler's point of view: a locker must return `ErrLockHeld` (wrapped is fine, checked with `errors.Is`) rather than wait. Blocking until the other replica finishes would then run the same window again.

**Optional transactional hook.** If the returned `Lock` implements `interface{ Save(ctx, taskID string, State) error }` (the same method set as `StateStore.Save`), state is saved via the lock before `Unlock`. A Postgres backend can open a transaction in `Lock`, take `pg_advisory_xact_lock` / `SELECT … FOR UPDATE`, write state in `Save` and commit in `Unlock`. Alternative: a combined `LockingStateStore` interface — rejected; type-asserting the lock keeps the two options independent and needs no extra configuration.

**Key format.** `crunner/<taskID>/<windowEnd UTC, RFC3339Nano>`. Using the window end (not start) means a replica whose first window started at a different registration time still collides with the others once their window ends agree. Readable keys help debugging; backends needing integers hash them.

**Sequence in the task goroutine.** drift check (main loop) → `Lock` → if state store configured: `Load`, skip if `WindowEnd >= current end` → run → callbacks → `Save` (via lock or store) → `Unlock` → reschedule. Locking happens in the task goroutine so backend latency never blocks the main loop. Drift is checked before locking; lock latency does not count as drift.

**Contention and already-handled are silent.** They are the normal steady state on all but one replica; invoking cancel/error callbacks would be noise. They are logged at debug level (and become a "skip" event once `replace-summary-logging-with-hooks` lands, which is not a dependency).

**Realignment.** Interval tasks registered at different times on different replicas have different first windows. The already-handled check moves a lagging replica's next window start to the recorded end, so replicas converge after one cycle. Without a state store there is no realignment; the README states that deterministic window ends (cron schedules, or backfill from a common `NotBefore`) are then required.

**Lock errors count as run errors.** The user must learn that their coordination backend is failing, and the error callback is the existing channel. The window is not retried, matching how failed runs are treated.

## Risks / Trade-offs

- [Lock expires while a long task still runs (TTL-based backends) → second replica runs the same window] → documented; the already-handled check prevents it only after the first run saved. Backends should size TTLs or renew.
- [Replicas with skewed clocks compute the "same" window at different moments] → lock still excludes concurrent runs; the late replica hits the already-handled check when a state store is configured.
- [No state store + interval schedule anchored at registration time → keys never collide] → documented; recommend configuring a state store whenever a locker is configured.

## Migration Plan

Additive. Rolling out to an existing multi-replica deployment: configure the same locker (and state store) on all replicas; until all are upgraded, old replicas keep running every window.

## Open Questions

- Whether to expose the lock's lifetime to the task (e.g. a lock-scoped context) can be added later as another optional interface on `Lock` without changing this design.
