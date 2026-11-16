## MODIFIED Requirements

### Requirement: Interval schedule
The library SHALL provide an interval schedule, constructed with `Every(interval)` and returned as a `Schedule`, whose next run is the last run time plus the interval, or the scheduling start plus the interval when the task has not run yet.

#### Scenario: First interval run
- **WHEN** an interval task has not run yet
- **THEN** its next run is the scheduling start plus the interval

#### Scenario: Subsequent interval run
- **WHEN** an interval task has run before
- **THEN** its next run is its last run start time plus the interval

#### Scenario: Constructor returns Schedule
- **WHEN** `Every(10 * time.Second)` is called
- **THEN** the result can be used wherever a `Schedule` is accepted

### Requirement: Cron schedule
The library SHALL provide a cron schedule, constructed with `Cron(spec, options...)` and returned as a `Schedule` or an error, built from a standard 5-field cron expression or, with the seconds option, a 6-field expression with seconds, including `@` descriptors, evaluated in the time zone given by the location option.

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
