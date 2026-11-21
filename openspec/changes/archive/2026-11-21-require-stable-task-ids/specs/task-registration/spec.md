## MODIFIED Requirements

### Requirement: Task identity
Each task SHALL have an ID that is fixed at registration and does not change for the lifetime of the registration. When no ID is given, or the given ID is empty, the scheduler SHALL generate exactly one random ID prefixed with `task-` at registration and use it for every subsequent lookup, scheduling, replacement and removal of that task.

#### Scenario: Explicit ID
- **WHEN** a task is registered with an ID option
- **THEN** the task is stored under that ID

#### Scenario: Generated ID
- **WHEN** a task is registered without an ID option
- **THEN** the task is stored under a generated `task-` ID

#### Scenario: Empty ID
- **WHEN** a task is registered with an empty ID option on a running scheduler
- **THEN** the task is stored under a generated `task-` ID and runs on its schedule

#### Scenario: Generated ID is stable
- **WHEN** a task with a generated ID has been registered
- **THEN** the same ID is used for every run of the task and removing the task by that ID succeeds

#### Scenario: Generated IDs are distinct
- **WHEN** two tasks are registered without an ID option
- **THEN** each is stored under a different generated ID and both run
