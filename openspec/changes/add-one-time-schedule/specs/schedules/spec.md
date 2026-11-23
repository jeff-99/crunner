## ADDED Requirements

### Requirement: One-time schedule
The library SHALL provide a one-time schedule that triggers a task exactly once at a given time and reports that it has no further runs afterwards.

#### Scenario: Future time
- **WHEN** a task is registered with a one-time schedule for a time 10 seconds in the future
- **THEN** it runs once, 10 seconds later, and never again

#### Scenario: Window of the single run
- **WHEN** a one-time task without backfill runs
- **THEN** its window starts at the registration time and ends at the given time

#### Scenario: Past time within drift
- **WHEN** a task is registered with a one-time schedule for a time 5 seconds in the past and MaxDrift is 30 seconds
- **THEN** it runs immediately, once

#### Scenario: Past time beyond drift
- **WHEN** a task is registered with a one-time schedule for a time one hour in the past and MaxDrift is 30 seconds
- **THEN** the run is skipped as for any late run and the task is not run at all

### Requirement: Exhausted schedules
A schedule SHALL be able to signal that it has no further runs by returning a dedicated, exported sentinel error. The scheduler MUST treat this signal as normal completion, not as a failure.

#### Scenario: Custom schedule ends
- **WHEN** a user-defined schedule returns the no-more-runs sentinel error
- **THEN** no error is logged and the task is not scheduled again

#### Scenario: Other schedule errors unchanged
- **WHEN** a schedule returns any other error
- **THEN** the error is logged and the task is not scheduled again
