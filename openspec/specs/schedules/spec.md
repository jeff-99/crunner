# schedules Specification

## Purpose

Defines how the next run time of a task is computed, through a pluggable schedule abstraction and the built-in interval and cron schedules.

## Requirements

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

### Requirement: Cron schedule
The library SHALL provide a cron schedule built from a standard 5-field cron expression or, optionally, a 6-field expression with seconds, including `@` descriptors, evaluated in a given time zone. When no time zone is given (nil), the local time zone SHALL be used. Every next-run computation, including the first run of a task, SHALL evaluate the expression in that time zone. A `CRON_TZ=` or `TZ=` prefix in the expression SHALL take precedence over the given time zone.

#### Scenario: Valid expression
- **WHEN** a cron schedule is created from `*/5 * * * *`
- **THEN** the task runs at every minute divisible by five

#### Scenario: Seconds field
- **WHEN** a cron schedule is created with seconds enabled from `* * * * * *`
- **THEN** the task runs every second

#### Scenario: Invalid expression
- **WHEN** a cron schedule is created from an invalid expression
- **THEN** an error is returned

#### Scenario: Nil time zone
- **WHEN** a cron schedule is created with a nil time zone and its next run is computed with and without a last run time
- **THEN** no panic occurs and the next run is computed in the local time zone

#### Scenario: First run honours the time zone
- **WHEN** a cron schedule `0 9 * * *` is created with time zone `America/New_York` and its first next run is computed from 2026-01-01T12:00:00Z
- **THEN** the next run is 2026-01-01T09:00:00 America/New_York (14:00:00Z)

#### Scenario: Later runs honour the time zone
- **WHEN** the same schedule computes its next run after a last run at 2026-01-01T14:00:00Z
- **THEN** the next run is 2026-01-02T09:00:00 America/New_York

#### Scenario: Prefix overrides the time zone
- **WHEN** a cron schedule is created from `CRON_TZ=Asia/Tokyo 0 9 * * *` with time zone `America/New_York`
- **THEN** runs occur at 09:00 Asia/Tokyo
