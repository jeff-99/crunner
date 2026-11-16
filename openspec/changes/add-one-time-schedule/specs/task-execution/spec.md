## ADDED Requirements

### Requirement: Task completion
When a task's schedule reports that it has no further runs, the scheduler SHALL remove the task from its registered tasks without invoking the task's cancel callback.

#### Scenario: One-time task removed after run
- **WHEN** a one-time task has run
- **THEN** it is no longer listed by ScheduledTasks and its cancel callback was not called

#### Scenario: Task ID reusable after completion
- **WHEN** a one-time task has completed and a new task is registered under the same ID
- **THEN** the new task is scheduled normally and no cancel callback is called for the completed task
