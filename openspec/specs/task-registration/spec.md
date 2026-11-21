# task-registration Specification

## Purpose

Defines how tasks are registered with, replaced in, removed from and listed by a scheduler.

## Requirements

### Requirement: Registering a task
The scheduler SHALL accept a task function, a schedule and per-task options, and store the task under its ID. Registration is allowed whether or not the scheduler is running.

#### Scenario: Register while running
- **WHEN** a task is registered on a running scheduler
- **THEN** the task is scheduled immediately according to its schedule

#### Scenario: Register while idle
- **WHEN** a task is registered on an idle scheduler
- **THEN** the task is stored and scheduled when the scheduler starts

#### Scenario: Valid registration returns no error
- **WHEN** a non-nil task is registered with a non-nil schedule and valid options
- **THEN** RegisterTask returns nil

### Requirement: Task identity
Each task SHALL have an ID. When no ID is given, a random ID prefixed with `task-` is generated.

#### Scenario: Explicit ID
- **WHEN** a task is registered with an ID option
- **THEN** the task is stored under that ID

#### Scenario: Generated ID
- **WHEN** a task is registered without an ID option
- **THEN** the task is stored under a generated `task-` ID

### Requirement: Replacing a task
Registering a task with an ID that is already registered SHALL replace the existing task: its pending timer is stopped, its cancel callback is invoked asynchronously, and only the new task is scheduled from then on. Replacement SHALL be atomic: when registrations with the same ID race, exactly one task ends up registered and every other task registered under that ID is displaced and has its cancel callback invoked exactly once.

#### Scenario: Overwrite with same ID
- **WHEN** a second task is registered under an existing ID
- **THEN** the first task no longer runs, its cancel callback is called, and the second task runs on its schedule

#### Scenario: Concurrent registrations with the same ID
- **WHEN** N tasks are registered under the same ID concurrently
- **THEN** exactly one of them remains registered and runs, and the cancel callbacks of the other N-1 are each called exactly once

### Requirement: Removing a task
The scheduler SHALL remove a task by ID, stop its pending timer and invoke its cancel callback asynchronously. Removal SHALL be atomic: when removals of the same ID race, exactly one succeeds.

#### Scenario: Remove existing task
- **WHEN** a registered task is removed
- **THEN** it no longer runs and its cancel callback is called

#### Scenario: Remove unknown task
- **WHEN** RemoveTask is called with an ID that is not registered
- **THEN** an error is returned

#### Scenario: Concurrent removals of the same ID
- **WHEN** RemoveTask is called concurrently several times for one registered ID
- **THEN** exactly one call returns no error, the others return the not-found error, and the cancel callback is called exactly once

### Requirement: Listing tasks
The scheduler SHALL return the task functions of all registered tasks.

#### Scenario: List registered tasks
- **WHEN** two tasks are registered and ScheduledTasks is called
- **THEN** two task functions are returned

### Requirement: Registration validation
RegisterTask SHALL reject, with a non-nil error and without storing or scheduling anything, a nil task, a nil schedule, an interval schedule with a non-positive interval, a negative MaxDrift, a NotAfter that is not after NotBefore, backfill without NotBefore, and a nil callback function. The error SHALL match an exported sentinel error via errors.Is, and all detected problems SHALL be reported together.

#### Scenario: Nil task
- **WHEN** RegisterTask is called with a nil task
- **THEN** an error matching ErrInvalidTask is returned and no task is stored

#### Scenario: Nil schedule
- **WHEN** RegisterTask is called with a nil schedule
- **THEN** an error matching ErrInvalidSchedule is returned and no task is stored

#### Scenario: Zero interval
- **WHEN** RegisterTask is called with an interval schedule of 0
- **THEN** an error matching ErrInvalidSchedule is returned

#### Scenario: Negative drift
- **WHEN** RegisterTask is called with a negative MaxDrift
- **THEN** an error matching ErrInvalidOption is returned

#### Scenario: Inverted time range
- **WHEN** RegisterTask is called with NotAfter equal to or before NotBefore
- **THEN** an error matching ErrInvalidOption is returned

#### Scenario: Backfill without start
- **WHEN** RegisterTask is called with backfill enabled and no NotBefore
- **THEN** an error matching ErrInvalidOption is returned

#### Scenario: Nil callback
- **WHEN** RegisterTask is called with a success, error or cancel callback option holding a nil function
- **THEN** an error matching ErrInvalidOption is returned

#### Scenario: Multiple problems
- **WHEN** RegisterTask is called with a nil task and a negative MaxDrift
- **THEN** the returned error matches both ErrInvalidTask and ErrInvalidOption

### Requirement: Rejected registration leaves existing task intact
A registration that fails validation SHALL NOT replace, cancel or reschedule a task already registered under the same ID.

#### Scenario: Invalid overwrite
- **WHEN** a task is registered under ID "a" and then an invalid registration is attempted under ID "a"
- **THEN** the original task keeps running on its schedule and its cancel callback is not called
