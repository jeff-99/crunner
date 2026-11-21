## Context

See proposal.md. The registry is a `sync.Map` keyed by task ID. `sync.Map` already provides atomic `Swap` (returns the previous value) and `LoadAndDelete` (Go 1.20+); the module targets Go 1.26.

## Goals / Non-Goals

**Goals:**
- Exactly-once cancellation of every displaced or removed task, under any interleaving.

**Non-Goals:**
- Stable/empty ID handling (`require-stable-task-ids`), input validation (`validate-task-registration`), timer ownership (`fix-shared-state-races`).

## Decisions

- **`RegisterTask` uses `previous, loaded := s.tasks.Swap(id, newUnit)`** and runs the existing overwrite handling (stop old timer, async cancel callback) on `previous` only if `loaded`. Each Swap displaces a distinct unit, so each displaced unit is cancelled once. Alternative: a registry mutex around Load+Store — rejected as it adds a lock the `sync.Map` makes unnecessary and would be held while touching timers.
- **`RemoveTask` uses `s.tasks.LoadAndDelete(id)`**; not loaded → not-found error. 
- **Compute `id := newSchedulingUnit.ID()` once per call** and use it for Swap, logging and messages, so the key used for lookup and store is the same even before `require-stable-task-ids` lands.
- **Late send to the main loop is safe.** If a registration is itself displaced between its Swap and its send on `newTaskCh`, the main loop's existing "still the current unit" check drops it, so the loser never gets scheduled.

## Risks / Trade-offs

- [Order of cancel callbacks across racing registrations is unspecified] → Spec only promises exactly-once, not ordering.
- [Displaced unit's timer may already have fired and be queued in `tasksIn`] → The main loop's identity check drops it, as today.
