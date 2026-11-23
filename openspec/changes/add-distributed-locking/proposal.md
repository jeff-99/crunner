# Proposal

## Why

crunner runs in a single process only. When an application is deployed with several replicas, every replica runs every task for every window. Users need a way to ensure that, for a given task and trigger window, only one replica executes the run — without crunner taking a dependency on any particular coordination backend.

## What Changes

- Add exported `Locker` (`Lock(ctx, key string) (Lock, error)`) and `Lock` (`Unlock(ctx) error`) interfaces and a sentinel `ErrLockHeld` that a locker returns when another holder owns the key.
- Add a scheduler option `WithLocker(Locker)`. The default is a no-op locker that always succeeds, so behaviour without the option is unchanged.
- Before executing a run, the scheduler acquires a lock keyed by task ID and trigger window end; it releases it after the run and after the window state has been saved.
- When the lock is held elsewhere, the run is skipped silently (no callbacks) and the next window is scheduled.
- With a state store configured, the scheduler re-reads the task state while holding the lock and skips the window if another replica already handled it, realigning its next window to the recorded end.
- If the acquired `Lock` also implements `StateStore`'s `Save` method, state is saved through the lock so a backend can make "lock window, run, record done" transactional.
- Lock errors other than `ErrLockHeld` are reported to the task's error callback and the run is not executed.

## Capabilities

### New Capabilities
- `distributed-locking`: the locker abstraction, lock keys, the lock / run / record / unlock sequence, contention and error behaviour, and the transactional save hook.

### Modified Capabilities
<!-- none: without a configured locker behaviour is unchanged; all new behaviour is in distributed-locking -->

## Impact

- Code: `scheduler.go` (task goroutine started by `processTask`), new file `lock.go`.
- API: additive only — `Locker`, `Lock`, `ErrLockHeld`, `WithLocker`. No breaking change.
- Dependencies: no new modules; no backend shipped in this repo.
- Depends on: `add-state-store` (already-handled check and the transactional save hook use `State`/`StateStore`), `require-stable-task-ids` (lock keys must be identical across replicas), `make-trigger-windows-contiguous` (replicas must derive identical windows from the previous window end), `fix-shared-state-races` (lock handling runs in the task goroutine and updates window state).
