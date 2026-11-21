## MODIFIED Requirements

### Requirement: Scheduler construction
The library SHALL provide a constructor that returns an idle scheduler, configurable through functional options for the logger, the task summary interval, the task timeout, the schedule timeout and the internal queue size.

#### Scenario: Default construction
- **WHEN** a scheduler is created without options
- **THEN** it is idle, logs nothing, and uses a task summary interval of 30 seconds, a task timeout of 30 days, a schedule timeout of 1 minute and a queue size of 100

#### Scenario: Queue size clamped
- **WHEN** a scheduler is created with a queue size below 1
- **THEN** the scheduler uses a queue size of 1
