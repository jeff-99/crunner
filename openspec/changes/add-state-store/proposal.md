# Proposal

## Why

Window progress lives only in memory (`schedulingUnit.meta`, `scheduler.go` ~141-160). After a process restart a task starts over: without backfill its first window starts at the new registration time (the time between shutdown and restart is never covered), and with backfill it replays every window since `NotBefore` again. Users who rely on the trigger-window guarantee need the scheduler to resume where it left off.

## What Changes

- Add an exported `StateStore` interface (`Load(ctx, taskID) (State, bool, error)`, `Save(ctx, taskID, State) error`) and an exported `State` struct recording the last handled trigger window of a task.
- Add a scheduler option `WithStateStore(StateStore)`. The default is an in-memory store, so behaviour without the option is unchanged.
- When a task is first scheduled, the scheduler loads its state; if present, the first window starts at the recorded window end instead of the registration time / `NotBefore`.
- After every handled window (run finished with success or error, or run skipped for drift) the scheduler saves the window as handled.
- Store errors: a failed `Load` is treated like a schedule error (task not scheduled, error logged); a failed `Save` is logged and scheduling continues.
- No backend other than the in-memory store is shipped; persistent backends (e.g. Postgres) are implemented by users.

## Capabilities

### New Capabilities
- `state-persistence`: the state store abstraction, what is recorded per task, when it is loaded and saved, and how store errors are handled.

### Modified Capabilities
- `task-execution`: "Trigger windows" and "Backfill" now start from the persisted window end when state exists.

## Impact

- Code: `scheduler.go` (scheduling of the first window, task goroutine after run, drift-skip path), new file `state.go`.
- API: additive only — `State`, `StateStore`, `NewMemoryStateStore`, `WithStateStore`. No breaking change.
- Dependencies: no new modules.
- Depends on: `require-stable-task-ids` (state is keyed by task ID; random IDs make it useless), `make-trigger-windows-contiguous` (persisted "last window end" is only a sound resume point when the next window is derived from it), `fix-shared-state-races` and `fix-mainloop-self-send-deadlock` (store I/O happens off the main loop, including on the drift-skip path).
- Unlocks: `add-distributed-locking` builds on `State` to skip windows already handled by another replica.
