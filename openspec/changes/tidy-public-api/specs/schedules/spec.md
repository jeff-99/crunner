## MODIFIED Requirements

### Requirement: Interval schedule
The library SHALL provide an interval schedule, constructed with `Every(interval)` and returned as a `Schedule`, whose next run is the reference time plus the interval, or the scheduling start plus the interval when there is no reference time.

#### Scenario: First interval run
- **WHEN** an interval task has not run yet
- **THEN** its next run is the scheduling start plus the interval

#### Scenario: Subsequent interval run
- **WHEN** an interval task has a previous window
- **THEN** its next run is that window's end plus the interval, regardless of when the previous run started

#### Scenario: Constructor returns Schedule
- **WHEN** `Every(10 * time.Second)` is called
- **THEN** the result can be used wherever a `Schedule` is accepted

### Requirement: Cron schedule
The library SHALL provide a cron schedule, constructed with `Cron(spec, options...)` and returned as a `Schedule` or an error, built from a standard 5-field cron expression or, with the seconds option, a 6-field expression with seconds, including `@` descriptors, evaluated in the time zone given by the location option. When no location option is given, or it is nil, the local time zone SHALL be used. Every next-run computation, including the first run of a task, SHALL evaluate the expression in that time zone. A `CRON_TZ=` or `TZ=` prefix in the expression SHALL take precedence over the given time zone.

#### Scenario: Valid expression
- **WHEN** a cron schedule is created from `*/5 * * * *`
- **THEN** the task runs at every minute divisible by five

#### Scenario: Seconds field
- **WHEN** a cron schedule is created with the seconds option from `* * * * * *`
- **THEN** the task runs every second

#### Scenario: Invalid expression
- **WHEN** a cron schedule is created from an invalid expression
- **THEN** an error is returned

#### Scenario: Six fields without seconds option
- **WHEN** a cron schedule is created from `* * * * * *` without the seconds option
- **THEN** an error is returned

#### Scenario: Nil time zone
- **WHEN** a cron schedule is created without a location option, or with a nil location, and its next run is computed with and without a last run time
- **THEN** no panic occurs and the next run is computed in the local time zone

#### Scenario: First run honours the time zone
- **WHEN** a cron schedule `0 9 * * *` is created with location `America/New_York` and its first next run is computed from 2026-01-01T12:00:00Z
- **THEN** the next run is 2026-01-01T09:00:00 America/New_York (14:00:00Z)

#### Scenario: Later runs honour the time zone
- **WHEN** the same schedule computes its next run after a last run at 2026-01-01T14:00:00Z
- **THEN** the next run is 2026-01-02T09:00:00 America/New_York

#### Scenario: Prefix overrides the time zone
- **WHEN** a cron schedule is created from `CRON_TZ=Asia/Tokyo 0 9 * * *` with location `America/New_York`
- **THEN** runs occur at 09:00 Asia/Tokyo
