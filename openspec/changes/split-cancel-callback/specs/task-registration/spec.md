## MODIFIED Requirements

### Requirement: Replacing a task
Registering a task with an ID that is already registered SHALL replace the existing task: its pending timer is stopped, its cancel callback is invoked asynchronously, and only the new task is scheduled from then on. Replacement and removal are the only events that invoke a task's cancel callback.

#### Scenario: Overwrite with same ID
- **WHEN** a second task is registered under an existing ID
- **THEN** the first task no longer runs, its cancel callback is called, and the second task runs on its schedule

#### Scenario: Overwrite does not call skip callback
- **WHEN** a task with a skip callback is replaced
- **THEN** its skip callback is not called

