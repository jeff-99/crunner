## MODIFIED Requirements

### Requirement: Task timeout
Each run SHALL receive a context that is cancelled when the scheduler stops or when the task timeout elapses. The task timeout SHALL be the task's own timeout option if set, otherwise the scheduler-wide task timeout, which defaults to 30 days. A timeout of zero or less SHALL mean the run is only cancelled when the scheduler stops.

#### Scenario: Scheduler stop cancels the run
- **WHEN** the scheduler is stopped during a run
- **THEN** the run's context is cancelled

#### Scenario: Default timeout
- **WHEN** a task runs on a scheduler created without a task timeout option
- **THEN** the run's context has a deadline 30 days after the run starts

#### Scenario: Scheduler-wide timeout
- **WHEN** the scheduler is created with a task timeout of 5 seconds and a task runs for longer
- **THEN** the run's context is cancelled 5 seconds after the run starts

#### Scenario: Per-task override
- **WHEN** the scheduler task timeout is 5 seconds and a task is registered with its own timeout of 1 second
- **THEN** that task's run context is cancelled 1 second after the run starts

#### Scenario: Timeout disabled
- **WHEN** the effective task timeout is zero
- **THEN** the run's context has no deadline
