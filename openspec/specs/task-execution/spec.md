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
The windows of a task SHALL be contiguous and non-overlapping: each window SHALL start exactly at the end of the task's previous executed window, and its end SHALL be a point of the schedule's sequence computed from the previous window's end, independent of when runs actually started or finished. The first window SHALL start when the task is first scheduled — at registration on a running scheduler, at Start for a task registered while the scheduler was idle — or at NotBefore when backfilling. A window normally spans one step of the schedule; after a run is cancelled for drift, the next window spans several steps (see Maximum drift). Window starts and ends SHALL always be points of the schedule's sequence.

#### Scenario: First window
- **WHEN** a task without backfill is registered on a running scheduler and runs for the first time
- **THEN** its window starts at the registration time and ends at the schedule's next run time

#### Scenario: First window of a task registered before Start
- **WHEN** an interval task of 10 seconds is registered on an idle scheduler and the scheduler is started 1 minute later
- **THEN** its first window starts at the Start time and ends 10 seconds later, and the run is not cancelled

#### Scenario: Late start does not shift windows
- **WHEN** an interval task of 10 seconds with window [T, T+10s) starts its run 2 seconds late
- **THEN** its next window is [T+10s, T+20s), not [T+10s, T+22s)

#### Scenario: Windows stay on the grid over time
- **WHEN** an interval task of 10 seconds runs 100 times, each run starting with a small delay
- **THEN** the 100th window ends exactly 1000 seconds after the first window's start

### Requirement: Backfill
When backfill is enabled and NotBefore is set, the scheduler SHALL run the task immediately and repeatedly for each window from NotBefore until a window ends in the future, after which the task runs on its normal schedule.

#### Scenario: Catch up past windows
- **WHEN** an interval task of 10 seconds is registered with backfill and NotBefore 100 seconds in the past
- **THEN** it runs back-to-back for each past window and then continues on schedule

### Requirement: Drift skipping
The expected trigger time of a run SHALL be the later of its window end and the time the task's previous run finished. When a task that is not backfilling would start more than its maximum drift (default 30 seconds) after its expected trigger time, the run SHALL be skipped: it SHALL NOT be executed and the task's skip callback SHALL be invoked once. The next run's window SHALL start at the skipped window's start and end at the first point of the schedule's sequence after the current time, and that run SHALL be triggered at that end as normal. No part of the timeline SHALL be left uncovered. The skip callback SHALL be invoked asynchronously so that it never blocks scheduling, and any number of simultaneous skips SHALL NOT stall the scheduler. The removed callback SHALL NOT be invoked for a skipped run.

#### Scenario: Excessive drift
- **WHEN** an interval task of 10 seconds with MaxDrift 5 seconds has window [10s, 20s) and its timer fires at 27s
- **THEN** the run is not executed, the skip callback is called once, and the next run has window [10s, 30s) and is triggered at 30s

#### Scenario: Long pause
- **WHEN** the same task's timer for window [10s, 20s) fires at 83s because the process was paused
- **THEN** the skip callback is called once, no run is executed at 83s, and the next run has window [10s, 90s) and is triggered at 90s

#### Scenario: Restart after a long stop
- **WHEN** the scheduler is stopped at 15s while the same task's window [10s, 20s) is pending and started again at 83s
- **THEN** the skip callback is called once, and the next run has window [10s, 90s) and is triggered at 90s

#### Scenario: Long-running task is not skipped
- **WHEN** an interval task of 10 seconds with MaxDrift 1 second runs for 25 seconds
- **THEN** its next window starts at the previous window's end and spans one interval, the next run starts immediately after the long run finishes, and the skip callback is not called

#### Scenario: Blocking skip callback
- **WHEN** a run is skipped for drift and its skip callback blocks indefinitely
- **THEN** other tasks keep running on their schedules and the skipped task's next run is still scheduled

#### Scenario: Mass skip
- **WHEN** more runs are skipped for drift at the same moment than the scheduler's internal queue capacity (more than 100 tasks)
- **THEN** every skipped task is rescheduled and the scheduler continues to run tasks

#### Scenario: Stop waits for skip callbacks
- **WHEN** Stop is called while a skip callback is running
- **THEN** Stop returns only after that callback has returned

#### Scenario: Skip does not call the removed callback
- **WHEN** a run is skipped for drift on a task that has both a skip callback and a removed callback
- **THEN** only the skip callback is called

#### Scenario: Skip metadata
- **WHEN** a run with window [10s, 20s) is skipped for drift at 27s
- **THEN** the skip callback receives trigger metadata with window start 10s, window end 20s, the run's expected trigger time, and 27s as actual trigger time

### Requirement: Non-overlapping runs
A task SHALL NOT run concurrently with itself; the next run is scheduled only after the current run returns.

#### Scenario: Long-running task
- **WHEN** a task runs longer than its interval
- **THEN** no second instance of that task is started while it runs

### Requirement: Callbacks
The scheduler SHALL call the task's success callback after a run that returns nil, and its error callback with the error after a run that returns an error or panics. A panicking run SHALL be treated as an errored run whose error is a panic error carrying the panic value and stack trace. It SHALL call the task's skip callback when a run is skipped for drift (see Drift skipping), and the task's removed callback only when the task is removed or replaced.

#### Scenario: Error callback
- **WHEN** a task returns an error
- **THEN** the error callback is called with that error and the error count is incremented

#### Scenario: Success callback
- **WHEN** a task returns nil
- **THEN** the success callback is called and the success count is incremented

#### Scenario: Panic reported as error
- **WHEN** a task panics with value `"boom"`
- **THEN** the error callback is called with a panic error whose value is `"boom"` and which includes a stack trace, the error count is incremented, and the success callback is not called

### Requirement: Task timeout
Each run SHALL receive a context that is cancelled when the scheduler stops or when the task timeout elapses. The task timeout SHALL be the task's own timeout option if set, otherwise the scheduler-wide task timeout, which defaults to 30 days. A timeout of zero or less SHALL mean the run is only cancelled when the scheduler stops.

#### Scenario: Scheduler stop cancels the run
- **WHEN** the scheduler is stopped during a run
- **THEN** the run's context is cancelled

#### Scenario: Default timeout
- **WHEN** a task runs on a scheduler created without a task timeout option
- **THEN** the run's context has a deadline 30 days after the run starts

#### Scenario: Scheduler-wide timeout
- **WHEN** the scheduler is created with a task timeout of 5 seconds and a task runs for longer
- **THEN** the run's context is cancelled 5 seconds after the run starts

#### Scenario: Per-task override
- **WHEN** the scheduler task timeout is 5 seconds and a task is registered with its own timeout of 1 second
- **THEN** that task's run context is cancelled 1 second after the run starts

#### Scenario: Timeout disabled
- **WHEN** the effective task timeout is zero
- **THEN** the run's context has no deadline

### Requirement: Panic containment
A panic raised by a task function SHALL NOT terminate the process or the scheduler. After a panicking run the task SHALL be rescheduled as after any errored run, and other tasks SHALL keep running.

#### Scenario: Scheduler survives a panicking task
- **WHEN** a task panics on every run
- **THEN** the process keeps running, the task is run again on its next scheduled window, and other registered tasks keep running on their schedules

#### Scenario: Panic with an error value
- **WHEN** a task panics with an `error` value
- **THEN** the panic error passed to the error callback unwraps to that original error
