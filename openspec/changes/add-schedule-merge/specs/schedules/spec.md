## ADDED Requirements

### Requirement: Merged schedule
The library SHALL provide a schedule that combines several schedules and whose next run is the earliest next run of any of them. Each combined schedule SHALL be evaluated with the same scheduling start and reference time. Several schedules producing the same instant SHALL result in a single run at that instant. An error returned by any combined schedule SHALL be returned by the merged schedule.

#### Scenario: Times that one cron expression cannot express
- **WHEN** a merged schedule of the cron expressions `0 6 * * *` and `30 18 * * *` is evaluated at 12:00
- **THEN** the next run is 18:30 the same day, and the one after that is 06:00 the next day

#### Scenario: Different schedule types
- **WHEN** a merged schedule of an interval schedule and a cron schedule is evaluated
- **THEN** the next run is the earlier of the interval's next run and the cron's next run, and the interval counts from the reference time passed to the merged schedule

#### Scenario: Equal instants
- **WHEN** two combined schedules both have their next run at the same instant
- **THEN** the merged schedule's next run is that instant and only one run results

#### Scenario: Child error
- **WHEN** one of the combined schedules returns an error while computing its next run
- **THEN** the merged schedule returns that error

#### Scenario: Repeated evaluation after drift skips
- **WHEN** the merged schedule is evaluated repeatedly, each time with the previous result as reference time
- **THEN** each result is strictly later than the previous one and visits every occurrence of every combined schedule in order

### Requirement: Merged schedule validation
The merged schedule constructor SHALL return an error when no schedule is given or when any given schedule is nil. When exactly one schedule is given it SHALL return that schedule.

#### Scenario: No schedules
- **WHEN** a merged schedule is created from an empty list
- **THEN** an error is returned

#### Scenario: Nil schedule
- **WHEN** a merged schedule is created and one of the schedules is nil
- **THEN** an error is returned

#### Scenario: Single schedule
- **WHEN** a merged schedule is created from exactly one schedule
- **THEN** that schedule is returned
