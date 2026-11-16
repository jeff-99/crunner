## Why

Applications register groups of related tasks (per tenant, per pipeline, per feature) and need to remove a whole group at once. Today the only handle is the task ID (`RemoveTask(id)`, `scheduler.go` ~649), so callers must track ID lists themselves (gocron v2 has `WithTags` / `RemoveByTags`).

## What Changes

- Add a per-task option `WithTags(tags ...string)`; tags are stored deduplicated with the task. Empty tags are rejected by `RegisterTask`.
- Add `RemoveTasksByTag(tags ...string) int` to the `Scheduler` interface: removes every task that has at least one of the given tags, with the same effect as `RemoveTask` (timer stopped, cancel callback called), and returns the number removed.
- Replacing a task (same ID) replaces its tags.
- **BREAKING**: adds a method to the `Scheduler` interface (breaks third-party implementations/mocks). `RegisterTask` now returns an error for empty tags.

## Capabilities

### New Capabilities

### Modified Capabilities
- `task-registration`: tagging tasks and removing tasks by tag.

## Impact

- `scheduler.go`: `SchedulingOptions.Tags`, `WithTags`, `RemoveTasksByTag`.
- Dependencies: none required. `add-task-inspection` should expose tags in task info. If `validate-task-registration` lands first, put the empty-tag check in its validation function. If `make-task-registration-atomic` lands first, removal by tag must use the same synchronisation as `RemoveTask`.
