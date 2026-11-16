## MODIFIED Requirements

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
