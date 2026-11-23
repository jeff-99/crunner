## ADDED Requirements

### Requirement: Injectable clock
The scheduler SHALL obtain the current time and arm run timers exclusively through a clock supplied at construction. When no clock is supplied, it SHALL use the real system clock. Registration times, trigger windows, expected and actual trigger times, and drift checks SHALL all be based on that clock.

#### Scenario: Default clock
- **WHEN** a scheduler is created without a clock option
- **THEN** tasks run according to real wall-clock time, exactly as before

#### Scenario: Custom clock drives scheduling
- **GIVEN** a scheduler created with a user-supplied clock reporting a fixed time T
- **WHEN** an interval task of 1 minute is registered and the scheduler is started
- **THEN** the task's first window starts at T, ends at T plus 1 minute, and the run is armed on a timer created through the supplied clock

#### Scenario: Custom clock decides drift
- **GIVEN** a scheduler with a user-supplied clock and a task with MaxDrift of 30 seconds
- **WHEN** the task's timer fires while the clock reports a time more than 30 seconds after the expected trigger time
- **THEN** the run is skipped as excessive drift, regardless of real elapsed time
