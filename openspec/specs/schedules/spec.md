# schedules Specification

## Purpose

Defines how the next run time of a task is computed, through a pluggable schedule abstraction and the built-in interval and cron schedules.

## Requirements

### Requirement: Pluggable schedules
The library SHALL accept any schedule that, given a scheduling start time and an optional last run time, returns the next run time or an error.

#### Scenario: Custom schedule
- **WHEN** a task is registered with a user-defined schedule
- **THEN** the scheduler uses that schedule to compute each next run time

#### Scenario: Schedule error
- **WHEN** a schedule returns an error while computing the next run
- **THEN** the error is logged and the task is not scheduled again

### Requirement: Interval schedule
The library SHALL provide an interval schedule whose next run is the last run time plus the interval, or the scheduling start plus the interval when the task has not run yet.

#### Scenario: First interval run
- **WHEN** an interval task has not run yet
- **THEN** its next run is the scheduling start plus the interval

#### Scenario: Subsequent interval run
- **WHEN** an interval task has run before
- **THEN** its next run is its last run start time plus the interval

### Requirement: Cron schedule
The library SHALL provide a cron schedule built from a standard 5-field cron expression or, optionally, a 6-field expression with seconds, including `@` descriptors, evaluated in a given time zone.

#### Scenario: Valid expression
- **WHEN** a cron schedule is created from `*/5 * * * *`
- **THEN** the task runs at every minute divisible by five

#### Scenario: Seconds field
- **WHEN** a cron schedule is created with seconds enabled from `* * * * * *`
- **THEN** the task runs every second

#### Scenario: Invalid expression
- **WHEN** a cron schedule is created from an invalid expression
- **THEN** an error is returned
