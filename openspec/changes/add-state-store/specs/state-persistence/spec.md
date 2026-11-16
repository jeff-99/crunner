# Spec Delta

## Purpose

Lets the scheduler persist, per task, which trigger window was handled last, so that window coverage resumes across process restarts.

## ADDED Requirements

### Requirement: Pluggable state store
The library SHALL accept a user-supplied state store through a scheduler option. A state store loads the state of a task by task ID, reporting whether state exists, and saves the state of a task by task ID.

#### Scenario: Custom store is used
- **WHEN** a scheduler is created with a custom state store and a task runs
- **THEN** the scheduler loads and saves that task's state through the custom store

### Requirement: In-memory default store
Without a configured state store the scheduler SHALL use an in-memory store, so behaviour within a single process is unchanged and state does not survive the process.

#### Scenario: No store configured
- **WHEN** a scheduler is created without a state store and a task runs
- **THEN** the task's windows are identical to those of a scheduler without state persistence

### Requirement: Recorded task state
The state of a task SHALL contain the start and end of the last trigger window the scheduler handled for it.

#### Scenario: State after a run
- **WHEN** a run covering window [S, E) returns
- **THEN** the saved state of the task holds window start S and window end E

### Requirement: State saved after each handled window
The scheduler SHALL save a task's state after a run returns, whether it succeeded or failed, and after a run is skipped for excessive drift, before the next window is scheduled.

#### Scenario: Failed run is recorded
- **WHEN** a run returns an error
- **THEN** its window is saved as handled and is not retried

#### Scenario: Skipped run is recorded
- **WHEN** a run is skipped for excessive drift
- **THEN** its window is saved as handled

### Requirement: State loaded on first scheduling
The scheduler SHALL load a task's state once, before computing its first window after registration or after the scheduler starts.

#### Scenario: Restart resumes
- **WHEN** a process restarts with a persistent store and re-registers a task whose saved window end is E
- **THEN** the task's first window starts at E

### Requirement: Load failure
If loading a task's state fails, the scheduler SHALL log the error and not schedule the task, exactly as for a schedule error.

#### Scenario: Store unavailable at start
- **WHEN** the state store returns an error on load
- **THEN** the task is not scheduled and the error is logged

### Requirement: Save failure
If saving a task's state fails, the scheduler SHALL log the error and continue scheduling the task from its in-memory window state.

#### Scenario: Store unavailable after run
- **WHEN** the state store returns an error on save
- **THEN** the error is logged and the next window is scheduled as usual
