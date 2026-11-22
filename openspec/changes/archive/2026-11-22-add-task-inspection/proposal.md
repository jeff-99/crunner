## Why

`ScheduledTasks()` returns bare `Task` funcs, so callers cannot tell which tasks exist, when they run next, or how they have been doing. Since task IDs may be generated at registration, callers also have no way to discover them. Operators and tests need a read-only view of each task's identity, schedule state and run statistics, comparable to gocron v2's job inspection.

## What Changes

- **BREAKING**: `ScheduledTasks()` returns `[]TaskInfo` (sorted by ID) instead of `[]Task`.
- Add `Task(id string) (TaskInfo, bool)` to the `Scheduler` interface to inspect a single task.
- Add exported `TaskInfo` value type with the same field names `TriggerMetadata` uses: `ID`, `NextRunTime`, `WindowStart`, `WindowEnd`, `LastRunTime`, `LastSuccessfulRunTime`, `LastErroredRunTime`, `Count`, `SuccessfulCount`, `ErrorCount`, `LastError`, plus `Running`. Zero times mean "never" / "not scheduled".
- `WindowStart` and `WindowEnd` are the window of the upcoming run while the task is waiting, and the window of the executing run while `Running` is true.
- Returned values are snapshots: copies that never alias internal state and are safe to read concurrently with the scheduler, before `Start`, while running and after `Stop`.

## Capabilities

### New Capabilities
- `task-inspection`: read-only snapshots of registered tasks and their scheduling/run state.

### Modified Capabilities
- `task-registration`: the "Listing tasks" requirement now returns task info snapshots instead of task functions.
- `scheduler-lifecycle`: the "Concurrency safety" requirement also covers Start and Stop and the scheduler context. A test that inspects tasks while the scheduler is repeatedly started and stopped found that Start wrote the scheduler context while timers left over from the previous run, `RegisterTask`, `RemoveTask` and `Stop` read it.

## Impact

- Code: `scheduler.go` (scheduler context held in an atomic pointer and `Start`/`Stop` serialised by a mutex, `Scheduler` interface, `ScheduledTasks`, new `Task` method, new `TaskInfo` type, state published by the main loop), tests in `scheduler_test.go`.
- API: **BREAKING** — return type of `ScheduledTasks()` changes; `Scheduler` gains a method (breaks external implementations of the interface).
- Dependencies: none outstanding (`fix-shared-state-races` is merged: the main loop owns task state). Tags are not part of this change; `add-task-tags` adds a `Tags` field to `TaskInfo` when it lands. Skipped-run counters are left for a follow-up.
