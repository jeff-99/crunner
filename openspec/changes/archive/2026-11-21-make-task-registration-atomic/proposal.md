## Why

`RegisterTask` checks for an existing task with `s.tasks.Load` (`scheduler.go:229`) and later stores the new one with `s.tasks.Store` (`scheduler.go:261`). Two concurrent registrations with the same ID can both see "no existing task" (or the same existing task), so a displaced task is never cancelled, or one task's cancel callback fires twice while another displaced task is silently dropped. `RemoveTask` has the same Load-then-Delete gap (`scheduler.go:644/650`): two concurrent removals of the same ID can both succeed and both fire the cancel callback.

## What Changes

- Registration swaps the new task into the registry atomically and receives exactly the task it displaced, if any; that task (and only that task) is cancelled once.
- Removal atomically loads-and-deletes, so exactly one concurrent `RemoveTask` for an ID succeeds and the others return the not-found error.
- No public API change.

## Capabilities

### New Capabilities

### Modified Capabilities
- `task-registration`: "Replacing a task" and "Removing a task" gain guarantees for concurrent calls with the same ID.

## Impact

- Code: `scheduler.go` (`RegisterTask`, `RemoveTask`), tests.
- API: none.
- Dependencies: none.
- Depends on: none. Assumes a task's ID is stable for the duration of a call; with an empty ID (`WithID("")`) `schedulingUnit.ID()` returns a different value on each call, which is fixed separately by `require-stable-task-ids`. This change computes the ID once per call so it is correct regardless. Timer handling is left as is; `fix-shared-state-races` moves it to the main loop.
