# task-inspection Specification

## Purpose

Lets callers inspect registered tasks — identity, next run, current trigger window and run statistics — through read-only snapshots.

## Requirements

### Requirement: Inspect a single task
The scheduler SHALL return a snapshot of a registered task by ID, and report when no task with that ID is registered.

#### Scenario: Known task
- **WHEN** `Task(id)` is called for a registered task
- **THEN** a snapshot with that ID is returned and the found flag is true

#### Scenario: Unknown task
- **WHEN** `Task(id)` is called with an ID that is not registered
- **THEN** a zero snapshot is returned and the found flag is false

### Requirement: Task snapshot contents
A task snapshot SHALL contain the task ID, the next scheduled run time, the trigger window start and end, the last run, last successful run and last errored run times, the run, success and error counts, the last error, and whether a run is currently in progress. The window SHALL be the window of the upcoming run while the task is waiting, and the window of the executing run while a run is in progress. Times that do not apply SHALL be zero.

#### Scenario: After a successful run
- **WHEN** a task has run once successfully and is scheduled again
- **THEN** its snapshot reports count 1, success count 1, error count 0, its last run time, and the next run time and window of the upcoming run

#### Scenario: Run in progress
- **WHEN** a task is inspected while its run is executing
- **THEN** its snapshot reports a run in progress

#### Scenario: Errored run
- **WHEN** a task has returned an error once
- **THEN** its snapshot reports error count 1, the error, and the time of that run as last errored run time

#### Scenario: Not yet scheduled
- **WHEN** a task is registered on an idle scheduler and inspected
- **THEN** its snapshot reports zero counts and no next run time

#### Scenario: After Stop
- **WHEN** a scheduler with a waiting task is stopped and the task is inspected
- **THEN** its snapshot reports no next run time and no run in progress, and keeps its counts

#### Scenario: Schedule error
- **WHEN** a task's schedule returns an error and the task is inspected
- **THEN** its snapshot reports no next run time

### Requirement: Listing in a stable order
The scheduler SHALL return the snapshots of all registered tasks ordered by task ID, including tasks whose IDs were generated at registration.

#### Scenario: Generated IDs are discoverable
- **WHEN** a task is registered without an ID and the registered tasks are listed
- **THEN** the listing contains a snapshot whose ID is the generated ID, and that ID can be used with the single-task lookup and with removal

### Requirement: Snapshot isolation
Snapshots SHALL be independent copies: mutating a snapshot does not affect the scheduler, later scheduler activity does not change an already returned snapshot, and inspection is safe to call at any time, including concurrently with running tasks, Start and Stop.

#### Scenario: Snapshot does not change
- **WHEN** a snapshot is taken and the task runs again afterwards
- **THEN** the earlier snapshot still reports the previous counts

#### Scenario: Concurrent inspection
- **WHEN** inspection is called repeatedly from another goroutine while tasks run
- **THEN** the race detector reports no data race
