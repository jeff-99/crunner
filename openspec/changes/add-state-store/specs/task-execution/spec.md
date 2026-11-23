# Spec Delta

## MODIFIED Requirements

### Requirement: Trigger windows
The first window of a task SHALL start at the window end recorded in its persisted state when such state exists; otherwise at its registration time (or at NotBefore when backfilling). Each later window SHALL start at the previous window's end.

#### Scenario: First window
- **WHEN** a task without backfill and without persisted state runs for the first time
- **THEN** its window starts at the registration time and ends at the schedule's next run time

#### Scenario: First window with persisted state
- **WHEN** a task without backfill has persisted state with window end E and runs for the first time in this process
- **THEN** its window starts at E and ends at the schedule's next run time

### Requirement: Backfill
When backfill is enabled and NotBefore is set, the scheduler SHALL run the task immediately and repeatedly for each window from NotBefore, or from the persisted window end when it is later than NotBefore, until a window ends in the future, after which the task runs on its normal schedule.

#### Scenario: Catch up past windows
- **WHEN** an interval task of 10 seconds is registered with backfill and NotBefore 100 seconds in the past
- **THEN** it runs back-to-back for each past window and then continues on schedule

#### Scenario: Resume backfill after restart
- **WHEN** a task with backfill is re-registered after a restart and its persisted window end is later than NotBefore
- **THEN** backfill starts at the persisted window end and windows before it are not run again
