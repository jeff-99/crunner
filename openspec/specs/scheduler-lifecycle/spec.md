# scheduler-lifecycle Specification

## Purpose

Defines how a scheduler instance is created, started and stopped, and how scheduler-wide options apply.

## Requirements

### Requirement: Scheduler construction
The library SHALL provide a constructor that returns an idle scheduler, configurable through functional options for the logger and the task summary interval.

#### Scenario: Default construction
- **WHEN** a scheduler is created without options
- **THEN** it is idle, logs nothing, and uses a task summary interval of 30 seconds

### Requirement: Starting the scheduler
Starting an idle scheduler SHALL derive a scheduler context from the given context and schedule every task registered so far.

#### Scenario: Start with pre-registered tasks
- **WHEN** tasks are registered and the scheduler is then started
- **THEN** each registered task is scheduled for its first run

#### Scenario: Start twice
- **WHEN** Start is called on a scheduler that is already running
- **THEN** Start returns an error and the running scheduler is unaffected

### Requirement: Stopping the scheduler
Stopping a running scheduler SHALL cancel the scheduler context, stop all pending timers, wait for running tasks to return, and return the scheduler to idle.

#### Scenario: Stop waits for running tasks
- **WHEN** Stop is called while a task is executing
- **THEN** the task's context is cancelled and Stop returns only after the task has returned

#### Scenario: Stop when not running
- **WHEN** Stop is called on an idle scheduler
- **THEN** Stop returns no error

#### Scenario: Restart after stop
- **WHEN** a stopped scheduler is started again
- **THEN** its registered tasks are scheduled again

### Requirement: Periodic task summaries
While running, the scheduler SHALL periodically log, at info level, a summary per task that ran since the previous summary, with run, success and error counts.

#### Scenario: Summary only for active tasks
- **WHEN** the summary interval elapses and a task has not run since the last summary
- **THEN** no summary is logged for that task

### Requirement: Concurrency safety
All exported scheduler methods SHALL be safe to call concurrently with each other and with running tasks, and the scheduler SHALL NOT perform unsynchronised concurrent access to per-task state such as timers and trigger metadata.

#### Scenario: Concurrent register, remove and stop under the race detector
- **WHEN** tasks are registered, replaced and removed from several goroutines while the scheduler is running and tasks are executing, and then Stop is called, with the race detector enabled
- **THEN** no data race is reported and Stop returns after all running tasks have returned

#### Scenario: Metadata updated while summaries are emitted
- **WHEN** tasks complete runs while the periodic summary is being produced, with the race detector enabled
- **THEN** no data race is reported and run counts reflect every completed run

#### Scenario: Replaced task's pending timer does not run it
- **WHEN** a task with a pending timer is replaced or removed and its old timer fires afterwards
- **THEN** the old task is not executed
