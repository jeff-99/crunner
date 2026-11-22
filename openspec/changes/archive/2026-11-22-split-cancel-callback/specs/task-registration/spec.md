## MODIFIED Requirements

### Requirement: Replacing a task
Registering a task with an ID that is already registered SHALL replace the existing task: its pending timer is stopped, its removed callback is invoked asynchronously, and only the new task is scheduled from then on. Replacement SHALL be atomic: when registrations with the same ID race, exactly one task ends up registered and every other task registered under that ID is displaced and has its removed callback invoked exactly once. Replacement and removal are the only events that invoke a task's removed callback.

#### Scenario: Overwrite with same ID
- **WHEN** a second task is registered under an existing ID
- **THEN** the first task no longer runs, its removed callback is called, and the second task runs on its schedule

#### Scenario: Concurrent registrations with the same ID
- **WHEN** N tasks are registered under the same ID concurrently
- **THEN** exactly one of them remains registered and runs, and the removed callbacks of the other N-1 are each called exactly once

#### Scenario: Overwrite does not call the skip callback
- **WHEN** a task with a skip callback is replaced
- **THEN** its skip callback is not called

### Requirement: Removing a task
The scheduler SHALL remove a task by ID, stop its pending timer and invoke its removed callback asynchronously. Removal SHALL be atomic: when removals of the same ID race, exactly one succeeds.

#### Scenario: Remove existing task
- **WHEN** a registered task is removed
- **THEN** it no longer runs and its removed callback is called

#### Scenario: Remove unknown task
- **WHEN** RemoveTask is called with an ID that is not registered
- **THEN** an error is returned

#### Scenario: Concurrent removals of the same ID
- **WHEN** RemoveTask is called concurrently several times for one registered ID
- **THEN** exactly one call returns no error, the others return the not-found error, and the removed callback is called exactly once

#### Scenario: Removal does not call the skip callback
- **WHEN** a task with a skip callback is removed
- **THEN** its skip callback is not called

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
- **WHEN** RegisterTask is called with a success, error, skip or removed callback option holding a nil function
- **THEN** an error matching ErrInvalidOption is returned

#### Scenario: Multiple problems
- **WHEN** RegisterTask is called with a nil task and a negative MaxDrift
- **THEN** the returned error matches both ErrInvalidTask and ErrInvalidOption

### Requirement: Rejected registration leaves existing task intact
A registration that fails validation SHALL NOT replace, remove or reschedule a task already registered under the same ID.

#### Scenario: Invalid overwrite
- **WHEN** a task is registered under ID "a" and then an invalid registration is attempted under ID "a"
- **THEN** the original task keeps running on its schedule and its removed callback is not called
