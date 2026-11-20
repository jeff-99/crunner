# task-execution Specification

## Purpose

Defines how a scheduled task is executed: trigger windows and metadata, backfill, the maximum-drift rule, callbacks and run statistics.

## Requirements

### Requirement: Trigger metadata
Each task invocation SHALL receive a context and a snapshot of its trigger metadata: expected and actual trigger time, trigger window start and end, last run, last successful and last errored run times, run counts and the last error.

#### Scenario: Task receives its window
- **WHEN** a task is invoked
- **THEN** the metadata contains the window start and end that this run covers

### Requirement: Trigger windows
The first window of a task SHALL start at its registration time (or at NotBefore when backfilling). Each later window SHALL start at the previous window's end.

#### Scenario: First window
- **WHEN** a task without backfill runs for the first time
- **THEN** its window starts at the registration time and ends at the schedule's next run time

### Requirement: Backfill
When backfill is enabled and NotBefore is set, the scheduler SHALL run the task immediately and repeatedly for each window from NotBefore until a window ends in the future, after which the task runs on its normal schedule.

#### Scenario: Catch up past windows
- **WHEN** an interval task of 10 seconds is registered with backfill and NotBefore 100 seconds in the past
- **THEN** it runs back-to-back for each past window and then continues on schedule

### Requirement: Maximum drift
The scheduler SHALL skip a run whose actual start is later than its expected trigger time by more than the task's maximum drift (default 30 seconds), invoke its cancel callback, and schedule the next run from the current time. The cancel callback SHALL be invoked asynchronously so that it never blocks the scheduling of other tasks, and skipping any number of runs at once SHALL NOT stall the scheduler.

#### Scenario: Excessive drift
- **WHEN** a run starts more than MaxDrift after its expected trigger time
- **THEN** the task is not executed, the cancel callback is called and the next run is scheduled

#### Scenario: Blocking cancel callback
- **WHEN** a run is skipped for drift and its cancel callback blocks indefinitely
- **THEN** other tasks keep running on their schedules and the skipped task's next run is still scheduled

#### Scenario: Mass skip
- **WHEN** more runs are skipped for drift at the same moment than the scheduler's internal queue capacity (more than 100 tasks)
- **THEN** every skipped task is rescheduled and the scheduler continues to run tasks

#### Scenario: Stop waits for skip callbacks
- **WHEN** Stop is called while a drift-skip cancel callback is running
- **THEN** Stop returns only after that callback has returned

### Requirement: Non-overlapping runs
A task SHALL NOT run concurrently with itself; the next run is scheduled only after the current run returns.

#### Scenario: Long-running task
- **WHEN** a task runs longer than its interval
- **THEN** no second instance of that task is started while it runs

### Requirement: Callbacks
The scheduler SHALL call the task's success callback after a run that returns nil, and its error callback with the error after a run that returns an error.

#### Scenario: Error callback
- **WHEN** a task returns an error
- **THEN** the error callback is called with that error and the error count is incremented

#### Scenario: Success callback
- **WHEN** a task returns nil
- **THEN** the success callback is called and the success count is incremented

### Requirement: Task timeout
Each run SHALL receive a context that is cancelled when the scheduler stops or after 30 days.

#### Scenario: Scheduler stop cancels the run
- **WHEN** the scheduler is stopped during a run
- **THEN** the run's context is cancelled
