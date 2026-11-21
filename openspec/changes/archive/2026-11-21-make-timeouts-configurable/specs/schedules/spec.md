## ADDED Requirements

### Requirement: Schedule computation timeout
Each call that computes a task's next run time SHALL receive a context that is cancelled when the scheduler stops or after the scheduler's schedule timeout, which defaults to 1 minute. A schedule timeout of zero or less SHALL mean the context is only cancelled when the scheduler stops.

#### Scenario: Default schedule timeout
- **WHEN** a schedule's next-run computation is invoked on a scheduler without a schedule timeout option
- **THEN** its context has a deadline 1 minute in the future

#### Scenario: Custom schedule timeout exceeded
- **WHEN** the schedule timeout is 1 second and a custom schedule blocks until its context is done and then returns the context error
- **THEN** the error is logged and the task is not scheduled again
