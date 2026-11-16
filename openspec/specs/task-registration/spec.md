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

### Requirement: Task identity
Each task SHALL have an ID. When no ID is given, a random ID prefixed with `task-` is generated.

#### Scenario: Explicit ID
- **WHEN** a task is registered with an ID option
- **THEN** the task is stored under that ID

#### Scenario: Generated ID
- **WHEN** a task is registered without an ID option
- **THEN** the task is stored under a generated `task-` ID

### Requirement: Replacing a task
Registering a task with an ID that is already registered SHALL replace the existing task: its pending timer is stopped, its cancel callback is invoked asynchronously, and only the new task is scheduled from then on.

#### Scenario: Overwrite with same ID
- **WHEN** a second task is registered under an existing ID
- **THEN** the first task no longer runs, its cancel callback is called, and the second task runs on its schedule

### Requirement: Removing a task
The scheduler SHALL remove a task by ID, stop its pending timer and invoke its cancel callback asynchronously.

#### Scenario: Remove existing task
- **WHEN** a registered task is removed
- **THEN** it no longer runs and its cancel callback is called

#### Scenario: Remove unknown task
- **WHEN** RemoveTask is called with an ID that is not registered
- **THEN** an error is returned

### Requirement: Listing tasks
The scheduler SHALL return the task functions of all registered tasks.

#### Scenario: List registered tasks
- **WHEN** two tasks are registered and ScheduledTasks is called
- **THEN** two task functions are returned
