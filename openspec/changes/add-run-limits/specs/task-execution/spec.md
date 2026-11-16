## ADDED Requirements

### Requirement: NotAfter limit
The scheduler SHALL NOT execute a run whose trigger window ends after the task's NotAfter time. When the next window would end after NotAfter, the task SHALL complete.

#### Scenario: Interval task stops at NotAfter
- **WHEN** a 10-second interval task is registered with NotAfter 35 seconds after registration
- **THEN** it runs for the windows ending at 10, 20 and 30 seconds and then completes

#### Scenario: Window ending exactly at NotAfter
- **WHEN** a task's next window ends exactly at NotAfter
- **THEN** that run is executed

#### Scenario: Backfill bounded by NotAfter
- **WHEN** a task with backfill has NotBefore and NotAfter both in the past
- **THEN** it backfills only the windows ending at or before NotAfter and then completes

### Requirement: Maximum number of runs
The scheduler SHALL complete a task after it has executed the configured maximum number of runs. Successful and failed runs SHALL count; runs skipped for excessive drift SHALL NOT count.

#### Scenario: Limited runs
- **WHEN** an interval task is registered with a maximum of 3 runs
- **THEN** it runs exactly 3 times and then completes

#### Scenario: Skipped runs not counted
- **WHEN** a task with a maximum of 2 runs has one run skipped for drift
- **THEN** it still executes 2 runs before completing

### Requirement: Completion of limited tasks
When a task completes because of NotAfter or its maximum number of runs, the scheduler SHALL remove it from its registered tasks without invoking its cancel callback.

#### Scenario: Completed task removed
- **WHEN** a task reaches its maximum number of runs
- **THEN** it is no longer listed by ScheduledTasks and its cancel callback was not called
