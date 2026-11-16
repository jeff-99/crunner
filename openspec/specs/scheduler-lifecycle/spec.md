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
