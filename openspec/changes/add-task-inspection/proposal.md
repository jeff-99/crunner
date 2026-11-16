## Why

`ScheduledTasks()` (`scheduler.go:634`) returns bare `Task` funcs, so callers cannot tell which tasks exist, when they run next, or how they have been doing. Operators and tests need a read-only view of each task's identity, schedule state and run statistics, comparable to gocron v2's job inspection.

## What Changes

- **BREAKING**: `ScheduledTasks()` returns `[]TaskInfo` (sorted by ID) instead of `[]Task`.
- Add `Task(id string) (TaskInfo, bool)` to the `Scheduler` interface to inspect a single task.
- Add exported `TaskInfo` value type: ID, next run time, current trigger window (start/end), last run / last success / last error times, run / success / error counts, last error, and whether a run is in progress.
- Returned values are snapshots: copies that never alias internal state and are safe to read concurrently with the scheduler.

## Capabilities

### New Capabilities
- `task-inspection`: read-only snapshots of registered tasks and their scheduling/run state.

### Modified Capabilities
- `task-registration`: the "Listing tasks" requirement now returns task info snapshots instead of task functions.

## Impact

- Code: `scheduler.go` (`Scheduler` interface, `ScheduledTasks`, new `Task` method, new `TaskInfo` type); tests in `scheduler_test.go`.
- API: **BREAKING** — return type of `ScheduledTasks()` changes; `Scheduler` gains a method (breaks external implementations of the interface).
- Dependencies: depends on `fix-shared-state-races` (main loop as sole owner of task state, so snapshots can be taken without races). Tags are not part of this change; `add-task-tags` adds a `Tags` field to `TaskInfo` when it lands.
