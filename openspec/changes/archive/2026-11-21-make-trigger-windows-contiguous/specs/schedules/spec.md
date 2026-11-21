## MODIFIED Requirements

### Requirement: Pluggable schedules
The library SHALL accept any schedule that, given a scheduling start time and an optional reference time, returns the next run time or an error. The reference time SHALL be a previous point of the task's window sequence — normally the end of the task's previous trigger window — and never the actual start of a run. It is absent for the first window.

#### Scenario: Custom schedule
- **WHEN** a task is registered with a user-defined schedule
- **THEN** the scheduler uses that schedule to compute each next run time

#### Scenario: Schedule error
- **WHEN** a schedule returns an error while computing the next run
- **THEN** the error is logged and the task is not scheduled again

#### Scenario: Reference time is the previous window end
- **WHEN** a custom schedule is asked for the next run of a task whose previous window ended at T and whose run started at T+3s
- **THEN** the reference time passed to the schedule is T

### Requirement: Interval schedule
The library SHALL provide an interval schedule whose next run is the reference time plus the interval, or the scheduling start plus the interval when there is no reference time.

#### Scenario: First interval run
- **WHEN** an interval task has not run yet
- **THEN** its next run is the scheduling start plus the interval

#### Scenario: Subsequent interval run
- **WHEN** an interval task has a previous window
- **THEN** its next run is that window's end plus the interval, regardless of when the previous run started
